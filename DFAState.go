package gold

import "strings"

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
		chars := strings.NewReader(string(edge.CharSet))

		for {
			r, _, err := chars.ReadRune()
			if err == nil {
				result[r] = edge.Target
			} else {
				break
			}
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
