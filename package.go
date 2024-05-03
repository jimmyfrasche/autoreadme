package main

import (
	"bytes"
	"cmp"
	"go/doc"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type Package struct {
	Name             string
	Import           string
	Documentation    *Doc
	Data             any
	Library          bool
	Command          bool
	Notes            Notes
	Examples         map[string]Example
	ExternalExamples map[string]Example
}

type Note struct {
	p    *token.Position
	Kind string
	UID  string
	Body string
}

type Notes []Note

func (ns Notes) Kind(k string) Notes {
	var out Notes
	for _, n := range ns {
		if strings.EqualFold(k, n.Kind) {
			out = append(out, n)
		}
	}
	return out
}

func (ns Notes) UID(uid string) Notes {
	var out Notes
	for _, n := range ns {
		if strings.EqualFold(uid, n.UID) {
			out = append(out, n)
		}
	}
	return out
}

type Example struct {
	Name        string
	Doc         *Doc
	Code        string
	Output      string
	Playable    bool
	Unordered   bool
	EmptyOutput bool
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

	Doc := NewDoc(p.doc.Doc)

	var Notes Notes
	for k, ns := range p.doc.Notes {
		for _, n := range ns {
			p := fset.Position(n.Pos)
			Notes = append(Notes, Note{
				p:    &p,
				Kind: k,
				UID:  n.UID,
				Body: n.Body,
			})
		}
	}
	slices.SortFunc(Notes, func(a, b Note) int {
		return cmp.Or(
			strings.Compare(a.Kind, b.Kind),
			strings.Compare(a.UID, b.UID),
			strings.Compare(a.p.Filename, b.p.Filename),
			cmp.Compare(a.p.Line, b.p.Line),
		)
	})

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
