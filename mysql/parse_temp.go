package mysql

import (
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/wzshiming/namecase"
	"gopkg.in/ffmt.v1"
)

type Temp struct {
	Package    string
	Table      TableTemp
	SelectFunc []SelectFuncTemp
	UpdateFunc []UpdateFuncTemp
	InsertFunc []InsertFuncTemp
	DeleteFunc []DeleteFuncTemp
}

type TableTemp struct {
	Name    string
	Comment string
	Enums   map[string][]string // {"status":[1,2,3]}
	Columns []ColumnTemp
}

// ColumnTemp {"name","string","姓名"}
type ColumnTemp struct {
	Name    string
	Type    string // go type
	Comment string
}

type SelectFuncTemp struct {
	Name    string
	Table   string // 方法所属的表 联合查询以第一张为准
	IsOne   bool   // 返回单条信息
	Comment string
	Sql     string
	Params  []ColumnTemp
	Result  []ColumnTemp
}

type UpdateFuncTemp struct {
	Name      string
	Table     string
	Comment   string
	Sql       string
	Params    []ColumnTemp // 修改的列
	Condition []ColumnTemp // 条件
}

type InsertFuncTemp struct {
	Name      string
	Table     string
	Comment   string
	Sql       string
	Params    []ColumnTemp
	ValuesLen int
}

type DeleteFuncTemp struct {
	Name    string
	Table   string
	Comment string
	Sql     string
	Params  []ColumnTemp
}

var FuncMaps = map[string]interface{}{
	"SnakeName":    namecase.ToLowerSnake,
	"CamelName":    namecase.ToUpperHump,
	"CamelNameLow": namecase.ToCamel,
	"CompletePlaceholder": func(n int) string {
		if n == 0 {
			return ""
		}
		if n == 1 {
			return "?"
		}
		ns := make([]byte, n*2-1)
		for i := 0; i < 2*n-1; i += 2 {
			ns[i] = '?'
			if i == 2*n-2 {
				break
			}
			ns[i+1] = ','
		}
		return string(ns)
	},
	"TrimSpecial": func(s string) string {
		return strings.TrimSuffix(strings.TrimPrefix(s, "'"), "'")
	},
	"RangeNum": func(n int) []int {
		if n <= 0 {
			return []int{}
		}
		ns := make([]int, n)
		for i := 0; i < n; i++ {
			ns[i] = i
		}
		return ns
	},
}

func ParseTemp(tplPath string, outPath string, temp *Temp) {
	t, err := ioutil.ReadFile(tplPath)
	if err != nil {
		ffmt.Mark(err)
		return
	}
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		ffmt.Mark(err)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			ffmt.Mark(err)
		}
	}()
	tpl := template.New("query")
	tpl.Funcs(FuncMaps)
	tpl, err = tpl.Parse(string(t))
	if err != nil {
		ffmt.Mark(err)
		return
	}

	err = tpl.Execute(f, temp)
	if err != nil {
		ffmt.Mark(err)
	}
}
