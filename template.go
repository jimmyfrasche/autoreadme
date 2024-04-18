//Copyright 2013 James Frasche. All rights reserved.
//Use of this code is governed by a BSD-License found in the LICENSE file.

package main

import (
	_ "embed"
	"text/template"
)

func parseTemplate(src string) (*template.Template, error) {
	return template.New("").Parse(src)
}

//go:embed default.template
var defaultTemplateSrc string
