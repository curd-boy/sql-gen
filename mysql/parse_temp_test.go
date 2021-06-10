package mysql

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"text/template"
	"time"

	"gopkg.in/ffmt.v1"
)

func Test_ParseTemp(t *testing.T) {
	tpl := template.New("query")
	tpl.Funcs(FuncMaps)
	tpl, err := tpl.Parse(`
		{{range $k,$v := .Table.Enums}}
type {{CamelName $.Table.Name}}{{CamelName $k}} string
const (
    {{- range $i,$vv := $v}}
    {{CamelName $.Table.Name}}{{CamelName $k}}{{CamelName $vv}} = "{{$vv}}"
	{{- end}}
)
{{end}}`)
	if err != nil {
		ffmt.Mark(err)
		return
	}
	tables := TableTemp{
		Name: "users",
		Enums: map[string][]string{
			"gender": {"F", "M"},
			"status": {"on", "off"},
		},
	}
	err = tpl.Execute(os.Stdout, Temp{
		Table: tables,
	})
	if err != nil {
		ffmt.Mark(err)
	}
}

func Test_fmtTemp(t *testing.T) {
	n := time.Now()
	cols := []ColumnTemp{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "age", Type: "int"},
	}
	tables := TableTemp{
		Name:    "users",
		Columns: cols,
	}
	f := []SelectFuncTemp{{
		Name:   "GetUser",
		Sql:    `select * from users;`,
		Params: cols[:1],
		Result: cols[:2],
	}}
	_, _, _ = cols, tables, f
	bf, err := ParseSqlPath("./")
	if err != nil {
		t.Log(err)
		return
	}
	sqlTemps := GetSqlTemp(bufio.NewReader(bf))
	ts, err := Convert(sqlTemps, "mysql")
	if err != nil {
		t.Log(err)
		return
	}
	users := ts["users"]
	tp, err := os.ReadFile("./template/query.tpl")
	if err != nil {
		t.Log(err)
		return
	}
	g, err := os.OpenFile("./sqlGen.go", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		t.Log(err)
		return
	}
	defer g.Close()
	ParseTemp(bytes.NewReader(tp), g, &users)
	ffmt.Mark(time.Since(n))
}
