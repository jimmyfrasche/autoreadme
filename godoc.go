package main

import (
	"bytes"
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
		Name: in.Name,
	}

	buf.Reset()
	buf.WriteString("```\n")
	format.Node(buf, fset, in.Code)
	buf.WriteString("\n```\n")
	out.Code = buf.String()

	if in.Output != "" {
		out.Output = "```\n" + in.Output + "\n```\n"
	}
	return out
}
