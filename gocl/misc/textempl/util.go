package textempl

import (
	"bytes"
	"fmt"
	"io/fs"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/huandu/xstrings"
)

func IncludeFnMap(t *template.Template) map[string]any {
	return map[string]any{
		"include": func(name string, data any) (string, error) {
			var buf bytes.Buffer

			if err := t.ExecuteTemplate(&buf, name, data); err != nil {
				return "", fmt.Errorf("failed to call 'include' function; err: %w", err)
			}

			return buf.String(), nil
		},
	}
}

func TextFnMap() map[string]any {
	res := sprig.TxtFuncMap()
	if _, ok := res["pascalcase"]; !ok {
		res["pascalcase"] = xstrings.ToPascalCase
	}

	return res
}

// ParseFSWithInclude parses template definitions from specified file system.
// In addition, resulting template enriched with text functions map [TextFnMap] and 'include' function [IncludeFnMap] that
// allows to invoke template execution by its name.
func ParseFSWithInclude(fsys fs.FS, patterns ...string) (*template.Template, error) {
	t := template.New("")
	t.Funcs(TextFnMap()).Funcs(IncludeFnMap(t))

	return t.ParseFS(fsys, patterns...)
}

func Parse(text string) (*template.Template, error) {
	t := template.New("")
	t.Funcs(TextFnMap()).Funcs(IncludeFnMap(t))

	return t.Parse(text)
}
