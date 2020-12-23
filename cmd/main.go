package main

import (
	"flag"
	"github.com/curd-boy/sql-gen/mysql"
)

var sqlFP string
var sqlF string
var sqlS string
var fileOut string

func main() {
	flag.StringVar(&sqlFP, "fp", "./", "解析指定目录下所有.sql文件中的所有语句")
	flag.StringVar(&sqlF, "f", "./test.sql", "解析指定.sql文件中的所有语句")
	flag.StringVar(&sqlS, "s", "select * from users where id = ?;", "解析指定sql语句")
	flag.StringVar(&fileOut, "o", "./", "输出到指定目录")

	flag.Parse()

}
func parseFP(fp string) {
	sqlTemps, err := mysql.ParseSqlPath(fp)
	if err != nil {
		return
	}
	ts,err:= mysql.Convert(sqlTemps,"packName")
	if err != nil {
		return
	}
	for s, temp := range ts {
		mysql.ParseTemp("./mysql/template/query.tpl", fileOut+s+".go", &temp)
	}
}
func parseF(f string)   {

}
func parseS(s string)   {}
func output(out string) {}
