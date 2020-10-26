package mysql

import (
	"errors"
	"strings"

	"github.com/wzshiming/namecase"

	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	// TableDDL {"users":{"id","int","comment"}}
	TableDDL map[string]TableColumn
)

type TableColumn map[string]ColumnTemp

func init() {
	TableDDL = make(map[string]TableColumn)
}

func setTableDDL(ts []TableTemp) {
	for i := range ts {
		for i2 := range ts[i].Columns {
			t, ok := TableDDL[ts[i].Name]
			if !ok {
				t = make(TableColumn)
			}
			t[ts[i].Columns[i2].Name] = ts[i].Columns[i2]
			TableDDL[ts[i].Name] = t
		}
	}
}

func ParseDDL(sql string) (*TableTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		return nil, errors.New("invalid sql")
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
	return &table, err
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
		col.Type = namecase.ToUpperHump(t.Name + "_" + c.Name.String())
	default:
		return col
	}
	return col
}
