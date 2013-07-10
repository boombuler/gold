// Package that implements an engine for the gold parser system
package gold

import (
	"bufio"
	"fmt"
	"io"
)

// Implements parsing logic
type Parser interface {

	// reads the code from the reader and returns the syntax-tree or a parsing error
	// if trimReduction is set to true, tokens which have only one non-terminal sub-node are reduced.
	Parse(r io.Reader, trimReduce bool) (*Token, error)

	GetInformation() GrammarInformation
}

type GrammarInformation struct {
	Name    string
	Version string
	Author  string
	About   string
}

type parser struct {
	grammar      grammar
	isCgtGrammar bool
}

type grammar interface {
	getInformation() GrammarInformation
	readTokens(rd io.Reader) <-chan *parserToken
	getInitialLRState() *lrState
}

func (p parser) GetInformation() GrammarInformation {
	return p.grammar.getInformation()
}

type grammarError string

func (ge grammarError) Error() string {
	return string(ge)
}

// Creates a new parser by reading the grammar file from the passed Reader or an error.
// Currently only supports the cgt grammar files
func NewParser(grammar io.Reader) (Parser, error) {

	rd := bufio.NewReader(grammar)

	head, err := readString(rd)
	if err != nil {
		return nil, grammarError("Invalid grammar file format")
	}

	parser := new(parser)

	switch head {
	case cgtHeader:
		parser.grammar = loadCGTGrammar(rd)
		parser.isCgtGrammar = true
	case egtHeader:
		parser.grammar = loadEGTGrammar(rd)
		parser.isCgtGrammar = false
	default:
		return nil, grammarError("Unknown grammar file format: " + head)
	}
	if parser.grammar == nil {
		return nil, grammarError("Unable to read grammar file")
	}

	return parser, nil
}

func (p *parser) Parse(r io.Reader, trimReduce bool) (*Token, error) {
	input := p.grammar.readTokens(r)

	tokenStack := new(stack)
	stateStack := new(stack)

	stateStack.Push(p.grammar.getInitialLRState())
	var lastToken *parserToken = nil

	for nextToken := range input {
		lastToken = nextToken
		switch nextToken.Symbol.Kind {
		case stGroupStart, stCommentLine:
			continue
		case stNoise:
			continue
		case stError:
			return nil, &ParseError{Message: fmt.Sprintf("Unknown Token \"%s\"", nextToken.Text), Position: nextToken.Position}
		}

		tokenParsed := false
		for !tokenParsed {

			currentState := stateStack.Peek().(*lrState)

			action := currentState.Actions[nextToken.Symbol]

			if action == nil {
				return nil, &ParseError{Message: fmt.Sprintf("syntax Error: unexpected \"%s\"", nextToken.Text), Position: nextToken.Position}
			}

			switch action.Action {
			case actionShift:
				stateStack.Push(action.TargetState)
				tokenStack.Push(nextToken.toToken())
				tokenParsed = true

			case actionReduce:
				rule := action.TargetRule
				if trimReduce && len(rule.Symbols) == 1 && rule.Symbols[0].Kind == stNonTerminal {
					stateStack.Pop()
					currentState = stateStack.Peek().(*lrState)
				} else {
					tokenCnt := len(rule.Symbols)
					tokens := make([]*Token, tokenCnt)

					for idx, _ := range rule.Symbols {
						stateStack.Pop()
						tokens[tokenCnt-1-idx] = tokenStack.Pop().(*Token)
					}

					nttoken := &Token{
						Name:       rule.NonTerminal.String(),
						Text:       rule.String(),
						Tokens:     tokens,
						IsTerminal: false,
						Symbol:     SymbolId(rule.NonTerminal.Index),
						Rule:       RuleId(rule.Index),
					}
					tokenStack.Push(nttoken)

					currentState = stateStack.Peek().(*lrState)
				}
				gotoAction := currentState.Actions[rule.NonTerminal]
				stateStack.Push(gotoAction.TargetState)
			case actionAccept:
				return tokenStack.Pop().(*Token), nil
			}
		}
	}
	if lastToken != nil {
		return nil, &ParseError{Message: "Unexpected end of file", Position: lastToken.Position}
	}
	return nil, &ParseError{Message: "Unexpected end of file", Position: TextPosition{Line: 0, Column: 0}}
}
