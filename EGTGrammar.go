package gold

import (
	"bufio"
	"io"
)

const (
	egtHeader = "GOLD Parser Tables/v5.0"
)

type egtGrammar struct {
	goldGrammar

	groups groupTable
}

const (
	egtRIdCharSet    recordId = 99  // c
	egtRIdGroup      recordId = 103 // g
	egtRIdProperty   recordId = 112 // p
	egtRIdTableCount recordId = 116 // t
)

func loadEGTGrammar(rd *bufio.Reader) *egtGrammar {
	g := new(egtGrammar)

	records := readRecords(rd)

	for curRec := range records {
		// process record...
		recTyp := recordId((<-curRec).asByte())

		switch recTyp {
		case egtRIdProperty:
			idx := (<-curRec).asInt()
			<-curRec // skip name...
			value := (<-curRec).asString()

			switch idx {
			case 0:
				g.Name = value
			case 1:
				g.Version = value
			case 2:
				g.Author = value
			case 3:
				g.Version = value
			default:
				// Not needed...
			}
		case egtRIdTableCount:
			g.symbols = newSymbolTable((<-curRec).asInt(), true)
			g.charSets = make([]charSet, (<-curRec).asInt())
			g.rules = newRuleTable((<-curRec).asInt())
			g.dfaStates = newDFAStateTable((<-curRec).asInt())
			g.lrStates = newLRStateTable((<-curRec).asInt())
			g.groups = newGroupTable((<-curRec).asInt(), true)
		case egtRIdCharSet:
			cs := new(rangedCharSet)
			g.charSets[(<-curRec).asInt()] = cs
			cs.plane = (<-curRec).asInt()

			rangeCnt := (<-curRec).asInt()
			<-curRec // reserved...

			cs.ranges = make(charRanges, rangeCnt)
			var i uint16
			for i = 0; i < rangeCnt; i++ {
				cs.ranges[i].start = (<-curRec).asInt()
				cs.ranges[i].end = (<-curRec).asInt()
			}
			cs.optimize()

		case egtRIdGroup:
			group := g.groups[(<-curRec).asInt()]
			group.Name = (<-curRec).asString()

			group.Container = g.symbols[(<-curRec).asInt()]
			group.Container.Group = group // link back

			group.Start = g.symbols[(<-curRec).asInt()]
			group.Start.Group = group // link back

			group.End = g.symbols[(<-curRec).asInt()]
			group.End.Group = group // link back

			group.AdvanceMode = advanceMode((<-curRec).asInt())
			group.EndingMode = endingMode((<-curRec).asInt())

			<-curRec // reserved
			nestingCnt := (<-curRec).asInt()
			group.Nested = newGroupTable(nestingCnt, false)
			var i uint16
			for i = 0; i < nestingCnt; i++ {
				group.Nested[i] = g.groups[(<-curRec).asInt()]
			}
		default:
			g.goldGrammar.loadRecord(recordId(recTyp), curRec)
		}
	}

	return g
}

func (g *egtGrammar) readTokens(rd io.Reader) <-chan *parserToken {
	result := make(chan *parserToken, 100)

	sr := newSourceReader(rd)

	groupStack := newStack()

	go func() {

		nestGroup := false
		for {
			read := g.readToken(sr)
			if read.Symbol.Kind == stEnd {
				// EOF always stops the loop. The caller method (parse) can flag a runaway group error.
				result <- read
				break
			}
			// Groups (comments, etc.)
			// The logic - to determine if a group should be nested - requires that the top
			// of the stack and the symbol's linked group need to be looked at. Both of these
			// can be unset. So, this section sets a boolean and avoids errors. We will use
			// this boolean in the logic chain below.
			if read.Symbol.Kind == stGroupStart {
				if groupStack.Len() == 0 {
					nestGroup = true
				} else {
					nestGroup = groupStack.Peek().(*parserToken).Symbol.Group.Nested.contains(read.Symbol.Group)
				}
			} else {
				nestGroup = false
			}

			// Logic chain
			if nestGroup {
				groupStack.Push(read)
			} else if groupStack.Len() == 0 {
				// The token is ready to be analyzed
				result <- read
			} else if groupStack.Peek().(*parserToken).Symbol.Group.End == read.Symbol {
				// End the current group
				pop := groupStack.Pop().(*parserToken)

				// Ending logic
				if pop.Symbol.Group.EndingMode == emClosed {
					pop.Text = pop.Text + read.Text
				}
				if groupStack.Len() == 0 {
					// We are out of the group. Return pop'd token which contains all the group text
					pop.Symbol = pop.Symbol.Group.Container
					result <- pop
				} else {
					// Append group text to parent
					groupStack.Peek().(*parserToken).Text += pop.Text
				}
			} else {
				// We are in a group, Append to the Token on the top of the stack.
				// Take into account the Token group mode
				top := groupStack.Peek().(*parserToken)
				if top.Symbol.Group.AdvanceMode == amToken {
					// Append all text
					top.Text += read.Text
				} else {
					// Append one character
					runes := []rune(read.Text)
					top.Text += string(runes[0])
					sr.UnreadAll(runes[1:])
				}
			}
		}
		close(result)
	}()

	return result
}
