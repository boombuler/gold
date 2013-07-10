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
	unreadBuffer *stack
	Rune         rune
	Position     TextPosition
}

func newSourceReader(r io.Reader) *sourceReader {
	result := &sourceReader{bufReader: bufio.NewReader(r)}
	result.Position.Line = 1
	result.Position.Column = 0
	result.unreadBuffer = newStack()
	return result
}

func (r *sourceReader) Next() bool {
	if r.unreadBuffer.count > 0 {
		r.Rune = r.unreadBuffer.Pop().(rune)
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

func (sr *sourceReader) UnreadAll(runes []rune) {
	for i := len(runes) - 1; i >= 0; i-- {
		sr.unreadBuffer.Push(runes[i])
	}
}

func (sr *sourceReader) UnreadLast() {
	sr.unreadBuffer.Push(sr.Rune)
}
