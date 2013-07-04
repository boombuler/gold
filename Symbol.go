package gold

import "fmt"

type symbol struct {
	Index int
	Name  string
	Kind  symbolType
}

type symbolTable []*symbol

type symbolType byte

const (
	stNonTerminal  symbolType = 0 // Normal Nonterminal
	stTerminal     symbolType = 1 // Normal Terminal
	stWhitespace   symbolType = 2 // Whitespace Terminal
	stEnd          symbolType = 3 // End Character - End of File. This symbol is used to represent the end of the file or the end of the source input.
	stCommentStart symbolType = 4 // Start of a block quote
	stCommentEnd   symbolType = 5 // End of a block quote
	stCommentLine  symbolType = 6 // Line Comment Terminal
	stError        symbolType = 7 // Error Terminal. If the parser encounters an error reading a token, this kind of symbol can used to differentiate it from other terminal types.
)

func newSymbolTable(count int, createSymbols bool) symbolTable {
	result := make(symbolTable, count)
	if createSymbols {
		for i := 0; i < count; i++ {
			result[i] = new(symbol)
		}
	}
	return result
}

func (s symbol) String() string {
	if s.Kind == stNonTerminal {
		return fmt.Sprintf("<%s>", s.Name)
	}
	return s.Name
}
