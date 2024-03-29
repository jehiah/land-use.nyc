package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// func commaInt(i int) string {
// 	return humanize.Comma(int64(i))
// }

var nonASCII = regexp.MustCompile(`[^a-z0-9]+`)

func cssClass(s string) string {
	return nonASCII.ReplaceAllString(strings.ToLower(s), "-")
}

func newTemplate(fs fs.FS, n string, funcs ...template.FuncMap) *template.Template {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		// "Comma":      commaInt,
		// "Time":       humanize.Time,
		"RFC3339": func(t time.Time) string { return t.Format(time.RFC3339) },
		// "CSSClass":   cssClass,
		// "Slugify":    slug.Make,
		"TrimPrefix": strings.TrimPrefix,
	}
	if len(funcs) > 0 {
		for _, f := range funcs {
			for k, v := range f {
				funcMap[k] = v
			}
		}
	}
	t := template.New("empty").Funcs(funcMap)
	return template.Must(t.ParseFS(fs, filepath.Join("templates", n), "templates/base.html"))
}
