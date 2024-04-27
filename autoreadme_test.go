package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/txtar"
)

var join = filepath.Join

func mkdirp(t *testing.T, path ...string) {
	t.Helper()
	if err := os.MkdirAll(join(path...), 0777); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, data []byte, path ...string) {
	t.Helper()
	if err := os.WriteFile(join(path...), data, 0666); err != nil {
		t.Fatal(err)
	}
}

func cd(t *testing.T, path ...string) {
	t.Helper()
	if err := os.Chdir(join(path...)); err != nil {
		t.Fatalf("failed to cd: %s", err)
	}
}

// See testdata/README.md for instructions on how to add additional test cases.
func TestAutoreadme(t *testing.T) {
	// cd back to the original directory when we're done
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// switch out the default template for one specialized for testing
	src, err := readFile(join("testdata", "test.template"))
	if err != nil {
		t.Fatalf("reading testdata/test.template: %s", err)
	}
	originalSrc := defaultTemplateSrc
	defaultTemplateSrc = string(src)
	t.Cleanup(func() {
		defaultTemplateSrc = originalSrc
	})

	// use testdata/*.txtar files to build out scenarios for us to test
	cases, err := filepath.Glob(join("testdata", "*.txtar"))
	if err != nil {
		t.Fatalf("getting testdata/*.txtar files: %s", err)
	}
	for _, c := range cases {
		name := filepath.Base(c)
		name = name[:len(name)-len(".txtar")]

		t.Run(name, func(t *testing.T) {
			cd(t, originalDir)

			arc, err := txtar.ParseFile(c)
			if err != nil {
				t.Fatal(err)
			}

			tmp := t.TempDir()
			cd(t, tmp)

			// always create a .git/ directory at the root
			// to avoid having to specify it in every single test case
			mkdirp(t, ".git")

			// write out the archive in tmp
			for _, file := range arc.Files {
				name := filepath.FromSlash(file.Name)
				dir := filepath.Dir(name)
				if dir != "" {
					mkdirp(t, dir)
				}
				writeFile(t, file.Data, name)
			}

			// do the thing
			if err = Main(context.Background()); err != nil {
				t.Fatalf("running autoreadme: %s", err)
			}

			// collect all generated README.md and README.md.expect contents, keyed by location
			expect := map[string]string{}
			got := map[string]string{}
			err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				name := d.Name()
				switch name {
				default:
					return nil
				case "README.md", "README.md.expect":
				}
				bs, err := readFile(path)
				if err != nil {
					return err
				}
				contents := string(bs)
				path = filepath.Dir(path)
				if name == "README.md" {
					expect[path] = contents
				} else {
					got[path] = contents
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			for k, v := range expect {
				if _, ok := got[k]; !ok {
					t.Errorf("README.md.expect with no README.md at %s\n %s", k, v)
				}
			}
			for k, v := range got {
				v2, ok := expect[k]
				if !ok {
					t.Errorf("README.md without README.md.expect at %s:\n%s", k, v)
				} else if v != v2 {
					t.Errorf("README.md difference at %s:\nGOT:\n%s\nEXPECTED:\n%s", k, v, v2)
				}
			}
		})
	}
}
