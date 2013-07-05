package gold

import (
	"bufio"
	"io"
	"unicode/utf16"
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

type cgtRecEntry struct {
	typ   cgtRecEntryTyp
	value interface{}
}

func (re cgtRecEntry) asString() string {
	if re.typ == etString {
		return re.value.(string)
	}
	return ""
}

func (re cgtRecEntry) asByte() byte {
	if re.typ == etByte {
		return re.value.(byte)
	}
	return 0
}

func (re cgtRecEntry) asInt() int {
	if re.typ == etInt16 {
		return int(re.value.(int16))
	}
	return 0
}

func (re cgtRecEntry) asBool() bool {
	if re.typ == etBool {
		return re.value.(bool)
	}
	return false
}

type cgtRecord []*cgtRecEntry

type grammarerror string

func (ge grammarerror) Error() string {
	return string(ge)
}

var cgtFormatError grammarerror = "Invalid CGT File"

const (
	cgtHeader = "GOLD Parser Tables/v1.0"
)

type cgtRecEntryTyp byte

const (
	etEmpty cgtRecEntryTyp = iota
	etBool
	etByte
	etInt16
	etString
)

type cgtRecordId byte

const (
	recordIdParameters  cgtRecordId = 80 //P
	recordIdTableCounts cgtRecordId = 84 //T
	recordIdInitial     cgtRecordId = 73 //I
	recordIdSymbols     cgtRecordId = 83 //S
	recordIdCharSets    cgtRecordId = 67 //C
	recordIdRules       cgtRecordId = 82 //R
	recordIdDFAStates   cgtRecordId = 68 //D
	recordIdLRTables    cgtRecordId = 76 //L
	recordIdComment     cgtRecordId = 33 //!
)

func readUInt16(r *bufio.Reader) (uint16, error) {
	b1, err1 := r.ReadByte()
	b2, err2 := r.ReadByte()

	if err1 != nil {
		return 0, err1
	} else if err2 != nil {
		return 0, err2
	}

	return uint16(b2)<<8 | uint16(b1), nil
}

func readString(r *bufio.Reader) (string, error) {
	result := make([]uint16, 0)
	for {
		v, err := readUInt16(r)
		if err != nil {
			return "", err
		}
		if v == 0 {
			break
		}
		result = append(result, v)
	}

	return string(utf16.Decode(result)), nil
}

func readRecordEntry(r *bufio.Reader) (*cgtRecEntry, error) {
	eTypS, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch string(eTypS) {
	case "E":
		return &cgtRecEntry{typ: etEmpty, value: nil}, nil
	case "B":
		val, err := r.ReadByte()
		return &cgtRecEntry{typ: etBool, value: val == 1}, err
	case "b":
		val, err := r.ReadByte()
		return &cgtRecEntry{typ: etByte, value: val}, err
	case "I":
		val, err := readUInt16(r)
		return &cgtRecEntry{typ: etInt16, value: int16(val)}, err
	case "S":
		val, err := readString(r)
		return &cgtRecEntry{typ: etString, value: val}, err
	}

	return nil, cgtFormatError
}

func readRecord(r *bufio.Reader) (cgtRecord, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if string(typ) == "M" {
		entryCnt, err := readUInt16(r)
		if err != nil {
			return nil, err
		} else if entryCnt < 1 {
			return nil, cgtFormatError
		}
		result := make(cgtRecord, entryCnt)

		var i uint16
		for i = 0; i < entryCnt; i++ {
			entr, err := readRecordEntry(r)
			if err != nil {
				return nil, err
			}
			result[i] = entr
		}
		return result, nil
	}

	return nil, cgtFormatError
}

func loadCGTGramar(r io.Reader) *cgtGramar {
	defer func() {
		recover()
	}()

	rd := bufio.NewReader(r)

	if head, err := readString(rd); head != cgtHeader || err != nil {
		return nil
	}
	result := new(cgtGramar)

	for {
		curRec, _ := readRecord(rd)
		if curRec == nil {
			break
		}
		// process record...
		recTyp := cgtRecordId(curRec[0].asByte())

		switch recTyp {
		case recordIdParameters:
			result.Name = curRec[1].asString()
			result.Version = curRec[2].asString()
			result.Author = curRec[3].asString()
			result.About = curRec[4].asString()
			result.CaseSensitive = curRec[5].asBool()
			result.startSymbol = curRec[6].asInt()

		case recordIdTableCounts:
			result.symbols = newSymbolTable(curRec[1].asInt(), true)
			result.charSets = newCharSetTable(curRec[2].asInt())
			result.rules = newRuleTable(curRec[3].asInt())
			result.dfaStates = newDFAStateTable(curRec[4].asInt())
			result.lrStates = newLRStateTable(curRec[5].asInt())

		case recordIdSymbols:
			idx := curRec[1].asInt()
			result.symbols[idx].Index = idx
			result.symbols[idx].Name = curRec[2].asString()
			result.symbols[idx].Kind = symbolType(curRec[3].asInt())

			switch result.symbols[idx].Kind {
			case stEnd:
				result.endSymbol = result.symbols[idx]
			case stError:
				result.errorSymbol = result.symbols[idx]
			}

		case recordIdCharSets:
			idx := curRec[1].asInt()
			result.charSets[idx] = charSet(curRec[2].asString())

		case recordIdRules:
			r := result.rules[curRec[1].asInt()]
			r.NonTerminal = result.symbols[curRec[2].asInt()]
			// Field No 3 is reserved...
			r.Symbols = newSymbolTable(len(curRec)-4, false)

			for idx, entry := range curRec[4:] {
				r.Symbols[idx] = result.symbols[entry.asInt()]
			}
			result.rules[r.Index] = r

		case recordIdInitial:
			result.initialDFAState = curRec[1].asInt()
			result.initialLRState = curRec[2].asInt()

		case recordIdDFAStates:
			idx := curRec[1].asInt()
			state := result.dfaStates[idx]

			if curRec[2].asBool() {
				state.AcceptSymbol = result.symbols[curRec[3].asInt()]
			} else {
				state.AcceptSymbol = nil
			}

			edges := curRec[5:]
			edgeCount := len(edges) / 3
			dfaedges := make([]dfaEdge, edgeCount)

			for i := 0; i < len(edges)/3; i++ {
				dfaedges[i].CharSet = result.charSets[edges[3*i+0].asInt()]
				dfaedges[i].Target = result.dfaStates[edges[3*i+1].asInt()]
			}

			state.TransitionVector = newTransitionVector(dfaedges)
		case recordIdLRTables:
			idx := curRec[1].asInt()

			actions := curRec[3:]
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
