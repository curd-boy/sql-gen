package mysql

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// ParseDDLPath 解析ddl语句文件地址
func ParseDDLPath(p string) ([]TableTemp, error) {
	ts := make([]TableTemp, 0)
	sqlTemps, err := ParseSqlPath(p)
	if err != nil {
		return ts, err
	}
	for _, temp := range sqlTemps {
		ddl, err := ParseDDL(temp.Sql)
		if err != nil {
			return ts, err
		}
		ts = append(ts, *ddl)
	}
	return ts, err
}

func ParseSqlPath(p string) ([]SqlTemp, error) {
	sqlTemps := make([]SqlTemp, 0)
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".sql") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sqlTemps = append(sqlTemps, ParseSql(bufio.NewReader(f))...)
		return nil
	})
	return sqlTemps, err
}

// Parse sql文件地址 指定包名
func Parse(p string, pack string) (map[string]Temp, error) {
	if pack == "" {
		pack = "default_package"
	}
	sqlTemps, err := ParseSqlPath(p)
	if err != nil {
		return nil, err
	}
	// 先解析ddl语句 得到表结构
	ts := make([]TableTemp, 0)

	ddlSql, selectSql, updateSql, insertSql, deleteSql := splitSqlType(sqlTemps)

	for i := range ddlSql {
		p, err := ParseDDL(ddlSql[i].Sql)
		if err != nil {
			continue
		}
		ts = append(ts, *p)
	}

	setTableDDL(ts)
	// 每张表一个文件
	temps := make(map[string]Temp, 0)
	for i := range ts {
		temps[ts[i].Name] = Temp{
			Package: pack,
			Table:   ts[i],
		}
	}
	for _, s := range selectSql {
		f, err := parseCommentSelect(s.Comment)
		if err != nil {
			return nil, err
		}
		tables, params, results, err := ParseSelectSql(s.Sql)
		if err != nil {
			return nil, err
		}
		f.Table = tables[0].Table
		f.Params = params
		f.Result = results
		f.Sql = s.Sql
		temp := temps[f.Table]
		temp.SelectFunc = append(temp.SelectFunc, *f)
		temps[f.Table] = temp
	}
	for _, s := range updateSql {
		f, err := parseCommentUpdate(s.Comment)
		if err != nil {
			return nil, err
		}
		tables, params, cond, err := ParseUpdateSql(s.Sql)
		if err != nil {
			return nil, err
		}
		f.Table = tables[0].Table
		f.Params = params
		f.Sql = s.Sql
		f.Condition = cond
		temp := temps[f.Table]
		temp.UpdateFunc = append(temp.UpdateFunc, *f)
		temps[f.Table] = temp
	}
	for _, s := range deleteSql {
		f, err := parseCommentDelete(s.Comment)
		if err != nil {
			return nil, err
		}
		tables, params, err := ParseDeleteSql(s.Sql)
		if err != nil {
			return nil, err
		}
		f.Table = tables[0].Table
		f.Params = params
		f.Sql = s.Sql
		temp := temps[f.Table]
		temp.DeleteFunc = append(temp.DeleteFunc, *f)
		temps[f.Table] = temp
	}
	for _, s := range insertSql {
		f, err := parseCommentInsert(s.Comment)
		if err != nil {
			return nil, err
		}
		tables, params, lens, err := ParseInsertSql(s.Sql)
		if err != nil {
			return nil, err
		}
		f.Table = tables[0].Table
		f.Params = params
		f.Sql = s.Sql
		f.ValuesLen = lens
		temp := temps[f.Table]
		temp.InsertFunc = append(temp.InsertFunc, *f)
		temps[f.Table] = temp
	}

	return temps, nil
}

type SqlTemp struct {
	Comment []string
	Sql     string
}

func ParseSql(r *bufio.Reader) []SqlTemp {
	sqlTemps := make([]SqlTemp, 0)
	sql := bytes.NewBufferString("")
	sqlTemp := SqlTemp{}
	for {
		bs, _, err := r.ReadLine()
		if err != nil {
			break
		}
		s := string(bytes.TrimSpace(bs)) + " "
		if s == " " {
			continue
		}
		if strings.HasPrefix(s, "--") ||
			strings.HasPrefix(s, "/*") ||
			strings.HasSuffix(s, "*/") {
			sqlTemp.Comment = append(sqlTemp.Comment, s)
			continue
		}
		sql.WriteString(s)
		if strings.HasSuffix(s, "; ") {
			sqlTemp.Sql = sql.String()
			sqlTemps = append(sqlTemps, sqlTemp)
			sqlTemp = SqlTemp{}
			sql.Reset()
		}
	}
	return sqlTemps
}

func splitSqlType(sqlTemps []SqlTemp) ([]SqlTemp, []SqlTemp, []SqlTemp, []SqlTemp, []SqlTemp) {
	ddlSql := make([]SqlTemp, 0)
	selectSql := make([]SqlTemp, 0)
	updateSql := make([]SqlTemp, 0)
	insertSql := make([]SqlTemp, 0)
	deleteSql := make([]SqlTemp, 0)
	for i := range sqlTemps {
		if strings.HasPrefix(sqlTemps[i].Sql, "create") || strings.HasPrefix(sqlTemps[i].Sql, "CREATE") {
			ddlSql = append(ddlSql, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "select") ||
			strings.HasPrefix(sqlTemps[i].Sql, "SELECT") ||
			strings.HasPrefix(sqlTemps[i].Sql, "(select") || // union
			strings.HasPrefix(sqlTemps[i].Sql, "(SELECT") {
			selectSql = append(selectSql, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "update") || strings.HasPrefix(sqlTemps[i].Sql, "UPDATE") {
			updateSql = append(updateSql, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "insert") || strings.HasPrefix(sqlTemps[i].Sql, "INSERT") {
			insertSql = append(insertSql, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "delete") || strings.HasPrefix(sqlTemps[i].Sql, "INSERT") {
			deleteSql = append(deleteSql, sqlTemps[i])
		}
	}
	return ddlSql, selectSql, updateSql, insertSql, deleteSql
}
