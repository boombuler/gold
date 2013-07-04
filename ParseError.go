package gold

import "fmt"

// represents a error while parsing.
type ParseError struct {
	// contains a message which specifies what was wrong with the input string
	Message string
	// the position of the input text which produced the error
	Position TextPosition
}

// returns the error message as string
func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s at %s", pe.Message, pe.Position.String())
}
