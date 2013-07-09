package gold

type simpleCharSet string

type simpleCharSetTable []simpleCharSet

func newSimpleCharSetTable(count uint16) simpleCharSetTable {
	result := make(simpleCharSetTable, count)
	var i uint16
	for i = 0; i < count; i++ {
		result[i] = ""
	}
	return result
}

type charRange struct {
	start uint16
	end   uint16
}

type charSet struct {
	plane  uint16
	ranges []charRange
}

func newCharSetTable(count uint16) []*charSet {
	result := make([]*charSet, count)
	var i uint16
	for i = 0; i < count; i++ {
		result[i] = new(charSet)
	}
	return result
}

func (cs charSet) contains(r rune) bool {
	plane := uint16((uint64(r) >> 16) & 0xFF)
	if plane == cs.plane {
		chr := uint16(uint64(r) & 0xFFFF)

		for _, r := range cs.ranges {
			if (r.start <= chr) && (r.end >= chr) {
				return true
			}
		}
	}
	return false
}
