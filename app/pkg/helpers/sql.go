package helpers

import (
	"fmt"
	"strings"
)

// EnumerateSQLArgs - returns an SQL-ready string representation of a counter.
//
// As a special case, a counter <= 0 returns an empty string.
//
// Usage:
//
//	EnumerateSQLArgs(3, 0, func(i, c int) string { return "?" }) // ?, ?, ?
//	EnumerateSQLArgs(3, 0, func(i, c int) string { return fmt.Sprintf("$%d", i) }) // $1, $2, $3
func EnumerateSQLArgs(counter int, offset int, valueFunc func(index, counter int) string) string {
	var out string

	for i := 1; i <= counter; i++ {
		out += fmt.Sprintf(`%v, `, valueFunc(i, counter))
	}

	out = strings.TrimSpace(out)
	out = strings.TrimRight(out, ",")

	return out
}

// func A()
