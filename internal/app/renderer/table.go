package renderer

import (
	"io"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

type Data interface {
	Columns() []string
	Rows() ([][]any, error)
	Caption() string
}

// TableData implements the Data interface.
type TableData struct {
	columns []string
	rows    [][]any
	caption string
}

func NewTableData(columns []string, rows [][]any, caption string) *TableData {
	return &TableData{
		columns: columns,
		rows:    rows,
		caption: caption,
	}
}

func (t *TableData) Columns() []string {
	return t.columns
}

func (t *TableData) Rows() ([][]any, error) {
	return t.rows, nil
}

func (t *TableData) Caption() string {
	return t.caption
}

func Table(data Data, w io.Writer, c *config.Config) error {
	t := tablewriter.NewTable(w, tablewriter.WithRenderer(renderer.NewColorized(GetTableStyle(c))))
	rows, err := data.Rows()
	if err != nil {
		return err
	}

	t.Header(data.Columns())
	if err := t.Bulk(rows); err != nil {
		return err
	}

	if captionText := data.Caption(); captionText != "" {
		captionColor := getCaptionColor(c.Table.Color.Caption)
		caption := tw.Caption{
			Text: color.New(captionColor).Sprint(captionText),
			Spot: tw.SpotBottomLeft,
		}
		t.Caption(caption)
	}
	return t.Render()
}
