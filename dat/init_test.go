package dat

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/mgutz/dat/postgres"
	"gopkg.in/stretchr/testify.v1/assert"
)

type varargs []interface{}

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
	writeIdentifier(&buffer, column)
	return buffer.String()
}

func checkSliceEqual(t *testing.T, expected, actual []interface{}, msgAndArgs ...interface{}) bool {
	if fmt.Sprintf("%v", expected) != fmt.Sprintf("%v", actual) {
		return assert.Fail(t, fmt.Sprintf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual), msgAndArgs...)
	}

	return true
}
