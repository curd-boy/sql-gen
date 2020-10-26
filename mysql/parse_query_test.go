package mysql

import (
	"fmt"
	"testing"
)

func TestParseSelectQuery(t *testing.T) {
	sql := `select t1.age, t2.name 
from db.users t1 left join info t2 on t1.id =t2.tid 
where t1.id = 1 or t2.name = ? and t2.name != ? `

	// right join db.age t2 on t1.id = t2.tid  `
	// st, _ := sqlparser.Parse(sql)
	// ffmt.P(st)
	fmt.Println(ParseSelectSql(sql))
}
