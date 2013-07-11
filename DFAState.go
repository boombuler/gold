package gold

type dfaEdge struct {
	CharSet charSet
	Target  *dfaState
}

type dfaTransition func(r rune) (*dfaState, bool)

type dfaState struct {
	AcceptSymbol     *symbol
	TransitionVector dfaTransition
}

func newTransitionVector(edges []dfaEdge) dfaTransition {
	return func(r rune) (*dfaState, bool) {
		for _, edge := range edges {
			if edge.CharSet.contains(r) {
				return edge.Target, true
			}
		}
		return nil, false
	}
}

type dfaStateTable []*dfaState

func newDFAStateTable(count uint16) dfaStateTable {
	result := make(dfaStateTable, count)
	var i uint16
	for i = 0; i < count; i++ {
		result[i] = new(dfaState)
	}
	return result
}
