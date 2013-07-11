package gold

import "bytes"

type GrammarInformation struct {
	Name    string
	Version string
	Author  string
	About   string
}

func (gi GrammarInformation) getInformation() GrammarInformation {
	return gi
}

type goldGrammar struct {
	GrammarInformation

	initialDFAState uint16
	initialLRState  uint16

	symbols   symbolTable
	rules     ruleTable
	dfaStates dfaStateTable
	lrStates  lrStateTable
	charSets  []charSet

	errorSymbol *symbol
	endSymbol   *symbol
}

func (g *goldGrammar) getInitialDfaState() *dfaState {
	return g.dfaStates[g.initialDFAState]
}

func (g *goldGrammar) getInitialLRState() *lrState {
	return g.lrStates[g.initialLRState]
}

func (g *goldGrammar) readToken(r *sourceReader) *parserToken {

	dfa := g.getInitialDfaState()

	tText := new(bytes.Buffer)

	result := new(parserToken)
	result.Text = ""
	result.Symbol = g.errorSymbol
	result.Position = r.Position
	for {
		if !r.Next() {
			if r.Rune == 0 && tText.Len() == 0 {
				result.Text = ""
				result.Symbol = g.endSymbol
			}
			return result
		}

		nextState, ok := dfa.TransitionVector(r.Rune)
		if ok {
			tText.WriteRune(r.Rune)

			dfa = nextState
			if dfa.AcceptSymbol != nil {
				result.Text = string(tText.Bytes())
				result.Symbol = dfa.AcceptSymbol
			}
		} else {
			if result.Symbol == g.errorSymbol {
				tText.WriteRune(r.Rune)
			} else {
				r.UnreadLast()
			}
			break
		}
	}
	result.Text = string(tText.Bytes())

	return result
}

type recordId byte

const (
	rIdDFAStates recordId = 68 // D
	rIdInitial   recordId = 73 // I
	rIdLRTables  recordId = 76 // L
	rIdRules     recordId = 82 // R
	rIdSymbol    recordId = 83 // S
)

func (g *goldGrammar) loadRecord(recId recordId, record recordEntry) {
	switch recId {
	case rIdInitial:
		g.initialDFAState = (<-record).asInt()
		g.initialLRState = (<-record).asInt()
	case rIdLRTables:
		idx := (<-record).asInt()
		(<-record) // reserved...

		actions := record.readTillEnd()
		actnCount := len(actions) / 4

		lrstate := g.lrStates[idx]
		lrstate.Index = idx

		lrstate.Actions = make(lrActionTable)
		for i := 0; i < actnCount; i++ {
			symb := g.symbols[actions[4*i+0].asInt()]

			actn := new(lrAction)
			lrstate.Actions[symb] = actn
			actn.Symbol = symb
			actn.Action = action(actions[4*i+1].asInt())
			targetIdx := actions[4*i+2].asInt()
			actn.TargetRule = nil
			actn.TargetState = nil

			switch actn.Action {
			case actionShift, actionGoto:
				actn.TargetState = g.lrStates[targetIdx]
			case actionReduce:
				actn.TargetRule = g.rules[targetIdx]
			}

		}
	case rIdRules:
		r := g.rules[(<-record).asInt()]
		r.NonTerminal = g.symbols[(<-record).asInt()]
		<-record // Field No 3 is reserved...
		symbols := record.readTillEnd()

		r.Symbols = newSymbolTable(uint16(len(symbols)), false)

		for idx, entry := range symbols {
			r.Symbols[idx] = g.symbols[entry.asInt()]
		}
		g.rules[r.Index] = r
	case rIdSymbol:
		idx := (<-record).asInt()
		g.symbols[idx].Index = idx
		g.symbols[idx].Name = (<-record).asString()
		g.symbols[idx].Kind = symbolType((<-record).asInt())

		switch g.symbols[idx].Kind {
		case stEnd:
			g.endSymbol = g.symbols[idx]
		case stError:
			g.errorSymbol = g.symbols[idx]
		}
	case rIdDFAStates:
		idx := (<-record).asInt()
		state := g.dfaStates[idx]

		if (<-record).asBool() {
			state.AcceptSymbol = g.symbols[(<-record).asInt()]
		} else {
			(<-record) // not used...
			state.AcceptSymbol = nil
		}
		(<-record) // Reserved...

		edges := record.readTillEnd()
		edgeCount := len(edges) / 3
		dfaedges := make([]dfaEdge, edgeCount)

		for i := 0; i < len(edges)/3; i++ {
			dfaedges[i].CharSet = g.charSets[edges[3*i+0].asInt()]
			dfaedges[i].Target = g.dfaStates[edges[3*i+1].asInt()]
		}

		state.TransitionVector = newTransitionVector(dfaedges)
	default:
		record.readTillEnd() // skip this record...
	}
}
