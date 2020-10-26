package {{.Package}}

import (
"context"
"database/sql"
)

type {{.Table.Name}}Query struct {
db *sql.Tx
}

func NewUsersQuery(db *sql.Tx) *{{.Table.Name}}Query {
return &{{.Table.Name}}Query{
db: db,
}
}

{{range $k,$v := .Table.Enums}}
    type {{CamelName $.Table.Name}}{{CamelName $k}} string
    const (
    {{range $i,$vv := $v}}
        {{- CamelName $.Table.Name}}{{CamelName $k}}{{CamelName $vv}} = "{{$vv}}"
    {{- end}}
    )
{{end}}

// {{.Table.Name}} {{.Table.Comment}}
type {{.Table.Name}} struct {
{{range $i, $c := .Table.Columns }}
    {{CamelName $c.Name}}    {{$c.Type}}    `json:"{{$c.Name}}"`    // '{{$c.Comment}}',
{{end}}
}

{{range $i,$Func := .SelectFuncs}}
    type {{$Func.Name}}Result struct {
    {{range $i, $c := $Func.Result}}
        {{CamelName $c.Name}}       {{$c.Type}}
    {{end}}
    }

    const sql{{$Func.Name}} = `{{$Func.Sql}}`

    // {{$Func.Name}}  {{$Func.Comment}}
    func (q *{{.Table.Name}}Query) {{$Func.Name}}(ctx context.Context,
    {{range $i, $c := $Func.Params}}
        {{CamelNameLow $c.Name}}       {{$c.Type}},
    {{end}}
    ) ([]{{$Func.Name}}Result, error) {
    rows, err := q.db.QueryContext(ctx, sql{{$Func.Name}} ,
    {{range $i, $c := $Func.Params}}
        {{CamelNameLow $c.Name}},
    {{end}}
    )
    if err != nil {
    return nil, err
    }
    defer rows.Close()
    var items []{{$Func.Name}}Result
    for rows.Next() {
    var i {{$Func.Name}}Result
    if err := rows.Scan(
    {{range $i, $c := $Func.Result}}
        &i.{{CamelName $c.Name}},
    {{end}}
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