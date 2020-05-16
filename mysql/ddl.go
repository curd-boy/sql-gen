package mysql

import (
	"vitess.io/vitess/go/vt/sqlparser"
)

func ParseDDL(sql string) *TableTemp {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil
	}
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		return nil
	}
	table := TableTemp{}
	if !ddl.Table.IsEmpty() {
		table.Name = ddl.Table.Name.String()
	}
	cols := make([]ColumnTemp, 0)
	for _, col := range ddl.TableSpec.Columns {
		comment := ""
		if col.Type.Comment != nil {
			comment = string(col.Type.Comment.Val)
		}
		cols = append(cols, ColumnTemp{
			Name:    col.Name.String(),
			Type:    convertColumnType(&col.Type),
			Comment: comment,
		})
	}
	table.Columns = cols
	return &table
}
func convertColumnType(ct *sqlparser.ColumnType) string {
	// 可以直接转换
	if  t,ok :=types[ct.Type];ok {
		return t
	}
	// TODO 特殊类型
	// 枚举

	return types[ct.Type]
}

//
//-- name: Users
//CREATE TABLE users
//(
//id         int auto_increment primary key,
//first_name varchar(255)                                        default ''        not null,
//last_name  varchar(255)                                        default ''        null,
//age        int                                                 default 0         not null,
//job_status enum ('APPLIED', 'PENDING', 'ACCEPTED', 'REJECTED') default 'APPLIED' not null
//)
