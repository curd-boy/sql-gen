package mysql

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// ParseDDLPath 解析ddl语句文件地址
func ParseDDLPath(p string) []TableTemp {
	ts := make([]TableTemp, 0)
	sqlTemps := ParseSqlPath(p)
	for _, temp := range sqlTemps {
		ts = append(ts, *ParseDDL(temp.Sql))
	}
	return ts
}

func ParseSqlPath(p string) []SqlTemp {
	sqlTemps := make([]SqlTemp, 0)
	filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
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
	return sqlTemps
}

// Parse sql文件地址
func Parse(p string) {
	sqlTemps := ParseSqlPath(p)
	// 先解析ddl语句 得到表结构
	ts := make([]TableTemp, 0)

	ddlSqls, selectSqls, updateSqls, insertSqls, deleteSqls := splitSqlType(sqlTemps)
	for i := range ddlSqls {
		ts = append(ts, *ParseDDL(ddlSqls[i].Sql))
	}
	_, _, _, _ = selectSqls, updateSqls, insertSqls, deleteSqls
	setTableDDL(ts)

	// 每张表一个文件
	funcs := make([]FuncTemp, 0)
	for i := range selectSqls {
		f := parseComment(selectSqls[i].Comment)
		cols := ParseSelectQuery(selectSqls[i].Sql)
		// Param解析 TODO
		f.Result = cols
	}
	_ = funcs
	// 解析sql语句
	ss := ParseSqlPath("")
	_ = ss
	for _, t := range ts {
		// TODO 解析sql
		_ = t
	}
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
		s := string(bytes.TrimSpace(bs))
		if strings.HasPrefix(s, "--") ||
			strings.HasPrefix(s, "/*") ||
			strings.HasSuffix(s, "*/") {
			sqlTemp.Comment = append(sqlTemp.Comment, s)
			continue
		}
		sql.WriteString(s)
		if strings.HasSuffix(s, ";") {
			sql.WriteString(s)
			sqlTemp.Sql = sql.String()
			sqlTemps = append(sqlTemps, sqlTemp)
			sqlTemp = SqlTemp{}
			sql.Reset()
		}
	}
	return sqlTemps
}

func splitSqlType(sqlTemps []SqlTemp) ([]SqlTemp, []SqlTemp, []SqlTemp, []SqlTemp, []SqlTemp) {
	ddlSqls := make([]SqlTemp, 0)
	selectSqls := make([]SqlTemp, 0)
	updateSqls := make([]SqlTemp, 0)
	insertSqls := make([]SqlTemp, 0)
	deleteSqls := make([]SqlTemp, 0)
	for i := range sqlTemps {
		if strings.HasPrefix(sqlTemps[i].Sql, "create") || strings.HasPrefix(sqlTemps[i].Sql, "CREATE") {
			ddlSqls = append(ddlSqls, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "select") || strings.HasPrefix(sqlTemps[i].Sql, "SELECT") {
			selectSqls = append(selectSqls, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "update") || strings.HasPrefix(sqlTemps[i].Sql, "UPDATE") {
			updateSqls = append(updateSqls, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "insert") || strings.HasPrefix(sqlTemps[i].Sql, "INSERT") {
			insertSqls = append(insertSqls, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "delete") || strings.HasPrefix(sqlTemps[i].Sql, "INSERT") {
			deleteSqls = append(deleteSqls, sqlTemps[i])
		}
	}
	return ddlSqls, selectSqls, updateSqls, insertSqls, deleteSqls
}
