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
	filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sqlTemps := ParseSql(bufio.NewReader(f))
		for _, temp := range sqlTemps {
			ts = append(ts, *ParseDDL(temp.Sql))
		}
		return nil
	})
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

func Parse(p string) {
	// 先解析ddl语句 得到表结构
	ts := ParseDDLPath("")
	setTableDDL(ts)
	// 解析sql语句
	ss := ParseSqlPath("")
	for _, s := range ss {
		// TODO 解析sql
		_ = s
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
