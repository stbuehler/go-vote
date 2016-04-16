package types

import (
	"encoding/json"
)

func Alloc2DimInt(cols, rows int) [][]int {
	cells := make([]int, cols*rows)
	table := make([][]int, rows)
	for r := 0; r < rows; r++ {
		table[r], cells = cells[:cols], cells[cols:]
	}
	return table
}

func JsonMustEncode(v interface{}) []byte {
	if s, err := json.Marshal(v); nil != err {
		panic(err)
	} else {
		return s
	}
}

func JsonMustEncodeString(v interface{}) string {
	return string(JsonMustEncode(v))
}
