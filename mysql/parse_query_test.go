
package mysql

import (
	"testing"
)

func TestParseSelectQuery(t *testing.T) {
	sql := `select t1.age,t2.name from db.users t1 left join info t2 on t1.id =t2.tid `
	// right join db.age t2 on t1.id = t2.tid  `
	//st,_ := sqlparser.Parse(sql)
	//ffmt.P(st)
	ParseSelectQuery(sql)
}