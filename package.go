package main

import (
	"bytes"
	"cmp"
	"fmt"
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
	Examples         Examples
	ExternalExamples Examples
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

type Examples []Example

func (es Examples) Playable() Examples {
	var out Examples
	for _, e := range es {
		if e.Playable {
			out = append(out, e)
		}
	}
	return out
}

func (es Examples) Named(name string) (Example, error) {
	for _, e := range es {
		if e.Name == name {
			return e, nil
		}
	}
	return Example{}, fmt.Errorf("no example named %q found", name)
}

func (es Examples) Matching(pattern string) (Examples, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var out Examples
	for _, e := range es {
		if re.MatchString(e.Name) {
			out = append(out, e)
		}
	}
	return out, nil
}

func examplesFrom(buf *bytes.Buffer, fset *token.FileSet, in []*doc.Example) Examples {
	var acc Examples
	for _, ex := range in {
		acc = append(acc, renderExample(buf, fset, ex))
	}
	return acc
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
