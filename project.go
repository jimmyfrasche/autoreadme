package main

import (
	"context"
	"fmt"
	"go/doc"
	"go/token"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func Packages(ctx context.Context, project string) (*token.FileSet, []*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Dir:     project,
		Tests:   true,
		Mode:    packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Fset:    token.NewFileSet(),
	}
	ps, err := packages.Load(cfg, "./...")
	return cfg.Fset, ps, err
}

type info struct {
	pkg, test *packages.Package
	doc       *doc.Package
	examples  []*doc.Example
	dir       string
	data      any
	oldReadme []byte
	template  *template.Template
}

func PairPackagesWithXTests(in []*packages.Package) map[string]*info {
	m := map[string]*info{}
	for _, pkg := range in {
		test := strings.HasSuffix(pkg.Name, "_test")
		p, ok := m[pkg.PkgPath]
		if !ok {
			p = &info{}
		}
		if !test {
			p.pkg = pkg
		} else {
			p.test = pkg
		}
		m[pkg.PkgPath] = p
	}
	return m
}

func CollectProjectErrors(in map[string]*info) []error {
	var acc []error
	extend := func(errs []packages.Error) {
		for _, err := range errs {
			acc = append(acc, err)
		}
	}
	for k, p := range in {
		if p.pkg == nil {
			acc = append(acc, fmt.Errorf("%s only has external test files, cannot process", k))
		} else {
			extend(p.pkg.Errors)
		}
		if p.test != nil {
			extend(p.test.Errors)
		}
	}
	return acc
}

func ProcessPackageDir(fset *token.FileSet, p *info, defaultTemplate *template.Template) error {
	var err error
	p.doc, err = doc.NewFromFiles(fset, p.pkg.Syntax, p.pkg.PkgPath, doc.PreserveAST)
	if err != nil {
		return err
	}

	if p.test != nil {
		p.examples = doc.Examples(p.test.Syntax...)
	}
	// One day there may be a p.pkg.Dir but alas not today
	p.dir = filepath.FromSlash(path.Dir(p.pkg.GoFiles[0]))

	p.data, err = DirData(p.dir)
	if err != nil {
		return err
	}

	templateSrc, err := DirTemplate(p.dir)
	if err != nil {
		return err
	}

	p.template, err = parseTemplateOr(templateSrc, defaultTemplate)
	if err != nil {
		return err
	}

	p.oldReadme, err = DirReadme(p.dir)
	return err
}
