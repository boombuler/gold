package gold

import (
	"sort"
	"strings"
)

type charSet interface {
	contains(r rune) bool
}

type fixedCharSet string

func (f fixedCharSet) contains(r rune) bool {
	return strings.ContainsRune(string(f), r)
}

type charRange struct {
	start uint16
	end   uint16
}

type charRanges []charRange

type rangedCharSet struct {
	plane  uint16
	ranges charRanges
}

func (s charRanges) Len() int {
	return len(s)
}

func (s charRanges) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s charRanges) Less(i, j int) bool {
	return s[i].start < s[j].start
}

func newRangedCharSetTable(count uint16) []*charSet {
	result := make([]*charSet, count)
	var i uint16
	for i = 0; i < count; i++ {
		result[i] = new(charSet)
	}
	return result
}

func (cs *rangedCharSet) contains(r rune) bool {
	plane := uint16((uint64(r) >> 16) & 0xFF)
	if plane == cs.plane {
		chr := uint16(uint64(r) & 0xFFFF)

		for _, r := range cs.ranges {
			if chr < r.start {
				return false
			}
			if chr <= r.end {
				return true
			}

		}
	}
	return false
}

func (cs *rangedCharSet) optimize() {
	sort.Sort(cs.ranges)
}
