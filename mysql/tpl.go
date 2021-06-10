package mysql

import (
	_ "embed"
)

//go:embed template/query.tpl
var TPL string // 默认模板
