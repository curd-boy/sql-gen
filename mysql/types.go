package mysql

// 可以直接转换的类型
var types = map[string]string{
	"tinyint":    "int8",
	"smallint":   "int16",
	"mediumint":  "int",
	"int":        "int",
	"integer":    "int",
	"bigint":     "int64",
	"float":      "float32",
	"double":     "float64",
	"decimal":    "float64",
	"char":       "string",
	"varchar":    "string",
	"tinyblob":   "string",
	"tinytext":   "string",
	"blob":       "string",
	"text":       "string",
	"mediumblob": "string",
	"mediumtext": "string",
	"longblob":   "string",
	"longtext":   "string",
	"time":       "time.Time",
	"date":       "time.Time",
	"year":       "time.Time",
	"timestamp":  "time.Time",
	"datetime":   "time.Time",
}
