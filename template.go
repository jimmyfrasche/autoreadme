package main

import (
	_ "embed"
	"text/template"
)

func parseTemplate(src string) (*template.Template, error) {
	return template.New("").Parse(src)
}

func parseTemplateOr(src []byte, defaultTemplate *template.Template) (*template.Template, error) {
	if len(src) == 0 {
		return defaultTemplate, nil
	}
	return parseTemplate(string(src))
}

//go:embed default.template
var defaultTemplateSrc string
