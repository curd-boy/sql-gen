package mysql

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/ffmt.v1"
)

func TestParseDDLPath(t *testing.T) {
	sql := bytes.NewBufferString("")
	r := bufio.NewReader(strings.NewReader(`123
145
6111;`))
	for {
		bs, is, err := r.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		sql.Write(bs)
		fmt.Println(is, string(bs))
	}
	fmt.Println(sql.String())
}

func TestParse(t *testing.T) {
	Parse(strings.NewReader("select name,age from users where id = ? "), "test")
}

func TestConvert(t *testing.T) {
	reader, err := ParseSqlPath("./")
	if err != nil {
		t.Log(err)
		return
	}
	sqlTemps := GetSqlTemp(bufio.NewReader(reader))
	ffmt.P(Convert(sqlTemps, "mysql"))
}
