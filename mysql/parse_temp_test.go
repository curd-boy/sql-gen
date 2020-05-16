package mysql

import (
	"gopkg.in/ffmt.v1"
	"testing"
	"time"
)

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
	f := FuncTemp{
		Name:   "GetUser",
		Sql:    `select * from users;`,
		Params: cols[:1],
		Result: cols[:2],
	}
	fmtTemp("../mysql/query-template.tpl", "./sql-gen.go", &tables, &f)
	ffmt.Mark(time.Since(n))
}
