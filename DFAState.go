package gold

type simpleDfaEdge struct {
	CharSet simpleCharSet
	Target  *dfaState
}

type dfaEdge struct {
	CharSet *charSet
	Target  *dfaState
}

type dfaTransition func(r rune) (*dfaState, bool)

type dfaState struct {
	AcceptSymbol     *symbol
	TransitionVector dfaTransition
}

func newSimpleTransitionVector(edges []simpleDfaEdge) dfaTransition {
	vector := make(map[rune]*dfaState)

	for _, edge := range edges {
		for _, r := range edge.CharSet {
			vector[r] = edge.Target
		}
	}

	return func(r rune) (*dfaState, bool) {
		state, ok := vector[r]
		return state, ok
	}
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
