package gold

import (
	"bufio"
	"unicode/utf16"
)

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

	return &cgtRecEntry{typ: etEmpty, value: nil}, nil
}

type recordEntry (<-chan *cgtRecEntry)

func (re recordEntry) readTillEnd() []*cgtRecEntry {
	result := make([]*cgtRecEntry, 0)
	for x := range re {
		result = append(result, x)
	}
	return result
}

func readRecords(r *bufio.Reader) <-chan recordEntry {
	result := make(chan recordEntry)

	go func() {

		for {
			typ, err := r.ReadByte()
			if err != nil {
				close(result)
				return
			}

			if string(typ) == "M" {
				record := make(chan *cgtRecEntry)
				result <- record

				entryCnt, err := readUInt16(r)
				if err != nil {
					close(record)
					close(result)
					return
				} else if entryCnt < 1 {
					close(record)
					return
				}

				var i uint16
				for i = 0; i < entryCnt; i++ {
					entr, err := readRecordEntry(r)
					if err != nil {
						close(record)
						close(result)
						return
					}
					record <- entr
				}
				close(record)
			} else {
				close(result)
				return
			}
		}

	}()

	return result
}
