package main

import (
	"bytes"
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

var PrintTemplate = flag.Bool("print-template", false, "write the built in template to stdout and exit")
var Version = flag.Bool("version", false, "output version information")

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Printf("Usage of %s:", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *PrintTemplate {
		fmt.Println(defaultTemplateSrc)
		return
	}
	if *Version {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("no version information in build")
			return
		}
		m := info.Main
		sum := ""
		if m.Sum != "" {
			sum = fmt.Sprintf(" (%s)", m.Sum)
		}
		fmt.Printf("%s%s, built with %s\n", cmp.Or(m.Version, "unknown version"), sum, info.GoVersion)
		return
	}

	ctx := context.Background()

	err := Main(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func Main(ctx context.Context) error {
	defaultTemplate, err := parseTemplate(defaultTemplateSrc)
	if err != nil {
		return err
	}

	repo, project, err := Roots()
	if err != nil {
		return err
	}

	// get any global config from the repo

	repoTemplateSrc, err := RepoTemplate(repo)
	if err != nil {
		return err
	}

	// if there's a repo-level template, override the default template
	defaultTemplate, err = parseTemplateOr(repoTemplateSrc, defaultTemplate)
	if err != nil {
		return err
	}

	repoData, err := RepoData(repo)
	if err != nil {
		return err
	}

	ignore, err := RepoIgnore(repo)
	if err != nil {
		return err
	}

	mod, err := ModInfo(repo)
	if err != nil {
		return err
	}

	// get all the packages in this module

	fset, packages, err := Packages(ctx, project)
	if err != nil {
		return err
	}

	// associate X and X_test packages
	pairs := PairPackagesWithXTests(packages)
	// filter out any ignored packages
	for k := range ignore {
		delete(pairs, k)
	}

	// halt on any errors, in packages we plan to inspect
	errs := CollectProjectErrors(pairs)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		return fmt.Errorf("project contains errors, cannot proceed")
	}

	// grab directory info and compute local info from what's been gathered so far
	for imp, info := range pairs {
		if err := ProcessPackageDir(fset, info, defaultTemplate); err != nil {
			return fmt.Errorf("preparing %s: %w", imp, err)
		}
	}

	type Repository struct {
		Data any
	}
	repository := &Repository{
		Data: repoData,
	}
	type Context struct {
		Repository  *Repository
		Module      *Module
		Package     *Package
		ProjectRoot bool
	}

	// Execute the templates and, if they result in a change, queue them up for output
	type file struct {
		path     string
		contents []byte
	}
	var files []file
	var buf bytes.Buffer
	for imp, info := range pairs {
		buf.Reset()

		context := &Context{
			Repository:  repository,
			Module:      mod,
			Package:     PackageFromInfo(fset, info),
			ProjectRoot: repo == info.dir,
		}

		err := info.template.Execute(&buf, context)
		if err != nil {
			return fmt.Errorf("template for %s failed: %w", imp, err)
		}

		contents := buf.Bytes()
		// we only queue the write if there's been a change to the contents
		if !bytes.Equal(contents, info.oldReadme) {
			files = append(files, file{
				path:     filepath.Join(info.dir, "README.md"),
				contents: bytes.Clone(contents),
			})
		}
	}

	for _, file := range files {
		if err := os.WriteFile(file.path, file.contents, 0666); err != nil {
			return err
		}
	}

	return nil
}
