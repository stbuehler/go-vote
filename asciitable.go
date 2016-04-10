package main

import (
	"bytes"
	"fmt"
)

func AsciiIntTable(entries [][]int, tableName string, labels []string) string {
	var buf bytes.Buffer
	colSize := 0
	for i, _ := range entries {
		if colSize < len(labels[i]) {
			colSize = len(labels[i])
		}
	}
	colSize = colSize + len(tableName) + 5
	colFormat := fmt.Sprintf("%%%ds |", colSize)
	entryFormat := fmt.Sprintf("%%%dd |", colSize)

	fmt.Fprint(&buf, "|")
	fmt.Fprintf(&buf, colFormat, tableName)
	for i, _ := range entries {
		fmt.Fprintf(&buf, colFormat, fmt.Sprintf("%s[*,%s]", tableName, labels[i]))
	}
	fmt.Fprint(&buf, "\n")

	for i, line := range entries {
		fmt.Fprint(&buf, "|")
		fmt.Fprintf(&buf, colFormat, fmt.Sprintf("%s[%s,*]", tableName, labels[i]))
		for _, wins := range line {
			fmt.Fprintf(&buf, entryFormat, wins)
		}
		fmt.Fprint(&buf, "\n")
	}
	return buf.String()
}
