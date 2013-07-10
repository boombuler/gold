package gold

import (
	"bufio"
	"fmt"
	"io"
)

// represents a position in the input text
type TextPosition struct {

	// The Line, starts with line 1
	Line int
	// The Column, starts with column 1
	Column int
}

// returns a string representing the textposition
func (t TextPosition) String() string {
	return fmt.Sprintf("Line %d, Column %d", t.Line, t.Column)
}

type sourceReader struct {
	bufReader    *bufio.Reader
	unreadBuffer []rune
	Rune         rune
	Position     TextPosition
}

func newSourceReader(r io.Reader) *sourceReader {
	result := &sourceReader{bufReader: bufio.NewReader(r)}
	result.Position.Line = 1
	result.Position.Column = 0
	result.unreadBuffer = make([]rune, 0)
	return result
}

func (r *sourceReader) Next() bool {
	if len(r.unreadBuffer) > 0 {
		r.Rune = r.unreadBuffer[0]
		r.unreadBuffer = r.unreadBuffer[1:]
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

func (sr *sourceReader) UnreadAll(r []rune) {
	result := make([]rune, len(r)+len(sr.unreadBuffer))
	copy(result, r)
	copy(result[len(r):], sr.unreadBuffer)

	sr.unreadBuffer = result
}

func (sr *sourceReader) UnreadLast() {
	result := make([]rune, len(sr.unreadBuffer)+1)
	result[0] = sr.Rune
	copy(result[1:], sr.unreadBuffer)
	sr.unreadBuffer = result
}
