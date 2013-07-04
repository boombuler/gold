package gold

import (
	"bufio"
	"bytes"
	"io"
)
import "fmt"

// Implements parsing logic
type Parser interface {
	// reads the code from the reader and returns the syntax-tree or a parsing error
	// if trimReduction is set to true, tokens which have only one non-terminal sub-node are reduced.
	Parse(r io.Reader, trimReduce bool) (*Token, error)
}

type parser struct {
	grammar *cgtGramar
}

// Creates a new parser by reading the grammar file from the passed Reader or an error
func NewParser(grammar io.Reader) (Parser, error) {
	gr := loadCGTGramar(grammar)
	if gr == nil {
		return nil, cgtFormatError
	}
	parser := new(parser)
	parser.grammar = gr

	return parser, nil
}

func (p *parser) readTokens(rd io.Reader) <-chan *parserToken {
	c := make(chan *parserToken)
	r := newSourceReader(rd)
	go func() {
		for {
			t := p.readToken(r, true)
			c <- t
			if t.Symbol.Kind == stEnd || t.Symbol.Kind == stError {
				close(c)
				return
			}
		}
	}()
	return c
}

func (p *parser) readToken(r *sourceReader, readComments bool) *parserToken {

	dfa := p.grammar.getInitialDfaState()

	tText := new(bytes.Buffer)
	tWriter := bufio.NewWriter(tText)

	result := new(parserToken)
	result.Text = ""
	result.Symbol = p.grammar.errorSymbol
	result.Position = r.Position

	for {
		if !r.Next() {
			tWriter.Flush()
			if r.Rune == 0 && tText.Len() == 0 {
				result.Text = ""
				result.Symbol = p.grammar.endSymbol
			}
			return result
		}

		nextState, ok := dfa.TransitionVector[r.Rune]
		if ok {
			tWriter.WriteRune(r.Rune)

			dfa = nextState
			if dfa.AcceptSymbol != nil {
				tWriter.Flush()
				result.Text = string(tText.Bytes())
				result.Symbol = dfa.AcceptSymbol
			}
		} else {
			if result.Symbol == p.grammar.errorSymbol {
				tWriter.WriteRune(r.Rune)
			}

			r.SkipNextRead = true
			break
		}
	}

	tWriter.Flush()
	result.Text = string(tText.Bytes())

	if readComments {
		switch result.Symbol.Kind {
		case stCommentLine:
			result.Text += p.readLineComment(r)
		case stCommentStart:
			result.Text += p.readBlockComment(r)
		}
	}

	return result
}

func (p *parser) readBlockComment(r *sourceReader) string {
	buff := new(bytes.Buffer)

	for {
		token := p.readToken(r, false)

		symbolType := token.Symbol.Kind

		switch symbolType {

		case stEnd, stCommentEnd:
			buff.WriteString(token.Text)
			return string(buff.Bytes())
		case stError:
			buff.WriteString(token.Text)
			r.Next()
		default:
			buff.WriteString(token.Text)
		}
	}

	return string(buff.Bytes())
}

func (p *parser) readLineComment(r *sourceReader) string {
	result := new(bytes.Buffer)
	exit := false
	for r.Next() {
		sRune := string(r.Rune)
		if sRune == "\n" || sRune == "\r" {
			exit = true
		} else {
			if exit {
				r.SkipNextRead = true
				break
			}
			result.WriteRune(r.Rune)
		}
	}
	return string(result.Bytes())
}

func (p *parser) Parse(r io.Reader, trimReduce bool) (*Token, error) {
	input := p.readTokens(r)

	tokenStack := new(stack)
	stateStack := new(stack)

	stateStack.Push(p.grammar.getInitialLRState())
	var lastToken *parserToken = nil

	for nextToken := range input {
		lastToken = nextToken
		switch nextToken.Symbol.Kind {
		case stCommentStart, stCommentLine:
			continue
		case stWhitespace:
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

					nttoken := &Token{Name: rule.NonTerminal.String(), Text: rule.String(), Tokens: tokens, IsTerminal: false}
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
