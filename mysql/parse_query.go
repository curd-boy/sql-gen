package mysql

import (
	"gopkg.in/ffmt.v1"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	// {db users t1}  {db2 info t2}
	selectTables []TableName
	// TableDDL {"users":{"id","int"}}
	TableDDL map[string]map[string]string
)

func initTableDDL(ts []TableTemp) {
	for i := range ts {
		for i2 := range ts[i].Columns {
			TableDDL[ts[i].Name][ts[i].Columns[i2].Name] = ts[i].Columns[i2].Type
		}
	}
}
func ParseSelectQuery(sql string) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return
	}
	ffmt.P(stmt)
	switch expr := stmt.(type) {
	case *sqlparser.Union:
		ParseUnion(expr)
	case *sqlparser.Select:
		ParseSelect(expr)
	default:

	}
}

type TableName struct {
	DB    string
	Table string
	Alias string
}

func ParseUnion(u *sqlparser.Union) {
	l, ok := u.Left.(*sqlparser.Select)
	if !ok {
		return
	}
	ParseSelect(l)
	// unnion all 对于字段解析结果没有影响 只需要解析一部分即可
	// r, ok := u.Right.(*sqlparser.Select)
	// if !ok {
	// 	return
	// }
	// ParseSelect(r)
}

// 解析Select
func ParseSelect(s *sqlparser.Select) {
	// 优先解析from 获取表结构,以解析星号
	// {db users t1}  {db2 info t2}
	selectTables = parseFrom(s)
	defer func() {
		// 解析完成后清空
		selectTables = nil
	}()
	cols := parseSelectColumn(s)
	_ = cols
}

// t1.name as user_name
type Column struct {
	Table string // t1
	Name  string // name
	Alias string // user_name
}

// 函数操作要设置别名
func parseSelectColumn(s *sqlparser.Select) []Column {
	cols := make([]Column, 0)
	for i := range s.SelectExprs {
		n, ok := s.SelectExprs[i].(*sqlparser.AliasedExpr)
		if ok {
			cols = append(cols, parseSelectColumnNonStar(n)...)
		}
		star, ok := s.SelectExprs[i].(*sqlparser.StarExpr)
		if ok {
			cols = append(cols, parseSelectColumnStar(star)...)
		}
	}
	return cols
}

// 解析普通字段
func parseSelectColumnNonStar(n *sqlparser.AliasedExpr) []Column {
	cols := make([]Column, 0)
	col := Column{}
	if !n.As.IsEmpty() {
		col.Alias = n.As.String()
	}
	c, ok := n.Expr.(*sqlparser.ColName)
	if !ok {
		return nil
	}
	col.Name = c.Name.String()
	col.Table = c.Qualifier.Name.String()
	cols = append(cols, col)
	return cols
}

// 解析select星号字段
func parseSelectColumnStar(star *sqlparser.StarExpr) []Column {
	cols := make([]Column, 0)
	if star.TableName.IsEmpty() { // 只有星号,从查询的所有表结构中读取
		for i := range selectTables {
			cols = append(cols, parseAllColumnInTable(selectTables[i])...)
		}
		return cols
	}
	ta := star.TableName.Name.String()
	for i := range selectTables {
		if ta == selectTables[i].Alias {
			cols = append(cols, parseAllColumnInTable(selectTables[i])...)
		}
	}
	return cols
}

func parseAllColumnInTable(t TableName) []Column {
	cols := make([]Column, 0)
	// {"id","int"}
	for key := range TableDDL[t.Table] {
		cols = append(cols, Column{
			Name:  key,
			Alias: key,
		})
	}
	return cols
}

// 解析Select下的From
func parseFrom(s *sqlparser.Select) []TableName {
	tableNames := make([]TableName, len(s.From))
	for i := range s.From {
		switch s.From[i].(type) {
		case *sqlparser.AliasedTableExpr:
			tableNames = append(tableNames, parseFromAllJoin(s.From[i])...)
		case *sqlparser.JoinTableExpr:
			tableNames = append(tableNames, parseFromLeftJoin(s.From[i])...)
		default:
			return nil
		}
	}
	return tableNames
}
func parseAliasedTableExpr(expr *sqlparser.AliasedTableExpr) TableName {
	t := TableName{}
	if !expr.As.IsEmpty() {
		t.Alias = expr.As.String()
	}
	if node, ok := expr.Expr.(sqlparser.SQLNode); ok {
		if tn, ok := node.(*sqlparser.TableName); ok {
			if tn.Name.IsEmpty() {
				return t
			}
			t.Table = tn.Name.String()
			t.DB = tn.Qualifier.String()
		}
	}
	return t
}

// from user t1 , info t2
func parseFromAllJoin(expr sqlparser.TableExpr) []TableName {
	t := make([]TableName, 0)
	if ta, ok := expr.(*sqlparser.AliasedTableExpr); ok {
		t = append(t, parseAliasedTableExpr(ta))
	}
	return t
}

// from users t1 left join info t2 on t1.id = t2.tid
func parseFromLeftJoin(expr sqlparser.TableExpr) []TableName {
	t := make([]TableName, 0)
	node, ok := expr.(sqlparser.SQLNode)
	if !ok {
		return nil
	}
	exp, ok := node.(*sqlparser.JoinTableExpr)
	if !ok {
		return nil
	}
	if ta, ok := exp.LeftExpr.(*sqlparser.AliasedTableExpr); ok {
		t = append(t, parseAliasedTableExpr(ta))
	}
	if ta, ok := exp.RightExpr.(*sqlparser.AliasedTableExpr); ok {
		t = append(t, parseAliasedTableExpr(ta))
	}
	return t
}
