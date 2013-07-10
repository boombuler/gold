package gold

import (
	"bufio"
	"bytes"
	"io"
)

func (g *cgtGrammar) readTokens(rd io.Reader) <-chan *parserToken {
	c := make(chan *parserToken)
	r := newSourceReader(rd)
	go func() {
		for {
			t := g.readToken(r, true)

			switch t.Symbol.Kind {
			case stNonTerminal, stTerminal:
				c <- t

			case stNoise, stGroupStart, stGroupEnd, stCommentLine:
				// do nothing
			case stError, stEnd:
				c <- t
				close(c)
				return
			}
		}
	}()
	return c
}

func (g *cgtGrammar) readToken(r *sourceReader, readComments bool) *parserToken {

	dfa := g.getInitialDfaState()

	tText := new(bytes.Buffer)
	tWriter := bufio.NewWriter(tText)

	result := new(parserToken)
	result.Text = ""
	result.Symbol = g.errorSymbol
	result.Position = r.Position

	for {
		if !r.Next() {
			tWriter.Flush()
			if r.Rune == 0 && tText.Len() == 0 {
				result.Text = ""
				result.Symbol = g.endSymbol
			}
			return result
		}

		nextState, ok := dfa.TransitionVector(r.Rune)
		if ok {
			tWriter.WriteRune(r.Rune)

			dfa = nextState
			if dfa.AcceptSymbol != nil {
				tWriter.Flush()
				result.Text = string(tText.Bytes())
				result.Symbol = dfa.AcceptSymbol
			}
		} else {
			if result.Symbol == g.errorSymbol {
				tWriter.WriteRune(r.Rune)
			}
			r.UnreadLast()
			break
		}
	}

	tWriter.Flush()
	result.Text = string(tText.Bytes())

	if readComments {
		switch result.Symbol.Kind {
		case stCommentLine:
			result.Text += g.readLineComment(r)
		case stGroupStart:
			result.Text += g.readBlockComment(r)
		}
	}

	return result
}

func (g *cgtGrammar) readBlockComment(r *sourceReader) string {
	buff := new(bytes.Buffer)

	for {
		token := g.readToken(r, false)

		symbolType := token.Symbol.Kind

		switch symbolType {

		case stEnd, stGroupEnd:
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

func (g *cgtGrammar) readLineComment(r *sourceReader) string {
	result := new(bytes.Buffer)
	exit := false
	for r.Next() {
		if r.Rune == '\n' || r.Rune == '\r' {
			exit = true
		} else {
			if exit {
				r.UnreadLast()
				break
			}
			result.WriteRune(r.Rune)
		}
	}
	return string(result.Bytes())
}
