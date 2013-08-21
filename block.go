package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

//everything in this file was copied from the go/doc package as this is not exported
//It can be found in go/doc/comment.go

func indentLen(s string) int {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

func isBlank(s string) bool {
	return len(s) == 0 || (len(s) == 1 && s[0] == '\n')
}

func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[0:i]
}

func unindent(block []string) {
	if len(block) == 0 {
		return
	}

	// compute maximum common white prefix
	prefix := block[0][0:indentLen(block[0])]
	for _, line := range block {
		if !isBlank(line) {
			prefix = commonPrefix(prefix, line[0:indentLen(line)])
		}
	}
	n := len(prefix)

	// remove
	for i, line := range block {
		if !isBlank(line) {
			block[i] = line[n:]
		}
	}
}

// heading returns the trimmed line if it passes as a section heading;
// otherwise it returns the empty string.
func heading(line string) string {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return ""
	}

	// a heading must start with an uppercase letter
	r, _ := utf8.DecodeRuneInString(line)
	if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
		return ""
	}

	// it must end in a letter or digit:
	r, _ = utf8.DecodeLastRuneInString(line)
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		return ""
	}

	// exclude lines with illegal characters
	if strings.IndexAny(line, ",.;:!?+*/=()[]{}_^°&§~%#@<\">\\") >= 0 {
		return ""
	}

	// allow "'" for possessive "'s" only
	for b := line; ; {
		i := strings.IndexRune(b, '\'')
		if i < 0 {
			break
		}
		if i+1 >= len(b) || b[i+1] != 's' || (i+2 < len(b) && b[i+2] != ' ') {
			return "" // not followed by "s "
		}
		b = b[i+2:]
	}

	return line
}

type op int

const (
	opPara op = iota
	opHead
	opPre
)

type block struct {
	op    op
	lines []string
}

func blocks(text string) []block {
	var (
		out  []block
		para []string

		lastWasBlank   = false
		lastWasHeading = false
	)

	close := func() {
		if para != nil {
			out = append(out, block{opPara, para})
			para = nil
		}
	}

	lines := strings.SplitAfter(text, "\n")
	unindent(lines)
	for i := 0; i < len(lines); {
		line := lines[i]
		if isBlank(line) {
			// close paragraph
			close()
			i++
			lastWasBlank = true
			continue
		}
		if indentLen(line) > 0 {
			// close paragraph
			close()

			// count indented or blank lines
			j := i + 1
			for j < len(lines) && (isBlank(lines[j]) || indentLen(lines[j]) > 0) {
				j++
			}
			// but not trailing blank lines
			for j > i && isBlank(lines[j-1]) {
				j--
			}
			pre := lines[i:j]
			i = j

			unindent(pre)

			// put those lines in a pre block
			out = append(out, block{opPre, pre})
			lastWasHeading = false
			continue
		}

		if lastWasBlank && !lastWasHeading && i+2 < len(lines) &&
			isBlank(lines[i+1]) && !isBlank(lines[i+2]) && indentLen(lines[i+2]) == 0 {
			// current line is non-blank, surrounded by blank lines
			// and the next non-blank line is not indented: this
			// might be a heading.
			if head := heading(line); head != "" {
				close()
				out = append(out, block{opHead, []string{head}})
				i += 2
				lastWasHeading = true
				continue
			}
		}

		// open paragraph
		lastWasBlank = false
		lastWasHeading = false
		para = append(para, lines[i])
		i++
	}
	close()

	return out
}
