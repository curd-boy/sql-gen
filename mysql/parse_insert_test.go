package mysql

import (
	"fmt"
	"testing"
)

func TestParseInsertSql(t *testing.T) {
	sql := `insert into  db.users (name,age,creat_at) values (?,?,?),(?,?,?);`
	fmt.Println( ParseInsertSql(sql))
}
