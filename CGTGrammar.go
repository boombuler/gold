package gold

import (
	"bufio"
	"bytes"
	"io"
)

const (
	cgtHeader = "GOLD Parser Tables/v1.0"
)

const (
	cgtRIdCharSets    recordId = 67 // C
	cgtRIdParameters  recordId = 80 // P
	cgtRIdTableCounts recordId = 84 // T
)

type cgtGrammar struct {
	goldGrammar
}

type cgtRecEntryTyp byte

func loadCGTGrammar(rd *bufio.Reader) *cgtGrammar {
	g := new(cgtGrammar)

	for curRec := range readRecords(rd) {
		recTyp := recordId((<-curRec).asByte())

		switch recTyp {
		case cgtRIdParameters:
			g.Name = (<-curRec).asString()
			g.Version = (<-curRec).asString()
			g.Author = (<-curRec).asString()
			g.About = (<-curRec).asString()
			<-curRec
			<-curRec
		case cgtRIdTableCounts:
			g.symbols = newSymbolTable((<-curRec).asInt(), true)
			g.charSets = make([]charSet, (<-curRec).asInt())
			g.rules = newRuleTable((<-curRec).asInt())
			g.dfaStates = newDFAStateTable((<-curRec).asInt())
			g.lrStates = newLRStateTable((<-curRec).asInt())
		case cgtRIdCharSets:
			idx := (<-curRec).asInt()
			g.charSets[idx] = fixedCharSet((<-curRec).asString())

		default:
			g.goldGrammar.loadRecord(recordId(recTyp), curRec)
		}
	}

	return g
}

func (g *cgtGrammar) readTokens(rd io.Reader) <-chan *parserToken {
	c := make(chan *parserToken)
	r := newSourceReader(rd)
	go func() {
		for {
			t := g.readToken(r)

			switch t.Symbol.Kind {
			case stNonTerminal, stTerminal:
				c <- t
			case stCommentLine:
				t.Text += g.readLineComment(r)
			case stGroupStart:
				t.Text += g.readBlockComment(r)
			case stNoise, stGroupEnd:
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

func (g *cgtGrammar) readBlockComment(r *sourceReader) string {
	buff := new(bytes.Buffer)

	for {
		token := g.readToken(r)

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
