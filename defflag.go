package main

import (
	"errors"
	"strings"
	"unicode"
)

type def struct {
	Name  string
	Value string
}

// UpperClone returns a clone of the receiver, with the first character of the
// name in upper case.
func (d def) UpperClone() def {
	rs := []rune(d.Name)
	return def{
		Name:  string(unicode.ToUpper(rs[0])) + string(rs[1:]),
		Value: d.Value,
	}
}

type defFlag []def

func (df *defFlag) String() string {
	return "" // Not used
}

func (df *defFlag) Set(value string) error {
	// Make sure it has the form $name=$value
	i := strings.Index(value, "=")
	if i <= 0 {
		return errors.New("invalid def flag: missing equals token")
	}
	*df = append(*df, def{
		Name:  strings.ToLower(value[:i]),
		Value: value[i+1:],
	})
	return nil
}
