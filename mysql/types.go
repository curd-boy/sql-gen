package mysql

// 可以直接转换的类型
var types = map[string]string{
	"varchar":   "string",
	"int":       "int",
	"tinyint":   "int8",
	"text":      "string",
	"timestamp": "time.Time",
}
