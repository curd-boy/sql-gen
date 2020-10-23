package mysql

import (
	"errors"
	"strings"
	"vitess.io/vitess/go/vt/sqlparser"
)
// Column t1.name as user_name
type Column struct {
	Table string // t1
	Name  string // name
	Alias string // user_name
}


func parseComment(cs []string) (*SelectFuncTemp, error) {
	// 至少要指定函数名
	if len(cs) < 1 || !strings.Contains(cs[0], "name:") {
		return nil, errors.New("need function name")
	}
	f := SelectFuncTemp{}
	for i := range cs {
		// -- name: GetUser :one/:many 函数注释 -- 默认many,one需要指定
		// -- params:  -- 由sql语句反推生成到函数中,直接指定为条件扩展sql,暂时不支持指定(TODO)
		// -- result: id,last_name -- sql反推,指定则定义相应结构体GetUserRes,暂时不支持指定(TODO)
		ops := strings.Split(cs[i], " ")
		if strings.HasPrefix(cs[i], "-- name:") {
			if len(ops) < 3 {
				return nil, errors.New("function name comment too less")
			}
			f.Name = ops[2]
			if len(ops) >= 4 {
				if ops[3] == ":one" {
					f.IsOne = true
				} else if ops[3] == ":many" {
					f.IsOne = false
				} else {
					f.Comment = strings.Join(ops[3:], " ")
				}
			}
		}
	}
	return &f, nil
}

func parseTableExpr(u sqlparser.TableExpr) []TableName {
	switch expr := u.(type) {
	case *sqlparser.AliasedTableExpr:
		return []TableName{parseAliasedTableExpr(expr)}
	case *sqlparser.JoinTableExpr:
		return parseJoinTableExpr(expr)
	default:
		return nil
	}
}

func parseAliasedTableExpr(expr *sqlparser.AliasedTableExpr) TableName {
	t := TableName{}
	if !expr.As.IsEmpty() {
		t.Alias = expr.As.String()
	}
	tn, ok := expr.Expr.(sqlparser.TableName)
	if ok {
		if tn.Name.IsEmpty() {
			return t
		}
		t.Table = tn.Name.String()
		t.DB = tn.Qualifier.String()
	}
	return t
}

// from users t1 left join info t2 on t1.id = t2.tid
func parseJoinTableExpr(expr *sqlparser.JoinTableExpr) []TableName {
	t := make([]TableName, 0)
	if ta, ok := expr.LeftExpr.(*sqlparser.AliasedTableExpr); ok {
		t = append(t, parseAliasedTableExpr(ta))
	}
	if ta, ok := expr.RightExpr.(*sqlparser.AliasedTableExpr); ok {
		t = append(t, parseAliasedTableExpr(ta))
	}
	return t
}

func parseAndExpr(expr sqlparser.Expr) []Column {
	cs := make([]Column, 0)
	switch t := expr.(type) {
	case *sqlparser.AndExpr:
		cs = append(cs, parseAndExpr(t.Left)...)
		cs = append(cs, parseAndExpr(t.Right)...)
	case *sqlparser.OrExpr:
		cs = append(cs, parseAndExpr(t.Left)...)
		cs = append(cs, parseAndExpr(t.Right)...)
	case *sqlparser.ComparisonExpr:
		colName := ""
		tableAlias := ""
		l, ok := t.Left.(*sqlparser.ColName)
		if !ok {
			return nil
		}
		colName = l.Name.String()
		tableAlias = l.Qualifier.Name.String()
		r, ok := t.Right.(*sqlparser.SQLVal)
		if !ok {
			return nil
		}
		if r.Type == sqlparser.ValArg {
			cs = append(cs, Column{
				Alias: colName,
				Name:  colName,
				Table: tableAlias,
			})
		}
	default:
		return nil
	}
	return cs
}