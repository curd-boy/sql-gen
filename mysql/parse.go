package mysql

func Parse(p string) {
	// 先解析ddl语句 得到表结构
	ParseDDL(p)
	ts := []TableTemp{{
		Name: "users",
		Columns: []ColumnTemp{
			{
				Name: "id",
				Type: "int",
			}, {
				Name: "age",
				Type: "int",
			}, {
				Name: "name",
				Type: "string",
			},
		},
	}}
	initTableDDL(ts)
}
