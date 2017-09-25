//mysql alike table
package cliTable

import (
	"fmt"
	"io"
	"strings"
)

// DataTable contains data to print
type DataTable [][]interface{}

// Sprint to string: fmt.Sprint
func Sprint(dataTable DataTable) string {
	var ret = ""
	var colMaxWidth []int
	var formatElem []string

	// Calc colWidth
	colMaxWidth = make([]int, len(dataTable[0]))

	for _, dataRow := range dataTable {
		for i, c := range dataRow {
			slen := len(c.(string))
			if colMaxWidth[i] < slen {
				colMaxWidth[i] = slen
			}
		}
	}

	for _, v := range colMaxWidth {
		formatElem = append(formatElem, fmt.Sprintf("  %%%ds", v))
	}

	formatString := strings.Join(formatElem, "  |")

	ret += fmt.Sprintf("\r\n")

	for i, dataRow := range dataTable {
		str := ""
		str = fmt.Sprintf(formatString, dataRow...)
		if i == 0 {
			str += "\r\n" + fmt.Sprintf(strings.Repeat("-", len(str)+2))
		}
		ret += fmt.Sprintf("%s\r\n", str)
	}
	ret += fmt.Sprintf("\r\n")

	return ret
}

// Fprint print table to writer
func Fprint(w io.Writer, dataTable DataTable) (int, error) {
	return fmt.Fprint(w, Sprint(dataTable))
}

// Print print using fmt.Print
func Print(dataTable DataTable) (int, error) {
	return fmt.Print(Sprint(dataTable))
}
