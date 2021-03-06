package mysql

import (
	"errors"
	"fmt"
	"vitess.io/vitess/go/vt/sqlparser"
)

type UpdateSQL struct {
}

var _defaultUpdate = UpdateSQL{}

func ParseUpdateSql(sql string) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil, nil, err
	}
	switch t := stmt.(type) {
	case *sqlparser.Update:
		ts, args, cond, err := _defaultUpdate.parseUpdate(t)
		return ts, convertColsToTemps(ts, args), convertColsToTemps(ts, cond), err
	default:
		return nil, nil, nil, errors.New(fmt.Sprintf("unknown type %v", stmt))
	}
}

func (s *UpdateSQL) parseUpdate(u *sqlparser.Update) ([]TableName, []Column, []Column, error) {
	ts := make([]TableName, 0)
	args := make([]Column, 0)
	conditions := make([]Column, 0)

	for _, expr := range u.Exprs {
		args = append(args, *s.parseUpdateExpr(expr))
	}
	conditions = append(conditions, parseAndExpr(u.Where.Expr)...)
	for _, expr := range u.TableExprs {
		ts = append(ts, parseTableExpr(expr)...)
	}
	return ts, args, conditions, nil
}

func (s *UpdateSQL) parseUpdateExpr(u *sqlparser.UpdateExpr) *Column {
	return s.parseColName(u.Name)
	// v := u.Expr.(*sqlparser.SQLVal)
	// set age = 11  type = 0; val = 11;
	// fmt.Println( v.Type, string(v.Val))
}

func (s *UpdateSQL) parseColName(c *sqlparser.ColName) *Column {
	return &Column{
		Table: c.Qualifier.Name.String(),
		Name:  c.Name.String(),
		Alias: "",
	}
}
