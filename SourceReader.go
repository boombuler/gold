package gold

import (
	"bufio"
	"fmt"
	"io"
)

type TextPosition struct {
	Line   int
	Column int
}

func (t TextPosition) String() string {
	return fmt.Sprintf("Line %d, Column %d", t.Line, t.Column)
}

type sourceReader struct {
	bufReader    *bufio.Reader
	SkipNextRead bool
	Rune         rune
	Position     TextPosition
}

func newSourceReader(r io.Reader) *sourceReader {
	result := &sourceReader{bufReader: bufio.NewReader(r)}
	result.Position.Line = 1
	result.Position.Column = 0
	result.SkipNextRead = false
	return result
}

func (r *sourceReader) Next() bool {
	if r.SkipNextRead {
		r.SkipNextRead = false
		return true
	}
	cur, _, err := r.bufReader.ReadRune()
	r.Rune = cur
	if err != nil {
		return false
	}
	if string(cur) == "\n" {
		r.Position.Line++
		r.Position.Column = 0
	} else {
		r.Position.Column++
	}

	return true
}
