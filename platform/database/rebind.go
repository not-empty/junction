package database

import (
	"strconv"
	"strings"
)

// rebindDollar turns the ? placeholders into Postgres' $1, $2, ... so the same
// query runs on both drivers. A ? inside a string literal, a quoted identifier
// or a comment is kept, and ?? writes a literal ?, for the jsonb operators.
func rebindDollar(query string) string {
	if !strings.Contains(query, "?") {
		return query
	}

	var out strings.Builder
	out.Grow(len(query) + 8)

	n := 0

	for i := 0; i < len(query); i++ {
		c := query[i]

		switch {
		case c == '\'' || c == '"':
			end := closing(query, i+1, string(c))
			out.WriteString(query[i:end])
			i = end - 1
		case c == '-' && strings.HasPrefix(query[i:], "--"):
			end := closing(query, i+2, "\n")
			out.WriteString(query[i:end])
			i = end - 1
		case c == '/' && strings.HasPrefix(query[i:], "/*"):
			end := closing(query, i+2, "*/")
			out.WriteString(query[i:end])
			i = end - 1
		case c == '?' && strings.HasPrefix(query[i:], "??"):
			out.WriteByte('?')
			i++
		case c == '?':
			n++
			out.WriteByte('$')
			out.WriteString(strconv.Itoa(n))
		default:
			out.WriteByte(c)
		}
	}

	return out.String()
}

// closing returns the index right after the next delimiter from start, or the
// end of the query when it is never closed. An escaped quote (” or "") ends
// one run and opens the next, so its content is still skipped.
func closing(query string, start int, delimiter string) int {
	index := strings.Index(query[start:], delimiter)
	if index < 0 {
		return len(query)
	}

	return start + index + len(delimiter)
}
