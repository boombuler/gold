package gold

import (
	"io"
	"unicode/utf16"
)

func readByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	cnt, err := r.Read(buf)
	if cnt != 1 {
		return 0, io.ErrUnexpectedEOF
	} else if err != nil {
		return 0, err
	}

	return buf[0], nil
}

func readUInt16(r io.Reader) (uint16, error) {
	b1, err1 := readByte(r)
	b2, err2 := readByte(r)

	if err1 != nil {
		return 0, err1
	} else if err2 != nil {
		return 0, err2
	}

	return uint16(b2)<<8 | uint16(b1), nil
}

func readString(r io.Reader) (string, error) {
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
