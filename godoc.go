package main

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
	"strings"
)

func Synopsis(text string) string {
	d := ParseDoc(text)
	text = string(newPrinter(1).Text(d))
	var p doc.Package
	return p.Synopsis(text)
}

func ParseDoc(text string) *comment.Doc {
	p := &comment.Parser{
		LookupPackage: func(string) (string, bool) {
			return "", true
		},
		LookupSym: func(string, string) bool {
			return true
		},
	}
	return p.Parse(text)
}

type Doc struct {
	Empty    bool
	Synopsis string
	*comment.Doc
}

func NewDoc(text string) *Doc {
	return &Doc{
		Empty:    strings.TrimSpace(strings.ReplaceAll(text, "\n", "")) == "",
		Synopsis: Synopsis(text),
		Doc:      ParseDoc(text),
	}
}

func epsilon(*comment.Heading) string {
	return ""
}

func noLink(*comment.DocLink) string {
	return ""
}

func newPrinter(headingLevel int) *comment.Printer {
	return &comment.Printer{
		HeadingLevel: headingLevel,
		HeadingID:    epsilon,
		DocLinkURL:   noLink,
	}
}

func (d *Doc) Markdown(headingLevel int) string {
	p := newPrinter(headingLevel)
	return string(p.Markdown(d.Doc))
}

func renderExample(buf *bytes.Buffer, fset *token.FileSet, in *doc.Example) Example {
	out := Example{
		Name:          in.Name,
		Documentation: NewDoc(in.Doc),
		Playable:      in.Play != nil,
		Unordered:     in.Unordered,
		EmptyOutput:   in.EmptyOutput,
	}

	code := []any{in.Play}
	if !out.Playable {
		code = []any{}
		for _, line := range in.Code.(*ast.BlockStmt).List {
			code = append(code, line)
		}
	}

	buf.Reset()
	buf.WriteString("```go\n")
	for _, line := range code {
		format.Node(buf, fset, line)

		// playable examples end with a newline already
		if !out.Playable {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("```\n")
	out.Code = buf.String()

	if in.Output != "" {
		out.Output = "```\n" + in.Output + "```\n"
	}
	return out
}
