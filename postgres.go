package dat

import "bytes"

// PostgresDialect is the PostgeSQL dialect.
type PostgresDialect struct{}

// WriteStringLiteral is part of Dialect implementation.
//
// Postgres is much safer as of 9.1. Postgres does not allow any escape
// sequences by default. See
// http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE
func (pd *PostgresDialect) WriteStringLiteral(buf *bytes.Buffer, val string) {
	for _, r := range val {
		if r == '\\' {
			buf.WriteRune('E')
			break
		}
	}

	buf.WriteRune('\'')
	for _, char := range val {
		if char == '\\' {
			// slash
			buf.WriteString(`\\`)
		} else if char == '\'' {
			// apos
			buf.WriteString(`\'`)
		} else if char == 0 {
			panic("postgres doesn't support NULL char in text, see http://stackoverflow.com/questions/1347646/postgres-error-on-insert-error-invalid-byte-sequence-for-encoding-utf8-0x0")
		} else {
			buf.WriteRune(char)
		}
	}
	buf.WriteRune('\'')
}

// WriteIdentifier is part of Dialect implementation.
func (pd *PostgresDialect) WriteIdentifier(buf *bytes.Buffer, ident string) {
	buf.WriteRune('"')
	buf.WriteString(ident)
	buf.WriteRune('"')
}
