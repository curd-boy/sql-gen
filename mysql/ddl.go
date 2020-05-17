package mysql

import (
	"strings"

	"gopkg.in/ffmt.v1"
	"vitess.io/vitess/go/vt/sqlparser"
)

func ParseDDL(sql string) *TableTemp {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		ffmt.Mark(err)
		return nil
	}
	ffmt.P(stmt)
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		return nil
	}
	table := TableTemp{
		Enums: make(map[string][]string),
	}
	if !ddl.Table.IsEmpty() {
		table.Name = ddl.Table.Name.String()
	}
	cols := make([]ColumnTemp, 0)
	for i := range ddl.TableSpec.Columns {
		cols = append(cols, convertColumnType(table, ddl.TableSpec.Columns[i]))
	}
	ops := strings.Split(ddl.TableSpec.Options, " ")
	for i := range ops {
		if strings.HasPrefix(ops[i], "COMMENT") {
			table.Comment = strings.Split(ops[i], "=")[1]
			table.Comment = table.Comment[1 : len(table.Comment)-1]
		}
	}
	table.Columns = cols
	return &table
}
func convertColumnType(t TableTemp, c *sqlparser.ColumnDefinition) ColumnTemp {
	col := ColumnTemp{
		Name: c.Name.String(),
	}

	if c.Type.Comment != nil {
		col.Comment = string(c.Type.Comment.Val)
	}
	// 可以直接转换
	if t, ok := types[c.Type.Type]; ok {
		col.Type = t
		return col
	}
	// 特殊类型
	// 枚举
	switch c.Type.Type {
	case "enum":
		t.Enums[t.Name+c.Name.String()] = c.Type.EnumValues
	default:
		return col
	}
	return col
}
