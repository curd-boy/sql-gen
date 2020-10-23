package mysql

import (
	"errors"
	"log"
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
func ParseSelectQuery(sql string) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, nil, nil, err
	}
	tables := make([]TableName, 0)
	params := make([]ColumnTemp, 0)
	results := make([]ColumnTemp, 0)
	switch expr := stmt.(type) {
	case *sqlparser.Union:
		tables, params, results, err = ParseUnion(expr)
	case *sqlparser.Select:
		tables, params, results, err = ParseSelect(expr)
	default:
		err = errors.New("unknown select sql type")
	}
	return tables, params, results, nil
}

type TableName struct {
	DB    string
	Table string
	Alias string
}

func ParseUnion(u *sqlparser.Union) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	l, ok := u.Left.(*sqlparser.Select)
	if !ok {
		return nil, nil, nil, errors.New("left of sql is not union select")
	}
	return ParseSelect(l)
	// union all 对于字段解析结果没有影响 只需要解析一部分即可
}

// 解析Select
func ParseSelect(s *sqlparser.Select) ([]TableName, []ColumnTemp, []ColumnTemp, error) {
	// 优先解析from 获取表结构,以解析星号
	// [{db users t1},{db2 info t2}]
	selectTables  := parseFrom(s)
	if len(selectTables) == 0 {
		return nil, nil, nil, errors.New("no table found")
	}
	// {"t1":"users","t2":"info"}
	tables := make(map[string]string)
	for i := range selectTables {
		tables[selectTables[i].Alias] = selectTables[i].Table
	}
	cols := parseSelectColumn(selectTables, s)
	if len(cols) == 0 {
		return nil, nil, nil, errors.New("has no columns")
	}
	result := make([]ColumnTemp, len(cols))
	for i, col := range cols {
		// 多表查询 需要写表别名 否则无法定位字段归属
		// 无别名当作单表处理
		tableName := selectTables[0].Table
		if col.Table != "" {
			tableName = tables[col.Table]
		}
		if col.Alias == "" {
			col.Alias = col.Name
		}
		result[i] = convertColumnToTemp(tableName, col)
	}
	// 解析where
	wheres := parseWhere(s)
	params := make([]ColumnTemp, len(wheres))
	for i, col := range wheres {
		// 多表查询 需要写表别名 否则无法定位字段归属
		// 无别名当作单表处理
		tableName := selectTables[0].Table
		if wheres[i].Table != "" {
			tableName = tables[wheres[i].Table]
		}
		if col.Alias == "" {
			col.Alias = col.Name
		}
		params[i] = convertColumnToTemp(tableName, col)
	}
	return selectTables, params, result, nil
}


func convertColumnToTemp(tableName string, c Column) ColumnTemp {
	return ColumnTemp{
		Name:    c.Alias,
		Type:    TableDDL[tableName][c.Name].Type,
		Comment: TableDDL[tableName][c.Name].Comment,
	}
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

