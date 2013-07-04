package gold

type parserToken struct {
	// Token symbol.
	Symbol *symbol

	// Token text.
	Text string

	// Position within the source
	Position TextPosition
}

// Represents a node of a parsed syntax-tree
type Token struct {
	// contains the name of the token
	Name string
	// if the token is a terminal, Text contains the text of the terminal.
	// if the token is a non-terminal, Text contains the rule used to parse the non-terminal
	Text string
	// contains the sub-nodes of the syntax-tree node
	Tokens []*Token
	// tells if the token is a terminal or a non-terminal
	IsTerminal bool
}

func (pt *parserToken) toToken() *Token {
	return &Token{Name: pt.Symbol.String(), Text: pt.Text, Tokens: nil, IsTerminal: pt.Symbol.Kind == stTerminal}
}
