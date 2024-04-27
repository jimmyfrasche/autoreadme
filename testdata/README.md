# Writing test cases

Test cases are run from `TestAutoreadme` in `autoreadme_test.go`. Each test case is a [txtar](https://pkg.go.dev/golang.org/x/tools/txtar) file containing a complete Go project. autoreadme is run on the project, but with the default template replaced by `testdata/test.template`. Each package in the `.txtar` must contain a `README.md.expect`. The generated `README.md` is compared against the `README.md.expect`.

Each `X.txtar` is exposed as the subtest `TestAutoreadme/X`.
