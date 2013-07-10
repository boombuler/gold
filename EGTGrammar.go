package gold

import (
	"bufio"
)

const (
	egtHeader = "GOLD Parser Tables/v5.0"
)

type egtGrammar struct {
	Infos GrammarInformation

	initialDFAState uint16
	initialLRState  uint16

	charSets  []*charSet
	symbols   symbolTable
	rules     ruleTable
	dfaStates dfaStateTable
	lrStates  lrStateTable
	groups    groupTable

	errorSymbol *symbol
	endSymbol   *symbol
}

type egtRecordId byte

const (
	egtRIdProperty   egtRecordId = 112 // p
	egtRIdTableCount egtRecordId = 116 // t
	egtRIdCharSet    egtRecordId = 99  // c
	egtRIdSymbol     egtRecordId = 83  // S
	egtRIdRules      egtRecordId = 82  // R
	egtRIdInitial    egtRecordId = 73  // I
	egtRIdDFAStates  egtRecordId = 68  // D
	egtRIdLRTables   egtRecordId = 76  // L
	egtRIdGroup      egtRecordId = 103 // g
)

func (g egtGrammar) getInformation() GrammarInformation {
	return g.Infos
}

func (g egtGrammar) getInitialDfaState() *dfaState {
	return g.dfaStates[g.initialDFAState]
}

func (g egtGrammar) getInitialLRState() *lrState {
	return g.lrStates[g.initialLRState]
}

func loadEGTGrammar(rd *bufio.Reader) *egtGrammar {
	result := new(egtGrammar)

	records := readRecords(rd)

	for curRec := range records {
		// process record...
		recTyp := egtRecordId((<-curRec).asByte())

		switch recTyp {
		case egtRIdProperty:
			idx := (<-curRec).asInt()
			<-curRec // skip name...
			value := (<-curRec).asString()

			switch idx {
			case 0:
				result.Infos.Name = value
			case 1:
				result.Infos.Version = value
			case 2:
				result.Infos.Author = value
			case 3:
				result.Infos.Version = value
			default:
				// Not needed...
			}
		case egtRIdTableCount:
			result.symbols = newSymbolTable((<-curRec).asInt(), true)
			result.charSets = newCharSetTable((<-curRec).asInt())
			result.rules = newRuleTable((<-curRec).asInt())
			result.dfaStates = newDFAStateTable((<-curRec).asInt())
			result.lrStates = newLRStateTable((<-curRec).asInt())
			result.groups = newGroupTable((<-curRec).asInt(), true)
		case egtRIdCharSet:
			charSet := result.charSets[(<-curRec).asInt()]
			charSet.plane = (<-curRec).asInt()

			rangeCnt := (<-curRec).asInt()
			<-curRec // reserved...

			charSet.ranges = make([]charRange, rangeCnt)
			var i uint16
			for i = 0; i < rangeCnt; i++ {
				charSet.ranges[i].start = (<-curRec).asInt()
				charSet.ranges[i].end = (<-curRec).asInt()
			}

		case egtRIdSymbol:
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

		case egtRIdRules:
			r := result.rules[(<-curRec).asInt()]
			r.NonTerminal = result.symbols[(<-curRec).asInt()]
			<-curRec // Field No 3 is reserved...
			symbols := curRec.readTillEnd()

			r.Symbols = newSymbolTable(uint16(len(symbols)), false)

			for idx, entry := range symbols {
				r.Symbols[idx] = result.symbols[entry.asInt()]
			}
			result.rules[r.Index] = r
		case egtRIdInitial:
			result.initialDFAState = (<-curRec).asInt()
			result.initialLRState = (<-curRec).asInt()

		case egtRIdDFAStates:

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

		case egtRIdLRTables:

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

		case egtRIdGroup:
			group := result.groups[(<-curRec).asInt()]
			group.Name = (<-curRec).asString()

			group.Container = result.symbols[(<-curRec).asInt()]
			group.Container.Group = group // link back

			group.Start = result.symbols[(<-curRec).asInt()]
			group.Start.Group = group // link back

			group.End = result.symbols[(<-curRec).asInt()]
			group.End.Group = group // link back

			group.AdvanceMode = advanceMode((<-curRec).asInt())
			group.EndingMode = endingMode((<-curRec).asInt())

			<-curRec // reserved
			nestingCnt := (<-curRec).asInt()
			group.Nested = newGroupTable(nestingCnt, false)
			var i uint16
			for i = 0; i < nestingCnt; i++ {
				group.Nested[i] = result.groups[(<-curRec).asInt()]
			}
		default:
			curRec.readTillEnd()
		}
	}

	return result
}
