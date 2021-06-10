package main

import (
	"log"
	"os"
	"testing"

	"github.com/curd-boy/sql-gen/mysql"
)

func init() {
	bs, err := os.ReadFile("../mysql/template/query.tpl")
	if err != nil {
		log.Println(err)
		return
	}
	mysql.TPL = string(bs)
}

func Test_parseFP(t *testing.T) {
	parseFP("/Users/xx/go/src/sql-gen/mysql/sql/ddl.sql")
}
