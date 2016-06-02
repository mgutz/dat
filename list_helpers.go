package dat

import (
	"gopkg.in/mgutz/dat.v1/common"
)

var bufPool = common.NewBufferPool()

func writeIdentifiers(buf common.BufferWriter, columns []string, join string) {
	for i, column := range columns {
		if i > 0 {
			buf.WriteString(join)
		}
		Dialect.WriteIdentifier(buf, column)
	}
}

func writeIdentifier(buf common.BufferWriter, name string) {
	Dialect.WriteIdentifier(buf, name)
}

func buildPlaceholders(buf common.BufferWriter, start, length int) {
	// Build the placeholder like "($1,$2,$3)"
	buf.WriteRune('(')
	for i := start; i < start+length; i++ {
		if i > start {
			buf.WriteRune(',')
		}
		writePlaceholder(buf, i)
	}
	buf.WriteRune(')')
}

// joinPlaceholders returns $1, $2 ... , $n
func writePlaceholders(buf common.BufferWriter, length int, join string, offset int) {
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteString(join)
		}
		writePlaceholder(buf, i+offset)
	}
}
