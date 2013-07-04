package gold

type lrState struct {
	Index   int
	Actions lrActionTable
}

type lrStateTable []*lrState

func newLRStateTable(count int) lrStateTable {
	result := make(lrStateTable, count)
	for i := 0; i < count; i++ {
		result[i] = new(lrState)
	}
	return result
}

type lrAction struct {
	Symbol      *symbol
	Action      action
	TargetRule  *rule
	TargetState *lrState
}

type lrActionTable map[*symbol]*lrAction

func newLRActionTable(symbols symbolTable) lrActionTable {
	result := make(lrActionTable)
	for _, symbol := range symbols {
		result[symbol] = nil
	}

	return result
}

type action byte

const (
	actionShift  action = 1
	actionReduce action = 2
	actionGoto   action = 3
	actionAccept action = 4
)
