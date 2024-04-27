package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var errDone = errors.New("done")
var errTop = errors.New("top of fs")

func up(done func(dir string) error) error {
	top := false
	p, err := os.Getwd()
	if err != nil {
		return err
	}
	p = filepath.Clean(p)

	for {
		// check our predicate for this level
		if err := done(p); err != nil {
			return err
		}

		// already at the top but found nothing
		if top {
			return errTop
		}

		// go up one level
		// make a note if we're at the top
		new := filepath.Dir(p)
		top = new == p
		p = new
	}
}

func Roots() (repo, project string, err error) {
	stat := func(dir, p string) (os.FileInfo, error) {
		return os.Stat(filepath.Join(dir, p))
	}
	err = up(func(dir string) error {
		if repo == "" {
			fi, err := stat(dir, ".git")
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			if err == nil && fi.IsDir() {
				repo = dir
			}
		}

		if project == "" {
			fi, err := stat(dir, "go.mod")
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			if err == nil && fi.Mode().IsRegular() {
				project = dir
			}
		}

		// found both
		if repo != "" && project != "" {
			return errDone
		}
		return nil
	})
	// up always returns an error
	switch err {
	case errDone:
		return repo, project, nil

	case errTop:
		// no errors but missing go.mod and/or .git
		if project == "" {
			err = errors.New("could not find repository root")
		} else {
			err = errors.New("could not find project or repository root")
		}

	default:
		// stat error while searching for go.mod and .git
		if project == "" {
			err = fmt.Errorf("could not find repository root: %w", err)
		} else {
			err = fmt.Errorf("could not find project or repository root: %w", err)
		}
	}
	return "", "", err
}
