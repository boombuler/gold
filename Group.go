package gold

type group struct {
	Name        string
	Container   *symbol
	Start       *symbol
	End         *symbol
	AdvanceMode advanceMode
	EndingMode  endingMode
	Nested      groupTable
}

type advanceMode uint16

const (
	amToken     advanceMode = 0 // The group will advance a token at a time.
	amCharacter advanceMode = 1 // The group will advance by just one character at a time.
)

type endingMode uint16

const (
	emOpen   endingMode = 0 // The ending symbol will be left on the input queue.
	emClosed endingMode = 1 // The ending symbol will be consumed.
)

type groupTable []*group

func newGroupTable(count uint16, createGroups bool) groupTable {
	result := make(groupTable, count)
	if createGroups {
		var i uint16
		for i = 0; i < count; i++ {
			result[i] = new(group)
		}
	}
	return result
}

func (gt groupTable) contains(g *group) bool {
	for _, gc := range gt {
		if gc == g {
			return true
		}
	}
	return false
}
