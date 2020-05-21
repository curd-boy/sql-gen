package mysql

import (
	"strings"

	"gopkg.in/ffmt.v1"
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
func ParseSelectQuery(sql string) ([]TableName, []ColumnTemp, []ColumnTemp) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		ffmt.Mark(err, sql)
		return nil, nil, nil
	}
	tables := make([]TableName, 0)
	params := make([]ColumnTemp, 0)
	results := make([]ColumnTemp, 0)
	switch expr := stmt.(type) {
	case *sqlparser.Union:
		tables, params, results = ParseUnion(expr)
	case *sqlparser.Select:
		tables, params, results = ParseSelect(expr)
	default:

	}
	return tables, params, results
}

type TableName struct {
	DB    string
	Table string
	Alias string
}

func ParseUnion(u *sqlparser.Union) ([]TableName, []ColumnTemp, []ColumnTemp) {
	l, ok := u.Left.(*sqlparser.Select)
	if !ok {
		return nil, nil, nil
	}
	return ParseSelect(l)
	// unnion all 对于字段解析结果没有影响 只需要解析一部分即可
	// r, ok := u.Right.(*sqlparser.Select)
	// if !ok {
	// 	return
	// }
	// return ParseSelect(r)
}

// 解析Select
func ParseSelect(s *sqlparser.Select) ([]TableName, []ColumnTemp, []ColumnTemp) {
	// 优先解析from 获取表结构,以解析星号
	// [{db users t1},{db2 info t2}]
	selectTables := parseFrom(s)
	// {"t1":"users","t2":"info"}
	tables := make(map[string]string)
	for i := range selectTables {
		tables[selectTables[i].Alias] = selectTables[i].Table
	}
	cols := parseSelectColumn(selectTables, s)
	result := make([]ColumnTemp, len(cols))
	for i, col := range cols {
		// 多表查询 需要写表别名 否则无法定位字段归属
		// 无别名当作单表处理
		tableName := selectTables[0].Table
		if col.Table != "" {
			tableName = tables[col.Table]
		}
		result[i] = ColumnTemp{
			Name:    col.Alias,
			Type:    TableDDL[tableName][col.Name].Type,
			Comment: TableDDL[tableName][col.Name].Comment,
		}
	}
	// 解析where
	wheres := parseWhere(s)
	params := make([]ColumnTemp, len(wheres))
	for i := range wheres {
		// 多表查询 需要写表别名 否则无法定位字段归属
		// 无别名当作单表处理
		tableName := selectTables[0].Table
		if wheres[i].Table != "" {
			tableName = tables[wheres[i].Table]
		}
		params[i] = ColumnTemp{
			Name:    wheres[i].Alias,
			Type:    TableDDL[tableName][wheres[i].Name].Type,
			Comment: TableDDL[tableName][wheres[i].Name].Comment,
		}
	}
	return selectTables, params, result
}

// t1.name as user_name
type Column struct {
	Table string // t1
	Name  string // name
	Alias string // user_name
}

// mysql函数操作要设置别名
func parseSelectColumn(selectTables []TableName, s *sqlparser.Select) []Column {
	cols := make([]Column, 0)
	for i := range s.SelectExprs {
		n, ok := s.SelectExprs[i].(*sqlparser.AliasedExpr)
		if ok {
			cols = append(cols, parseSelectColumnNonStar(n)...)
		}
		star, ok := s.SelectExprs[i].(*sqlparser.StarExpr)
		if ok {
			cols = append(cols, parseSelectColumnStar(selectTables, star)...)
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

func parseWhere(s *sqlparser.Select) []Column {
	if s.Where == nil {
		return nil
	}
	return parseAndExpr(s.Where.Expr)
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

func parseComment(cs []string) *SelectFuncTemp {
	// 至少要指定函数名
	if len(cs) < 1 || !strings.Contains(cs[0], "name:") {
		return nil
	}
	f := SelectFuncTemp{}
	for i := range cs {
		//-- name: GetUser :one/:many 函数注释 -- 默认many,one需要指定
		//-- params:  -- 由sql语句反推生成到函数中,直接指定为条件扩展sql,暂时不支持指定(TODO)
		//-- result: id,last_name -- sql反推,指定则定义相应结构体GetUserRes,暂时不支持指定(TODO)
		ops := strings.Split(cs[i], " ")
		if strings.HasPrefix(cs[i], "-- name:") {
			if len(ops) < 3 {
				return nil
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
	return &f
}
