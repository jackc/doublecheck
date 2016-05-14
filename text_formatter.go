package doublecheck

import (
	"io"
	"text/template"
)

type TextFormatter struct {
	w io.Writer
}

func NewTextFormatter(w io.Writer) *TextFormatter {
	return &TextFormatter{w: w}
}

func (tf *TextFormatter) Format(result *CheckResult) error {
	var t = template.Must(template.New("result").Parse(
		`Database:   {{.Database}}
Schema:     {{.Schema}}
User:       {{.User}}
Start Time: {{.StartTime}}
Duration:   {{.Duration}}
{{range $vr := .ViewResults}}
---
Name:       {{$vr.Name}}
Start Time: {{$vr.StartTime}}
Duration:   {{$vr.Duration}}
Error Rows: {{if $vr.Rows}}
{{range $i, $row := $vr.Rows}}  | {{range $key, $value := $row}}{{$key}}: {{$value}} | {{end}}
{{end}}{{else}}None
{{end}}{{end}}
`))

	return t.Execute(tf.w, result)
}
