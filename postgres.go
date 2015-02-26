package dat

import (
	"bytes"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// pgDollarTag is the double dollar tag for escaping strings.
var pgDollarTag string
var pgDollarMutex sync.Mutex

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	randomizePgDollarTag()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var lenLetters = len(letters)

func randomizePgDollarTag() {
	pgDollarMutex.Lock()
	defer pgDollarMutex.Unlock()
	var buf bytes.Buffer
	buf.WriteRune('$')
	buf.WriteString(RandomString(3))
	buf.WriteRune('$')
	pgDollarTag = buf.String()
}

// GetPgDollarTag returns the current Postgres string dollar quoting tag.
func GetPgDollarTag() string {
	return pgDollarTag
}

// RandomString returns a random string of length n.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(lenLetters)]
	}
	return string(b)
}

// PostgresDialect is the PostgeSQL dialect.
type PostgresDialect struct{}

// WriteStringLiteral is part of Dialect implementation.
//
// Postgres 9.1+ does not allow any escape
// sequences by default. See http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE
// In short, all backslashes are treated literally not as escape sequences.
func (pd *PostgresDialect) WriteStringLiteral(buf *bytes.Buffer, val string) {
	if val == "" {
		buf.WriteString("''")
		return
	}

	hasTag := true

	// don't use double dollar quote strings unless the string is long enough
	if len(val) > 64 {
		// if pgDollarTag unique tag is in string, try to create a new one (only once though)
		hasTag = strings.Contains(val, pgDollarTag)
		if hasTag {
			randomizePgDollarTag()
			hasTag = strings.Contains(val, pgDollarTag)
		}
	}

	if hasTag {
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
	} else {
		buf.WriteString(pgDollarTag)
		buf.WriteString(val)
		buf.WriteString(pgDollarTag)
	}
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
