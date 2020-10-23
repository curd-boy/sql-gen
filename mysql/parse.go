package mysql

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ffmt.v1"
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
func Parse(p string, pack string) error {
	if pack == "" {
		pack = "default_package"
	}
	sqlTemps, err := ParseSqlPath(p)
	if err != nil {
		return err
	}
	// 先解析ddl语句 得到表结构
	ts := make([]TableTemp, 0)

	ddlSqls, selectSqls, updateSqls, insertSqls, deleteSqls := splitSqlType(sqlTemps)

	for i := range ddlSqls {
		p, err := ParseDDL(ddlSqls[i].Sql)
		if err != nil {
			continue
		}
		ts = append(ts, *p)
	}
	_, _, _, _ = selectSqls, updateSqls, insertSqls, deleteSqls
	setTableDDL(ts)

	// 每张表一个文件
	// funcs := make([]SelectFuncTemp, 0)
	funcMaps := make(map[string][]SelectFuncTemp)
	for i := range selectSqls {
		f, err := parseComment(selectSqls[i].Comment)
		if err != nil {
			return err
		}
		tables, params, results, err := ParseSelectQuery(selectSqls[i].Sql)
		if err != nil {
			return err
		}
		f.Table = tables[0].Table
		f.Params = params
		f.Result = results
		f.Sql = selectSqls[i].Sql
		funcMaps[f.Table] = append(funcMaps[f.Table], *f)
	}
	// for i := range funcs {
	//	funcMaps[funcs[i].Table] = append(funcMaps[funcs[i].Table], funcs[i])
	// }
	// 组合成模板列表
	temps := make([]Temp, 0)
	for i := range ts {
		temps = append(temps, Temp{
			Package:     pack,
			Table:       ts[i],
			SelectFuncs: funcMaps[ts[i].Name],
		})
	}
	ffmt.P(temps)
	return nil
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
	ddlSqls := make([]SqlTemp, 0)
	selectSqls := make([]SqlTemp, 0)
	updateSqls := make([]SqlTemp, 0)
	insertSqls := make([]SqlTemp, 0)
	deleteSqls := make([]SqlTemp, 0)
	for i := range sqlTemps {
		if strings.HasPrefix(sqlTemps[i].Sql, "create") || strings.HasPrefix(sqlTemps[i].Sql, "CREATE") {
			ddlSqls = append(ddlSqls, sqlTemps[i])
		}
		if strings.HasPrefix(sqlTemps[i].Sql, "select") ||
			strings.HasPrefix(sqlTemps[i].Sql, "SELECT") ||
			strings.HasPrefix(sqlTemps[i].Sql, "(select") || // union
			strings.HasPrefix(sqlTemps[i].Sql, "(SELECT") {
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
