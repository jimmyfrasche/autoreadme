package main

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
)

func Synopsis(text string) string {
	var p doc.Package
	return p.Synopsis(text)
}

func ParseDoc(text string) *comment.Doc {
	var p comment.Parser
	return p.Parse(text)
}

type Doc struct {
	Synopsis string
	*comment.Doc
}

func NewDoc(text string) *Doc {
	return &Doc{
		Synopsis: Synopsis(text),
		Doc:      ParseDoc(text),
	}
}

func epsilon(*comment.Heading) string {
	return ""
}

func (d *Doc) Markdown(headingLevel int) string {
	p := &comment.Printer{
		HeadingLevel: headingLevel,
		HeadingID:    epsilon,
	}
	return string(p.Markdown(d.Doc))
}

func renderExample(buf *bytes.Buffer, fset *token.FileSet, in *doc.Example) Example {
	out := Example{
		Playable: in.Play != nil,
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
