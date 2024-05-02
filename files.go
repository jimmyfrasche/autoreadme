package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func readFile(path string) ([]byte, error) {
	bs, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return bs, nil
}

func repoPath(repo, file string) string {
	return filepath.Join(repo, ".github", "autoreadme", file)
}

func localPath(dir, ext string) string {
	var dot string
	if ext != "" {
		dot = "."
	}
	return filepath.Join(dir, "README.md"+dot+ext)
}

func toJson(bs []byte, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, nil
	}
	var v any
	err = json.Unmarshal(bs, &v)
	return v, err
}

func fromLines(bs []byte, err error) (map[string]bool, error) {
	if err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, nil
	}
	lines := strings.Split(string(bs), "\n")
	m := map[string]bool{}
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			m[line] = true
		}
	}
	return m, nil
}

func RepoTemplate(repo string) ([]byte, error) {
	return readFile(repoPath(repo, "README.md.template"))
}

func RepoData(repo string) (any, error) {
	return toJson(readFile(repoPath(repo, "README.md.data")))
}

func RepoIgnore(repo string) (map[string]bool, error) {
	return fromLines(readFile(repoPath(repo, "autoreadme.ignore")))
}

func DirTemplate(dir string) ([]byte, error) {
	return readFile(localPath(dir, "template"))
}

func DirData(dir string) (any, error) {
	return toJson(readFile(localPath(dir, "data")))
}

func DirReadme(dir string) ([]byte, error) {
	return readFile(localPath(dir, ""))
}
