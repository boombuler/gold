package gold

type charSet string

type charSetTable []charSet

func newCharSetTable(count int) charSetTable {
	result := make(charSetTable, count)
	for i := 0; i < count; i++ {
		result[i] = ""
	}
	return result
}
