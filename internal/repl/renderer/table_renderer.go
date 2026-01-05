package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/balaji01-4d/pgxspecial"
	"github.com/jackc/pgx/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func RenderRowsResult(result pgxspecial.SpecialCommandResult) (table.Writer, error) {
	resultRows, ok := result.(pgxspecial.RowResult)
	if !ok {
		return nil, fmt.Errorf("invalid row result type")
	}

	return renderRows(resultRows.Rows), nil
}

func RenderDescribeTableResult(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
	describeTableResult, ok := result.(pgxspecial.DescribeTableListResult)
	if !ok {
		return nil, fmt.Errorf("invalid describe table result type")
	}

	writers := make([]table.Writer, 0, len(describeTableResult.Results))

	for _, tableDesc := range describeTableResult.Results {
		writers = append(writers, renderTableDescription(tableDesc))
	}
	return writers, nil
}

func RenderExtensionVerboseResult(result pgxspecial.SpecialCommandResult) ([]table.Writer, error) {
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

func renderTableDescription(result pgxspecial.DescribeTableResult) table.Writer {
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

func renderRows(pgxRows pgx.Rows) table.Writer {
	defer pgxRows.Close()

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

func setColumnCellColor(s string) string {
	return text.FgCyan.Sprint(s)
}
