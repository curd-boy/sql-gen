package {{.Package}}

import (
	"context"
	"database/sql"
)

// {{CamelName .Table.Name}}DBS {{if .Table.Comment}}{{.Table.Comment}}{{else}}...{{end}}
type {{CamelName .Table.Name}}DBS struct {
	db *sql.Tx
}

// New{{CamelName .Table.Name}}Query ...
func New{{CamelName .Table.Name}}Query(db *sql.Tx) *{{CamelName .Table.Name}}DBS {
	return &{{CamelName .Table.Name}}DBS{
		db: db,
	}
}

{{range $k,$v := .Table.Enums}}
type {{$k}} string
const (
    {{- range $i,$vv := $v}}
    {{printf "%s_%s"  $k $vv | CamelName }} = "{{TrimSpecial $vv}}"
	{{- end}}
)
{{- end}}

// {{CamelName .Table.Name}} {{.Table.Comment}}
type {{CamelName .Table.Name}} struct {
{{- range $i, $c := .Table.Columns }}
{{CamelName $c.Name}}    {{$c.Type}}    `json:"{{$c.Name}}"`    // '{{$c.Comment}}',
{{- end}}
}

{{range $i,$Func := .InsertFunc}}
// {{$Func.Name}}Param {{$Func.Comment}}参数
type {{$Func.Name}}Param struct {
{{- range $i, $c := $Func.Params}}
	{{CamelName $c.Name}}       {{$c.Type}}
{{- end}}
}
// {{$Func.Name}}  {{$Func.Comment}}
func (q *{{$.Table.Name}}DBS) {{$Func.Name}}(ctx context.Context ,{{if eq $Func.ValuesLen 0 }}val *{{$Func.Name}}Param{{else}} val []*{{$Func.Name}}Param {{end}})(int64, error) {
    var sql{{$Func.Name}} = `{{$Func.Sql}}`
    {{if eq  $Func.ValuesLen 0}}
		var args = []interface{}{
		{{- range $i, $c := $Func.Params}}
    		val.{{CamelName $c.Name}},
    	{{- end}}
		}
    {{else}}
		var args = make([]interface{},0,{{len $Func.Params}}*len(val))
    	for i := range val{
			args =append(args,{{- range $i, $c := $Func.Params}}
			val[i].{{CamelName $c.Name}},
		{{- end}})
    	}
    {{end}}
    res, err :=  q.db.ExecContext(ctx, sql{{$Func.Name}},args...)

    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}
{{else}}
// {{CamelName .Table.Name}}Insert delete
func (q *{{.Table.Name}}DBS){{CamelName .Table.Name}}Insert(ctx context.Context, val *{{CamelName .Table.Name}})(int64,error){
    var sqlInsert = "insert into {{.Table.Name}} (
    {{- range $i, $c := .Table.Columns }}
	    {{- if eq $i 0 }}`{{SnakeName $c.Name}}`
	    {{- else}},`{{SnakeName $c.Name}}`
	    {{- end}}
    {{- end}}) values({{len .Table.Columns | CompletePlaceholder}});"
    res, err :=  q.db.ExecContext(ctx, sqlInsert,
    {{- range $i, $c := .Table.Columns }}
	    val.{{-  CamelName $c.Name}},
    {{- end}})
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}
{{end}}
{{range $i,$Func := .SelectFunc}}
type {{$Func.Name}}Result struct {
	{{- range $i, $c := $Func.Result}}
	{{CamelName $c.Name}}       {{$c.Type}}
	{{- end}}
}
{{if $Func.IsOne}}
// {{$Func.Name}}  {{$Func.Comment}}
func (q *{{$.Table.Name}}DBS) {{$Func.Name}}(ctx context.Context ,
    {{- range $i, $c := $Func.Params}}
    	{{CamelNameLow $c.Name}} {{$c.Type}},
    {{- end}}
) ({{$Func.Name}}Result, error) {
    var sql{{$Func.Name}} = `{{$Func.Sql}}`
    var item {{$Func.Name}}Result
    err := q.db.QueryRowContext(ctx, sql{{$Func.Name}} ,
    {{- range $i, $c := $Func.Params}}
    	{{CamelNameLow $c.Name}},
    {{- end}}
    ).Scan(
    {{- range $i, $c := $Func.Result}}
    	&item.{{CamelName $c.Name}},
    {{- end}}
    )
    if err != nil {
        return item, err
    }
    return item, nil
}
{{else}}
// {{$Func.Name}}  {{$Func.Comment}}
func (q *{{$.Table.Name}}DBS) {{$Func.Name}}(ctx context.Context ,
	{{- range $i, $c := $Func.Params}}
    	{{CamelNameLow $c.Name}}       {{$c.Type}},
    {{- end}}
	) ([]{{$Func.Name}}Result, error) {
    var sql{{$Func.Name}} = `{{$Func.Sql}}`
	rows, err := q.db.QueryContext(ctx, sql{{$Func.Name}} ,
	 {{- range $i, $c := $Func.Params}}
     	{{CamelNameLow $c.Name}},
     {{- end}}
)
	if err != nil {
		return nil, err
	}
	var items []{{$Func.Name}}Result
	for rows.Next() {
		var i {{$Func.Name}}Result
		if err := rows.Scan(
		{{- range $i, $c := $Func.Result}}
			&i.{{CamelName $c.Name}},
		{{- end}}
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
{{end}}
{{end}}