package mysql

import (
	"vitess.io/vitess/go/vt/sqlparser"
)

type DeleteSQL struct {
}

var _defaultDelete = DeleteSQL{}

// ParseDeleteSql ...
func ParseDeleteSql(sql string) ([]TableName, []ColumnTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil, err
	}
	ts := make([]TableName, 0)
	conditions := make([]Column, 0)

	switch t := stmt.(type) {
	case *sqlparser.Delete:
		for _, expr := range t.TableExprs {
			ts = append(ts, _defaultDelete.parseTableExpr(expr))
		}
		conditions = append(conditions, parseAndExpr(t.Where.Expr)...)
		return ts, convertColsToTemps(ts, conditions), nil
	default:
		return nil, nil, nil
	}
}

func (s *DeleteSQL) parseTableExpr(expr sqlparser.TableExpr) TableName {
	if te, ok := expr.(*sqlparser.AliasedTableExpr); ok {
		return parseAliasedTableExpr(te)
	}
	return TableName{}
}
