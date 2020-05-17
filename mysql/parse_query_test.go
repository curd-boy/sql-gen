package mysql

import (
	"testing"

	"gopkg.in/ffmt.v1"
	"vitess.io/vitess/go/vt/sqlparser"
)

func TestParseSelectQuery(t *testing.T) {
	sql := `select age, name from db.users t1 left join info t2 on t1.id =t2.tid `
	// right join db.age t2 on t1.id = t2.tid  `
	st, _ := sqlparser.Parse(sql)
	ffmt.P(st)
	//ParseSelectQuery(sql)
}
