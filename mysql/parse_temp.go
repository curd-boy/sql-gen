package mysql

import (
	"github.com/wzshiming/namecase"
	"gopkg.in/ffmt.v1"
	"io/ioutil"
	"os"
	"text/template"
)

type Temp struct {
	Package string
	Table   TableTemp
	Funcs   []FuncTemp
}

type TableTemp struct {
	Name    string
	Comment string
	Columns []ColumnTemp
}

// ColumnTemp {"name","string","姓名"}
type ColumnTemp struct {
	Name    string
	Type    string // go type
	Comment string
}

type FuncTemp struct {
	Name    string
	Comment string
	Sql     string
	Params  []ColumnTemp
	Result  []ColumnTemp
}

var FuncMaps = map[string]interface{}{
	"SnakeName":    namecase.ToLowerSnake,
	"CamelName":    namecase.ToUpperHump,
	"CamelNameLow": namecase.ToCamel,
}

func fmtTemp(tplPath string, outPath string, tables *TableTemp, fs []FuncTemp) {
	t, err := ioutil.ReadFile(tplPath)
	if err != nil {
		ffmt.Mark(err)
		return
	}
	f, err := os.OpenFile(outPath, os.O_CREATE, os.ModePerm)
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

	err = tpl.Execute(f, Temp{
		Package: "mysql",
		Table:   *tables,
		Funcs:   fs,
	})
	if err != nil {
		ffmt.Mark(err)
	}
}
