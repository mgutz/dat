package dat

import (
	"bytes"
	"strings"
)

// PostgresDialect is the PostgeSQL dialect.
type PostgresDialect struct{}

// WriteStringLiteral is part of Dialect implementation.
//
// Postgres is much safer as of 9.1. Postgres does not allow any escape
// sequences by default. See http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE
func (pd *PostgresDialect) WriteStringLiteral(buf *bytes.Buffer, val string) {
	if val == "" {
		buf.WriteString("''")
		return
	}

	buf.WriteRune('\'')
	if strings.Contains(val, "'") {
		for _, char := range val {
			// apos
			if char == '\'' {
				buf.WriteString(`''`)
			} else if char == 0 {
				panic("postgres doesn't support NULL char in text, see http://stackoverflow.com/questions/1347646/postgres-error-on-insert-error-invalid-byte-sequence-for-encoding-utf8-0x0")
			} else {
				buf.WriteRune(char)
			}
		}
	} else {
		buf.WriteString(val)
	}
	buf.WriteRune('\'')
}

// WriteIdentifier is part of Dialect implementation.
func (pd *PostgresDialect) WriteIdentifier(buf *bytes.Buffer, ident string) {
	if ident == "" {
		panic("Identifier is empty string")
	}

	buf.WriteRune('"')
	buf.WriteString(ident)
	buf.WriteRune('"')
}
