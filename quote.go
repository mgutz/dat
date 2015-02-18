package dat

import (
	"bytes"
)

// Quoter is the quoter to use for quoting text; use Mysql quoting by default
var Quoter = PostgresQuoter{}

// Interface for driver-swappable quoting
type quoter interface {
	WriteQuotedColumn(column string, sql *bytes.Buffer)
}

// MysqlQuoter implements Mysql-specific quoting
type MysqlQuoter struct{}

func (q MysqlQuoter) WriteQuotedColumn(column string, sql *bytes.Buffer) {
	sql.WriteRune('`')
	sql.WriteString(column)
	sql.WriteRune('`')
}

// PostgresQuoter implements Postgres-specific quoting
type PostgresQuoter struct{}

func (q PostgresQuoter) WriteQuotedColumn(column string, sql *bytes.Buffer) {
	sql.WriteRune('"')
	sql.WriteString(column)
	sql.WriteRune('"')
}
