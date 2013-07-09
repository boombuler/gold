package gold

type lrState struct {
	Index   uint16
	Actions lrActionTable
}

type lrStateTable []*lrState

func newLRStateTable(count uint16) lrStateTable {
	result := make(lrStateTable, count)
	var i uint16
	for i = 0; i < count; i++ {
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

type action byte

const (
	actionShift  action = 1
	actionReduce action = 2
	actionGoto   action = 3
	actionAccept action = 4
)
