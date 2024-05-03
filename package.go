package main

import (
	"bytes"
	"go/doc"
	"go/token"
	"path/filepath"
	"regexp"
)

type Package struct {
	Name             string
	Import           string
	Documentation    *Doc
	Data             any
	Library          bool
	Command          bool
	Notes            map[string][]Note
	Examples         map[string]Example
	ExternalExamples map[string]Example
}

type Note struct {
	UID  string
	Body string
}

type Example struct {
	Code     string
	Output   string
	Playable bool
}

func examplesFrom(buf *bytes.Buffer, fset *token.FileSet, in []*doc.Example) map[string]Example {
	if len(in) == 0 {
		return nil
	}
	m := make(map[string]Example, len(in))
	for _, ex := range in {
		m[ex.Name] = renderExample(buf, fset, ex)
	}
	return m
}

var versionRE = regexp.MustCompile("^v[0-9]+$")

func PackageFromInfo(fset *token.FileSet, p *info) *Package {
	Name := p.pkg.Name
	Command := Name == "main"

	// Compute the name that go build would use for the binary
	if Command {
		dir, file := filepath.Split(p.dir)
		if versionRE.MatchString(file) {
			file = filepath.Base(dir)
		}
		Name = file
	}

	Doc := &Doc{
		Synopsis: Synopsis(p.doc.Doc),
		Doc:      ParseDoc(p.doc.Doc),
	}

	Notes := make(map[string][]Note, len(p.doc.Notes))
	for k, ns := range p.doc.Notes {
		acc := make([]Note, 0, len(ns))
		for _, n := range ns {
			acc = append(acc, Note{
				UID:  n.UID,
				Body: n.Body,
			})
		}
		Notes[k] = acc
	}

	var buf bytes.Buffer
	Examples := examplesFrom(&buf, fset, p.examples)
	ExternalExamples := examplesFrom(&buf, fset, p.externalExamples)

	return &Package{
		Name:             Name,
		Import:           p.pkg.PkgPath,
		Documentation:    Doc,
		Data:             p.data,
		Library:          !Command,
		Command:          Command,
		Notes:            Notes,
		Examples:         Examples,
		ExternalExamples: ExternalExamples,
	}
}
