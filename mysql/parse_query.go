package mysql

import (
	"errors"
	"log"
	"vitess.io/vitess/go/vt/sqlparser"
)

type SelectSql struct {

}
func ParseSelectSql(sql string) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil, nil, err
	}
	tables := make([]TableName, 0)
	params := make([]ColumnTemp, 0)
	results := make([]ColumnTemp, 0)
	switch expr := stmt.(type) {
	case *sqlparser.Union:
		tables, params, results, err = parseUnion(expr)
	case *sqlparser.Select:
		tables, params, results, err = parseSelect(expr)
	default:
		err = errors.New("unknown select sql type")
	}
	return tables, params, results, nil
}

func parseUnion(u *sqlparser.Union) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	l, ok := u.Left.(*sqlparser.Select)
	if !ok {
		return nil, nil, nil, errors.New("left of sql is not union select")
	}
	return parseSelect(l)
	// union all 对于字段解析结果没有影响 只需要解析一部分即可
}

// 解析Select
func parseSelect(s *sqlparser.Select) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	// 优先解析from 获取表结构,以解析星号
	// [{db users t1},{db2 info t2}]
	ts := parseFrom(s)
	if len(ts) == 0 {
		return nil, nil, nil, errors.New("no table found")
	}

	cols := parseSelectColumn(ts, s)
	if len(cols) == 0 {
		return nil, nil, nil, errors.New("has no columns")
	}

	return ts, convertColsToTemps(ts, parseWhere(s)), convertColsToTemps(ts, cols), nil
}

// mysql函数操作要设置别名
func parseSelectColumn(selectTables []TableName, s *sqlparser.Select) []Column {
	cols := make([]Column, 0)
	for i := range s.SelectExprs {
		switch t := s.SelectExprs[i].(type) {
		case *sqlparser.AliasedExpr:
			cols = append(cols, parseSelectColumnNonStar(t)...)
		case *sqlparser.StarExpr:
			cols = append(cols, parseSelectColumnStar(selectTables, t)...)
		default:
			log.Println("unknown column type")
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
func parseSelectColumnStar(selectTables []TableName, star *sqlparser.StarExpr) []Column {
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
	tableNames := make([]TableName, 0)
	for _, expr := range s.From {
		tableNames = append(tableNames, parseTableExpr(expr)...)
	}
	return tableNames
}

func parseWhere(s *sqlparser.Select) []Column {
	if s.Where == nil {
		return nil
	}
	return parseAndExpr(s.Where.Expr)
}
