package gold

import (
	"bufio"
	"io"
)

type cgtGramar struct {
	Name            string
	Version         string
	Author          string
	About           string
	CaseSensitive   bool
	startSymbol     int
	initialDFAState int
	initialLRState  int

	symbols   symbolTable
	charSets  charSetTable
	rules     ruleTable
	dfaStates dfaStateTable
	lrStates  lrStateTable

	errorSymbol *symbol
	endSymbol   *symbol
}

func (g cgtGramar) getInitialDfaState() *dfaState {
	return g.dfaStates[g.initialDFAState]
}

func (g cgtGramar) getInitialLRState() *lrState {
	return g.lrStates[g.initialLRState]
}

const (
	cgtHeader = "GOLD Parser Tables/v1.0"
)

type cgtRecEntryTyp byte

func loadCGTGramar(r io.Reader) *cgtGramar {
	defer func() {
		recover()
	}()

	rd := bufio.NewReader(r)

	if head, err := readString(rd); head != cgtHeader || err != nil {
		return nil
	}
	result := new(cgtGramar)

	records := readRecords(rd)

	for curRec := range records {
		// process record...
		recTyp := cgtRecordId((<-curRec).asByte())

		switch recTyp {
		case recordIdParameters:
			result.Name = (<-curRec).asString()
			result.Version = (<-curRec).asString()
			result.Author = (<-curRec).asString()
			result.About = (<-curRec).asString()
			result.CaseSensitive = (<-curRec).asBool()
			result.startSymbol = (<-curRec).asInt()

		case recordIdTableCounts:
			result.symbols = newSymbolTable((<-curRec).asInt(), true)
			result.charSets = newCharSetTable((<-curRec).asInt())
			result.rules = newRuleTable((<-curRec).asInt())
			result.dfaStates = newDFAStateTable((<-curRec).asInt())
			result.lrStates = newLRStateTable((<-curRec).asInt())

		case recordIdSymbols:
			idx := (<-curRec).asInt()
			result.symbols[idx].Index = idx
			result.symbols[idx].Name = (<-curRec).asString()
			result.symbols[idx].Kind = symbolType((<-curRec).asInt())

			switch result.symbols[idx].Kind {
			case stEnd:
				result.endSymbol = result.symbols[idx]
			case stError:
				result.errorSymbol = result.symbols[idx]
			}

		case recordIdCharSets:
			idx := (<-curRec).asInt()
			result.charSets[idx] = charSet((<-curRec).asString())

		case recordIdRules:
			r := result.rules[(<-curRec).asInt()]
			r.NonTerminal = result.symbols[(<-curRec).asInt()]
			<-curRec // Field No 3 is reserved...

			symbols := curRec.readTillEnd()

			r.Symbols = newSymbolTable(len(symbols), false)

			for idx, entry := range symbols {
				r.Symbols[idx] = result.symbols[entry.asInt()]
			}
			result.rules[r.Index] = r

		case recordIdInitial:
			result.initialDFAState = (<-curRec).asInt()
			result.initialLRState = (<-curRec).asInt()

		case recordIdDFAStates:
			idx := (<-curRec).asInt()
			state := result.dfaStates[idx]

			if (<-curRec).asBool() {
				state.AcceptSymbol = result.symbols[(<-curRec).asInt()]
			} else {
				(<-curRec) // not used...
				state.AcceptSymbol = nil
			}
			(<-curRec) // Reserved...

			edges := curRec.readTillEnd()
			edgeCount := len(edges) / 3
			dfaedges := make([]dfaEdge, edgeCount)

			for i := 0; i < len(edges)/3; i++ {
				dfaedges[i].CharSet = result.charSets[edges[3*i+0].asInt()]
				dfaedges[i].Target = result.dfaStates[edges[3*i+1].asInt()]
			}

			state.TransitionVector = newTransitionVector(dfaedges)
		case recordIdLRTables:
			idx := (<-curRec).asInt()
			(<-curRec) // reserved...

			actions := curRec.readTillEnd()
			actnCount := len(actions) / 4

			lrstate := result.lrStates[idx]
			lrstate.Index = idx

			lrstate.Actions = make(lrActionTable)
			for i := 0; i < actnCount; i++ {
				symb := result.symbols[actions[4*i+0].asInt()]

				actn := new(lrAction)
				lrstate.Actions[symb] = actn
				actn.Symbol = symb
				actn.Action = action(actions[4*i+1].asInt())
				targetIdx := actions[4*i+2].asInt()
				actn.TargetRule = nil
				actn.TargetState = nil

				switch actn.Action {
				case actionShift, actionGoto:
					actn.TargetState = result.lrStates[targetIdx]
				case actionReduce:
					actn.TargetRule = result.rules[targetIdx]
				}

			}
		}
	}
	return result
}
