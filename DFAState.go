package gold

type dfaEdge struct {
	CharSet charSet
	Target  *dfaState
}

type dfaState struct {
	AcceptSymbol     *symbol
	TransitionVector transitionVector
}

type transitionVector map[rune]*dfaState

func newTransitionVector(edges []dfaEdge) transitionVector {
	result := make(transitionVector)

	for _, edge := range edges {
		for _, r := range edge.CharSet {
			result[r] = edge.Target
		}
	}
	return result
}

type dfaStateTable []*dfaState

func newDFAStateTable(count int) dfaStateTable {
	result := make(dfaStateTable, count)
	for i := 0; i < count; i++ {
		result[i] = new(dfaState)
	}
	return result
}
