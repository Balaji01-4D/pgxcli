package database

import (
	"context"
	"fmt"
	"io"
	"os"
	"pgcli/internals/logger"
	"pgcli/internals/repl"
	"strings"
	"time"

	osUser "os/user"

	"github.com/balaji01-4d/pgxspecial"
	"github.com/balaji01-4d/pgxspecial/database"
	_ "github.com/balaji01-4d/pgxspecial/dbcommands"
	"github.com/jackc/pgx/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	Exit pgxspecial.SpecialResultKind = 100 + iota
)

type Postgres struct {
	CurrentBD           string
	Executor            *Executor
	ForcePasswordPrompt bool
	NeverPasswordPrompt bool
	ctx                 context.Context
}

type PostgresExit struct{}

func (e PostgresExit) ResultKind() pgxspecial.SpecialResultKind {
	return Exit
}

func New(neverPasswordPrompt, forcePasswordPrompt bool, ctx context.Context) *Postgres {
	pgxspecial.RegisterCommand(pgxspecial.SpecialCommandRegistry{
		Cmd:         "\\q",
		Syntax:      "\\q",
		Description: "Quit Pgxcli",
		Handler: func(_ context.Context, _ database.Queryer, _ string, _ bool) (pgxspecial.SpecialCommandResult, error) {
			return &PostgresExit{}, nil
		},
		CaseSensitive: true,
	})
	return &Postgres{
		NeverPasswordPrompt: neverPasswordPrompt,
		ForcePasswordPrompt: forcePasswordPrompt,
		ctx:                 ctx,
	}
}

func (p *Postgres) Connect(host, user, password, database, dsn string, port uint16) error {

	if user == "" {
		currentUser, err := osUser.Current()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		user = currentUser.Username
	}

	if database == "" {
		database = user
	}

	if p.NeverPasswordPrompt && password == "" {
		password = os.Getenv("PGPASSWORD")
	}

	if p.ForcePasswordPrompt && password == "" {
		fmt.Print("Password: ")
		var pwd string
		fmt.Scanln(&pwd)
		password = strings.TrimSpace(pwd)
	}

	if dsn != "" {
		parsedDsn, err := pgx.ParseConfig(dsn)
		if err != nil {
			return fmt.Errorf("failed to parse DSN: %w", err)
		}

		host = parsedDsn.Host
		port = parsedDsn.Port
	}

	exec, err := NewExecutor(host, database, user, password, port, dsn, p.ctx)
	if err != nil {
		return err
	}
	p.Executor = exec
	p.CurrentBD = database
	logger.Log.Info("Database connection established", "database", database, "user", user)

	return nil

}

func (p *Postgres) ConnectDSN(dsn string) error {
	return p.Connect("", "", "", "", dsn, 0)
}

func (p *Postgres) ConnectURI(uri string) error {
	parsedURI, err := pgx.ParseConfig(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}
	return p.Connect(parsedURI.Host, parsedURI.User, parsedURI.Password, parsedURI.Database, "", parsedURI.Port)
}

func (p *Postgres) Close() {
	if p.Executor != nil {
		p.Executor.Close()
	}
}

func (p *Postgres) IsConnected() bool {
	return p.Executor != nil && p.Executor.IsConnected()
}

func (p *Postgres) GetConnectionInfo() {
	logger.Log.Debug("Connection information",
		"connection string", p.Executor.Pool.Config().ConnString(),
		"host", p.Executor.Host,
		"Port", p.Executor.Port,
		"Database", p.Executor.Database,
		"User", p.Executor.User,
		"URI", p.Executor.URI,
	)
}

func (p *Postgres) ChangeDatabase(dbName string) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	exec, err := NewExecutor(
		p.Executor.Host,
		dbName,
		p.Executor.User,
		p.Executor.Password,
		p.Executor.Port,
		"",
		p.ctx,
	)

	if err != nil {
		return fmt.Errorf("failed to create new executor: %w. Keeping the current database connection to %s", err, p.CurrentBD)
	}
	p.Executor = exec
	p.CurrentBD = dbName
	logger.Log.Info("Database changed", "database", dbName)

	return nil
}

func (p *Postgres) RunCli() error {
	defer p.Executor.Close()
	if !p.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}
	repl := repl.NewModel(p.CurrentBD)
	defer repl.Close()

	for {
		query := repl.Read()

		// check for empty string
		if strings.TrimSpace(query) == "" {
			continue
		}

		start := time.Now()

		if p.IsChangeDBCommand(query) {
			parts := strings.Fields(query)
			if len(parts) < 2 {
				repl.PrintError(fmt.Errorf("database name not provided"))
				continue
			}
			newDB := parts[1]
			err := p.ChangeDatabase(newDB)
			if err != nil {
				repl.PrintError(err)
			}
			continue
		}

		metaResult, okay, err := pgxspecial.ExecuteSpecialCommand(p.ctx, p.Executor.Pool, query)
		if err != nil {
			repl.PrintError(err)
			continue
		}
		if okay {
			// check for exit command
			if metaResult.ResultKind() == Exit {
				break
			}

			splCommandResults, err := HandleSpecialCommmand(metaResult)
			if err != nil {
				repl.PrintError(err)
				repl.PrintTime(time.Since(start))
				continue
			}

			execTime := time.Since(start)
			if len(splCommandResults) > 0 {
				var resultStr string
				for _, result := range splCommandResults {
					resultStr += result.Render() + "\n"
				}
				repl.Print(resultStr)
			}
			repl.PrintTime(execTime)
			continue
		}

		result, err := p.Executor.Execute(p.ctx, query)
		if err != nil {
			logger.Log.Error("Query execution failed", "error", err)
			repl.PrintError(err)
			continue
		}
		// TODO: Print using REPL
		HandleResult(result)
	}

	return nil
}

func HandleResult(result Result) {
	switch res := result.(type) {
	case *QueryResult:
		tw, err := res.Render()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(tw.Render())
	case *ExecResult:
		fmt.Println(res.Render())
	}
}

func (p *Postgres) IsChangeDBCommand(sql string) bool {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return false
	}
	first := strings.ToLower(strings.Fields(sql)[0])

	switch first {
	case "use", "\\c", "\\connect":
		return true
	}
	return false
}

func HandleSpecialCommmand(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
	switch result.ResultKind() {

	case pgxspecial.ResultKindRows:
		return HandleRowsResult(result)

	case pgxspecial.ResultKindDescribeTable:
		return handleDescribeTableResult(result)

	case pgxspecial.ResultKindExtensionVerbose:
		return handleExtensionVerboseResult(result)

	default:
		return nil, fmt.Errorf("unknown special command result kind")
	}
}

func HandleRowsResult(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
	resultRows, ok := result.(pgxspecial.RowResult)
	if !ok {
		return nil, fmt.Errorf("invalid row result type")
	}
	return []table.Writer{
		RenderRows(resultRows.Rows),
	}, nil
}

func RenderRows(pgxRows pgx.Rows) table.Writer {
	tw := table.NewWriter()

	columns := make(table.Row, len(pgxRows.FieldDescriptions()))
	for i, col := range pgxRows.FieldDescriptions() {
		columns[i] = setColumnCellColor(col.Name)
	}
	tw.AppendHeader(columns)

	for pgxRows.Next() {
		values, err := pgxRows.Values()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}
		row := make(table.Row, len(values))
		copy(row, values)
		tw.AppendRow(row)
	}

	return tw
}

func handleDescribeTableResult(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
	describeTableResult, ok := result.(pgxspecial.DescribeTableListResult)
	if !ok {
		return nil, fmt.Errorf("invalid describe table result type")
	}

	writers := make([]table.Writer, 0, len(describeTableResult.Results))

	for _, tableDesc := range describeTableResult.Results {
		writers = append(writers, RenderTableDescription(tableDesc))
	}
	return writers, nil
}

func RenderTableDescription(result pgxspecial.DescribeTableResult) table.Writer {
	tw := table.NewWriter()

	columns := make(table.Row, len(result.Columns))
	for i, col := range result.Columns {
		columns[i] = setColumnCellColor(col)
	}
	tw.AppendHeader(columns)
	okay := tw.ImportGrid(result.Data)
	if !okay {
		return nil
	}
	tw.SetCaption(renderTableFooter(result.TableMetaData))
	return tw
}

func renderTableFooter(meta pgxspecial.TableFooterMeta) string {
	var sb strings.Builder

	writeList := func(title string, v []string) {
		if len(v) == 0 {
			return
		}
		sb.WriteString(title)
		sb.WriteByte('\n')
		for _, s := range v {
			sb.WriteString("    ")
			sb.WriteString(s)
			sb.WriteByte('\n')
		}
	}

	writeValue := func(title string, v *string) {
		if v == nil || *v == "" {
			return
		}
		sb.WriteString(title)
		sb.WriteString(*v)
		sb.WriteByte('\n')
	}

	writeBool := func(title string, v *bool) {
		if v == nil {
			return
		}
		sb.WriteString(title)
		if *v {
			sb.WriteString("yes\n")
		} else {
			sb.WriteString("no\n")
		}
	}

	writeList("Indexes:", meta.Indexes)
	writeList("Check constraints:", meta.CheckConstraints)
	writeList("Foreign-key constraints:", meta.ForeignKeys)
	writeList("Referenced by:", meta.ReferencedBy)
	writeValue("View definition:\n", meta.ViewDefinition)

	writeList("Rules:", meta.RulesEnabled)
	writeList("Disabled rules:", meta.RulesDisabled)
	writeList("Rules firing always:", meta.RulesAlways)
	writeList("Rules firing on replica only:", meta.RulesReplica)

	writeList("Triggers:", meta.TriggersEnabled)
	writeList("Disabled triggers:", meta.TriggersDisabled)
	writeList("Triggers firing always:", meta.TriggersAlways)
	writeList("Triggers firing on replica only:", meta.TriggersReplica)

	writeList("Partition of:", meta.PartitionOf)
	writeList("Partition constraint:", meta.PartitionConstraints)
	writeValue("Partition key: ", meta.PartitionKey)
	writeList("Partitions:", meta.Partitions)
	writeValue("", meta.PartitionsSummary)

	writeList("Inherits:", meta.Inherits)
	writeList("Child tables:", meta.ChildTables)
	writeValue("", meta.ChildTablesSummary)
	writeValue("Typed table of type: ", meta.TypedTableOf)
	writeBool("Has OIDs: ", meta.HasOIDs)
	writeValue("Options: ", meta.Options)
	writeValue("Server: ", meta.Server)
	writeValue("FDW Options: ", meta.FDWOptions)
	writeValue("Owned by: ", meta.OwnedBy)

	return sb.String()
}

func handleExtensionVerboseResult(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
	extResult, ok := result.(pgxspecial.ExtensionVerboseListResult)
	if !ok {
		return nil, fmt.Errorf("invalid extension verbose result type")
	}
	writers := make([]table.Writer, 0, len(extResult.Results))

	for _, ext := range extResult.Results {
		writers = append(writers, renderExtensionVerbose(ext))
	}
	return writers, nil
}

func renderExtensionVerbose(ext pgxspecial.ExtensionVerboseResult) table.Writer {
	tw := table.NewWriter()
	tw.SetTitle(ext.Name)

	columns := table.Row{setColumnCellColor("Object Description")}
	tw.AppendHeader(columns)

	for _, objDesc := range ext.Description {
		row := table.Row{objDesc}
		tw.AppendRow(row)
	}
	return tw
}

func setColumnCellColor(s string) string {
	return text.FgCyan.Sprint(s)
}
