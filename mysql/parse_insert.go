package mysql

import (
	"errors"
	"fmt"
	"vitess.io/vitess/go/vt/sqlparser"
)

type InsertSql struct {
}

var _defaultInsert = InsertSql{}

func ParseInsertSql(sql string) ([]TableName, []ColumnTemp,int, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil,0, nil
	}
	switch t := stmt.(type) {
	case *sqlparser.Insert:
		ts, args,lens, err := _defaultInsert.parseInsert(t)
		return ts, convertColsToTemps(ts, args), lens,err
	default:
		return nil, nil, 0,errors.New(fmt.Sprintf("unknown type %v", stmt))
	}
}

func (s *InsertSql) parseInsert(i *sqlparser.Insert) ([]TableName, []Column, int, error) {
	ts := []TableName{{DB: i.Table.Qualifier.String(), Table: i.Table.Name.String()}}
	cols := make([]Column, 0, len(i.Columns))
	for _, col := range i.Columns {
		cols = append(cols, Column{
			Table: "",
			Name:  col.String(),
			Alias: "",
		})
	}
	lens := 0
	if vs, ok := i.Rows.(sqlparser.Values); ok {
		lens = len(vs)
	}
	return ts, cols, lens, nil
}
