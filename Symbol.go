package gold

import "fmt"

type symbol struct {
	Index uint16
	Name  string
	Kind  symbolType
	Group *group
}

type symbolTable []*symbol

type symbolType byte

const (
	stNonTerminal symbolType = 0 // Normal Nonterminal
	stTerminal    symbolType = 1 // Normal Terminal
	stNoise       symbolType = 2 // Noise terminal. These are ignored by the parser. Comments and whitespace are considered 'noise'.
	stEnd         symbolType = 3 // End Character - End of File. This symbol is used to represent the end of the file or the end of the source input.
	stGroupStart  symbolType = 4 // Start of a block quote
	stGroupEnd    symbolType = 5 // End of a block quote
	stCommentLine symbolType = 6 // Line Comment Terminal
	stError       symbolType = 7 // Error Terminal. If the parser encounters an error reading a token, this kind of symbol can used to differentiate it from other terminal types.
)

func newSymbolTable(count uint16, createSymbols bool) symbolTable {
	result := make(symbolTable, count)
	if createSymbols {
		var i uint16
		for i = 0; i < count; i++ {
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
