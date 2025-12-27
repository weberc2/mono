package offline

import (
	"strings"
	"text/template"
)

func Prompt(
	t *template.Template,
	name string,
	sources []string,
) (prompt string, err error) {
	var (
		w strings.Builder
		fields = struct {
			Name string
			Sources []string
		} {
			Name: name,
			Sources: sources,
		}
	)
	prompt = w.String()
	err = t.Execute(&w, &fields)
	return
}