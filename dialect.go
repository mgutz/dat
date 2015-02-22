package dat

import (
	"bytes"
)

// Dialect is the active SQLDialect.
var Dialect SQLDialect = &PostgresDialect{}

// SQLDialect represents a vendor specific SQL dialect.
type SQLDialect interface {
	// WriteStringLiteral writes a string literal.
	WriteStringLiteral(buf *bytes.Buffer, value string)
	// WriteIdentifier writes a quoted identifer such as a column or table.
	WriteIdentifier(buf *bytes.Buffer, column string)
}
