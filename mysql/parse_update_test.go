package mysql

import (
	"fmt"
	"testing"
)

func TestParseUpdateSql(t *testing.T) {
	sql := "-- 一个注释 \n" +
		"update users t1 left join info t2 on t1.id = t2.user_id set t1.`name`= '111' , -- 名字 \n " +
		" t2.`age` = ? where t1.`id` = ? and t2.age > 2  and t1.name = ?;"
	// sql = `update users set name = ? , age = ? ,del= 1 where id = ? and del =2 `
	fmt.Println(ParseUpdateSql(sql))
}
