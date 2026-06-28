package web

import (
	"html/template"
	"io"
	"log"
	"strings"
)

type Templates struct {
	funcMap template.FuncMap
	tmpl    *template.Template
}

func NewTemplates() *Templates {
	t := &Templates{
		funcMap: template.FuncMap{
			"formatBytes": func(b int64) string {
				if b >= 1048576 { return formatF(float64(b)/1048576) + " MB" }
				if b >= 1024 { return formatF(float64(b)/1024) + " KB" }
				return formatF(float64(b)) + " B"
			},
			"add": func(a, b int) int { return a + b },
			"sub": func(a, b int) int { return a - b },
			"seq": func(n int) []int {
				s := make([]int, n)
				for i := range s { s[i] = i + 1 }
				return s
			},
			"lower": strings.ToLower,
		},
	}

	// Parse ALL templates together so {{template "sidebar" .}} works
	t.tmpl = template.New("").Funcs(t.funcMap)
	for name, str := range templates {
		template.Must(t.tmpl.New(name).Parse(str))
	}

	return t
}

func formatF(f float64) string {
	n := int(f); d := int((f - float64(n)) * 10)
	if d > 0 { return itoa(n) + "." + itoa(d) }
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 { s = string(rune('0'+n%10)) + s; n /= 10 }
	return s
}

func (t *Templates) Execute(w io.Writer, name string, data interface{}) {
	if err := t.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("[WEB] Error executing template %s: %v", name, err)
	}
}
