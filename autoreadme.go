//Copyright 2013 James Frasche. All rights reserved.
//Use of this code is governed by a BSD-License found in the LICENSE file.

//Automatically generate a github README.md for your Go project.
//
//autoreadme(1) creates a github-formatted README.md using the same format as godoc(1).
//It includes the package summary and generates badges for godoc.org for the complete
//documentation and for travis-ci, if there's a .travis.yml file in the directory.
//It also includes copy-pastable installation instructions using the go(1) tool.
//
//HEURISTICS
//
//autoreadme(1) by default imports the Go code in the current directory, unless a directory is specified.
//
//If the -template argument is not given, it tries to use the README.md.template file in the same
//directory. If no such file exists, the built in template is used. These rules apply to each
//directory visited when -r is specified to run autoreadme(1) recursively. If a README.md already
//exists, it fails unless -f is specified.
//
//EXAMPLES
//
//To create a README.md for the directory a/b/c
//	autoreadme a/b/c
//
//To overwrite the README.md in the current directory
//	autoreadme -f
//
//To run in the current directory and all subdirectories that contain
//Go code
//	autoreadme -r
//
//Use the built in template as the basis for a custom template.
//	autoreadme -print-template >README.md.template
//
//To override both the default template and a local README.md.template
//	autoreadme -template=path/to/readme.template
//
//TEMPLATE VARIABLES
//
//If you wish to use your own template, These are the fields available to dot:
//
//	Name - The name of your packages.
//
//	Doc - The package-level documentation of your package.
//
//	Synopsis - The first sentence of .Doc.
//
//	Import - The import path of your package.
//
//	RepoPath - The import path without the github.com/ prefix.
//
//	Bugs - a []string of all bugs as per godoc.
//
//	Library - True if not a command.
//
//	Command - True if a command.
//
//	Today - The current date in YYYY.MM.DD format.
//
//	Travis - True if there is a .travis.yml file in the same directory
//		as your package.
//
//	Example - a map[name]Example with all examples from the _test files. These
//		can be used to include selective examples into the README.
//		The Example{} struct has these fields:
//			Name - name of the example
//			Code - renders example code similar to godoc
//			Output - example output, if any
//
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	Force         = flag.Bool("f", false, "Run even if README.md exists, overwriting original")
	Recursive     = flag.Bool("r", false, "Run in all subdirectories containing Go code")
	PrintTemplate = flag.Bool("print-template", false, "write the built in template to stdout and exit")
	Template      = flag.String("template", "", "specify a file to use as template, overrides built in template and README.md.template")
	Defs          defFlag
)

type Doc struct {
	Name, Import, Synopsis, Doc, Today string
	RepoPath                           string //Import sans github.com
	Bugs                               []string
	Library, Command                   bool
	Travis                             bool //if a .travis.yml file is statable
	Example                            map[string]Example
}

// Map returns the receiver as a map.
func (d Doc) Map() map[string]interface{} {
	return map[string]interface{}{
		"Name":     d.Name,
		"Import":   d.Import,
		"Synopsis": d.Synopsis,
		"Doc":      d.Doc,
		"Today":    d.Today,
		"RepoPath": d.RepoPath,
		"Bugs":     d.Bugs,
		"Library":  d.Library,
		"Command":  d.Command,
		"Travis":   d.Travis,
		"Example":  d.Example,
	}
}

type Example struct {
	Name   string
	Code   string
	Output string //the expected output, if not empty
}

func today() string {
	return time.Now().Format("2006.01.02")
}

var gopaths = build.Default.SrcDirs()

func importPath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for _, p := range gopaths {
		p, err = filepath.Rel(p, abs)
		if err != nil {
			return "", err
		}
		//not a relative path, therefore the correct one
		if len(p) > 0 && p[0] != '.' {
			return filepath.ToSlash(p), nil
		}
	}
	return "", fmt.Errorf("Not in go root or element of $GOPATH: %s", dir)
}

func fmtDoc(doc string) string {
	var acc []string
	push := func(ss ...string) {
		acc = append(acc, ss...)
	}
	nl := func() {
		push("\n")
	}
	for _, b := range blocks(doc) {
		ls := b.lines
		switch b.op {
		case opPara:
			push(ls...)
			nl()
		case opHead:
			push("## ")
			push(ls...)
			nl()
		case opPre:
			push("```")
			nl()
			push(ls...)
			push("```")
			nl()
			nl()
		}
	}
	return strings.Join(acc, "")
}

func getDoc(dir string) (Doc, error) {
	bi, err := build.ImportDir(dir, 0)
	if err != nil {
		return Doc{}, nil
	}

	ip, err := importPath(dir)
	if err != nil {
		return Doc{}, err
	}

	filter := func(fi os.FileInfo) bool {
		if fi.IsDir() {
			return false
		}
		nm := fi.Name()
		for _, f := range append(bi.GoFiles, bi.CgoFiles...) {
			if nm == f {
				return true
			}
		}
		return false
	}

	pkgs, err := parser.ParseDir(token.NewFileSet(), bi.Dir, filter, parser.ParseComments)
	if err != nil {
		return Doc{}, err
	}

	pkg := pkgs[bi.Name]
	docs := doc.New(pkg, bi.ImportPath, 0)

	examples, err := renderExamples(bi)
	if err != nil {
		return Doc{}, err
	}

	bugs := []string{}
	for _, bug := range docs.Notes["BUG"] {
		bugs = append(bugs, bug.Body)
	}

	name := bi.Name
	if name == "main" {
		name = filepath.Base(bi.Dir)
	}

	//get import path without the github.com/
	pathelms := strings.Split(ip, "/")[1:]
	repo := path.Join(pathelms...)

	return Doc{
		Name:     name,
		Import:   ip,
		Synopsis: bi.Doc,
		Doc:      fmtDoc(docs.Doc),
		Example:  examples,
		Today:    today(),
		RepoPath: repo,
		Bugs:     bugs,
		Library:  bi.Name != "main",
		Command:  bi.Name == "main",
	}, nil
}

func renderExamples(bi *build.Package) (map[string]Example, error) {
	examples := map[string]Example{}
	testFilenames := append(bi.TestGoFiles, bi.XTestGoFiles...)
	fset := token.NewFileSet()
	for _, filename := range testFilenames {
		path := filepath.Join(bi.Dir, filename)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, ex := range doc.Examples(file) {
			examples[ex.Name] = renderExample(ex)
		}
	}
	return examples, nil
}

func renderExample(ex *doc.Example) Example {
	e := Example{
		Name: strings.Replace(ex.Name, "_", " ", -1),
	}

	c := &bytes.Buffer{}
	format.Node(c, token.NewFileSet(), ex.Code)
	e.Code = fmt.Sprintf("Code:\n\n```\n%s\n```\n", c.String())

	if ex.Output != "" {
		e.Output = fmt.Sprintf("Output:\n\n```\n%s\n```\n", ex.Output)
	}

	return e
}

//only read and parse specified template once if -template and -r specified
var cached *template.Template

func getTemplate(dir string) (*template.Template, error) {
	if *Template != "" {
		if cached != nil {
			return cached, nil
		}
		bs, err := ioutil.ReadFile(*Template)
		if err != nil {
			return nil, err
		}
		cached, err = template.New("").Parse(string(bs))
		if err != nil {
			return nil, err
		}
		return cached, nil
	}

	bs, err := ioutil.ReadFile(filepath.Join(dir, "README.md.template"))
	//local template file found, prefer
	if err == nil {
		return template.New("").Parse(string(bs))
	}
	//the file was found but something else happened
	if !os.IsNotExist(err) {
		return nil, err
	}

	//the file was not found and no template specified, so use default
	return tmpl, nil
}

func getFile(dir string) (*os.File, error) {
	nm := filepath.Join(dir, "README.md")
	if !*Force {
		_, err := os.Stat(nm)
		if err == nil {
			return nil, fmt.Errorf("README.md already exists at %s. Use -f to overwrite", dir)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return os.Create(nm)
}

func hasTravisYml(dir string) (stated bool) {
	if _, err := os.Stat(filepath.Join(dir, ".travis.yml")); err == nil {
		stated = true
	}
	return
}

func dirs(start string) (ds []string, err error) {
	if !*Recursive {
		return []string{start}, nil
	}

	var follow func(string) error
	follow = func(dir string) error {
		fis, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}

		hasgo := false
		for _, fi := range fis {
			nm := fi.Name()
			if fi.IsDir() {
				if nm[0] != '.' {
					if err := follow(nm); err != nil {
						return err
					}
				}
			} else if strings.HasSuffix(nm, ".go") {
				hasgo = true
			}
		}
		if hasgo {
			ds = append(ds, dir)
		}
		return nil
	}

	if err = follow(start); err != nil {
		return nil, err
	}

	return ds, nil
}

func init() {
	log.SetFlags(0)
}

func init() {
	flag.Usage = func() {
		log.Printf("Usage of %s: %s [directory]", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
}

//Usage: %name %flags [directory]
func main() {
	flag.Var(&Defs, "def", "Template define having the form: name=value")
	flag.Parse()

	where := "."

	if args := flag.Args(); len(args) > 1 {
		log.Fatalln("Too many arguments: ", args[1:])
		flag.Usage()
	} else if len(args) == 1 {
		where = args[0]
	}

	where, err := filepath.Abs(where)
	if err != nil {
		log.Fatalln(err)
	}

	if *PrintTemplate {
		fmt.Print(tmplraw)
		return
	}

	ds, err := dirs(where)
	if err != nil {
		log.Fatalln(err)
	}

	warned := false
	warn := func(dir string, err error) {
		warned = true
		log.Println("Could not create README.md for", dir, "because:", err)
	}
	for _, dir := range ds {
		tmpl, err := getTemplate(dir)
		if err != nil {
			if *Template != "" && *Recursive {
				//don't want to show same error message for all dirs so just bail now
				log.Fatalln(err)
			}
			warn(dir, err)
			continue
		}

		doc, err := getDoc(dir)
		if err != nil {
			warn(dir, err)
			continue
		}

		doc.Travis = hasTravisYml(dir)

		f, err := getFile(dir)
		if err != nil {
			warn(dir, err)
			continue
		}

		// Convert the doc to a map, so we can add additional fields
		docm := doc.Map()
		for _, d := range Defs {
			// Lowercase define
			docm[d.Name] = d.Value
			// Uppercase define
			d = d.UpperClone()
			docm[d.Name] = d.Value
		}

		if err = tmpl.Execute(f, docm); err != nil {
			warn(dir, err)
		}
	}
	if warned {
		os.Exit(1)
	}
}
