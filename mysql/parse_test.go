package mysql

import (
	"bufio"
	"fmt"
	"gopkg.in/ffmt.v1"
	"strings"
	"testing"
)

func TestParseDDLPath(t *testing.T) {

	r := bufio.NewReader(strings.NewReader(`1231
456
111;`))
	for {
		bs, is, err := r.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(is, string(bs))
	}
}

func TestParse(t *testing.T) {
	ffmt.P(Parse("./", "mysql"))
}
