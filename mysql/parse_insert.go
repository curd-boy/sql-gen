package mysql

import (
	"errors"
	"fmt"
	"vitess.io/vitess/go/vt/sqlparser"
)

type InsertSql struct {
}

var _defaultInsert = InsertSql{}

func ParseInsertSql(sql string) ([]TableName, []ColumnTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil, nil
	}
	switch t := stmt.(type) {
	case *sqlparser.Insert:
		ts, args, err := _defaultInsert.parseInsert(t)
		return ts, convertColsToTemps(ts, args), err
	default:
		return nil, nil, errors.New(fmt.Sprintf("unknown type %v", stmt))
	}
}

func (s *InsertSql) parseInsert(i *sqlparser.Insert) ([]TableName, []Column, error) {
	ts := []TableName{{DB: i.Table.Qualifier.String(), Table: i.Table.Name.String()}}
	cols := make([]Column, 0, len(i.Columns))
	for _, col := range i.Columns {
		cols = append(cols, Column{
			Table: "",
			Name:  col.String(),
			Alias: "",
		})
	}
	return ts, cols, nil
}
