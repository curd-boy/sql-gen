package mysql

import (
	"fmt"
	"testing"
)

func TestParseDeleteSql(t *testing.T) {
	sql := `delete from users   where id = ? and name = ? and age = ?`
	fmt.Println(ParseDeleteSql(sql))
}
