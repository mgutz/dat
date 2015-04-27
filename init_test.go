package dat

import (
	"bytes"
	"fmt"

	"gopkg.in/mgutz/dat.v1/postgres"
)

func init() {
	Dialect = postgres.New()
}

func quoteSQL(sqlFmt string, cols ...string) string {
	args := make([]interface{}, len(cols))

	for i := range cols {
		args[i] = quoteColumn(cols[i])
	}

	return fmt.Sprintf(sqlFmt, args...)
}

func quoteColumn(column string) string {
	var buffer bytes.Buffer
	Dialect.WriteIdentifier(&buffer, column)
	return buffer.String()
}
