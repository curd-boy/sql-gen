package mysql

import (
	"testing"
)

func TestParseInsertSql(t *testing.T) {
	sql := `insert into  db.users (name,age,creat_at) values (?,?,?),(?,?,?);`
	ParseInsertSql(sql)
}
