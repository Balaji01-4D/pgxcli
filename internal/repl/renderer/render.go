package render

import "github.com/jedib0t/go-pretty/v6/table"

func RenderTables(tables []table.Writer, style table.Style) string {
	var str string
	for _, table := range tables {
		table.SetStyle(style)
		str += table.Render() + "\n"
	}
	return str
}
