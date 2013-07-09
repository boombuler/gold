package gold

import (
	"bufio"
)

const (
	cgtHeader = "GOLD Parser Tables/v1.0"
)

type cgtRecordId byte

const (
	cgtRIdParameters  cgtRecordId = 80 // P
	cgtRIdTableCounts cgtRecordId = 84 // T
	cgtRIdInitial     cgtRecordId = 73 // I
	cgtRIdSymbols     cgtRecordId = 83 // S
	cgtRIdCharSets    cgtRecordId = 67 // C
	cgtRIdRules       cgtRecordId = 82 // R
	cgtRIdDFAStates   cgtRecordId = 68 // D
	cgtRIdLRTables    cgtRecordId = 76 // L
)

type cgtGrammar struct {
	Infos           GrammarInformation
	CaseSensitive   bool
	initialDFAState uint16
	initialLRState  uint16

	symbols   symbolTable
	charSets  simpleCharSetTable
	rules     ruleTable
	dfaStates dfaStateTable
	lrStates  lrStateTable

	errorSymbol *symbol
	endSymbol   *symbol
}

func (g cgtGrammar) getInformation() GrammarInformation {
	return g.Infos
}

func (g cgtGrammar) getInitialDfaState() *dfaState {
	return g.dfaStates[g.initialDFAState]
}

func (g cgtGrammar) getInitialLRState() *lrState {
	return g.lrStates[g.initialLRState]
}

type cgtRecEntryTyp byte

func loadCGTGrammar(rd *bufio.Reader) *cgtGrammar {
	result := new(cgtGrammar)

	records := readRecords(rd)

	for curRec := range records {
		// process record...
		recTyp := cgtRecordId((<-curRec).asByte())

		switch recTyp {
		case cgtRIdParameters:
			result.Infos.Name = (<-curRec).asString()
			result.Infos.Version = (<-curRec).asString()
			result.Infos.Author = (<-curRec).asString()
			result.Infos.About = (<-curRec).asString()
			result.CaseSensitive = (<-curRec).asBool()
			(<-curRec).asInt()

		case cgtRIdTableCounts:
			result.symbols = newSymbolTable((<-curRec).asInt(), true)
			result.charSets = newSimpleCharSetTable((<-curRec).asInt())
			result.rules = newRuleTable((<-curRec).asInt())
			result.dfaStates = newDFAStateTable((<-curRec).asInt())
			result.lrStates = newLRStateTable((<-curRec).asInt())

		case cgtRIdSymbols:
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

		case cgtRIdCharSets:
			idx := (<-curRec).asInt()
			result.charSets[idx] = simpleCharSet((<-curRec).asString())

		case cgtRIdRules:
			r := result.rules[(<-curRec).asInt()]
			r.NonTerminal = result.symbols[(<-curRec).asInt()]
			<-curRec // Field No 3 is reserved...

			symbols := curRec.readTillEnd()

			r.Symbols = newSymbolTable(uint16(len(symbols)), false)

			for idx, entry := range symbols {
				r.Symbols[idx] = result.symbols[entry.asInt()]
			}
			result.rules[r.Index] = r

		case cgtRIdInitial:
			result.initialDFAState = (<-curRec).asInt()
			result.initialLRState = (<-curRec).asInt()

		case cgtRIdDFAStates:
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
			dfaedges := make([]simpleDfaEdge, edgeCount)

			for i := 0; i < len(edges)/3; i++ {
				dfaedges[i].CharSet = result.charSets[edges[3*i+0].asInt()]
				dfaedges[i].Target = result.dfaStates[edges[3*i+1].asInt()]
			}

			state.TransitionVector = newSimpleTransitionVector(dfaedges)
		case cgtRIdLRTables:
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
		default:
			curRec.readTillEnd()
		}
	}
	return result
}
