package postgres

import (
	"bytes"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgutz/dat.v1/common"
)

// pgDollarTag is the double dollar tag for escaping strings.
var pgDollarTag string
var pgDollarMutex sync.Mutex

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	randomizePgDollarTag()
}

func randomizePgDollarTag() {
	pgDollarMutex.Lock()
	defer pgDollarMutex.Unlock()
	var buf bytes.Buffer
	buf.WriteRune('$')
	buf.WriteString(common.RandomString(3))
	buf.WriteRune('$')
	pgDollarTag = buf.String()
}

// GetPgDollarTag returns the current Postgres string dollar quoting tag.
func GetPgDollarTag() string {
	return pgDollarTag
}

// Postgres is the PostgeSQL dialect.
type Postgres struct{}

// New returns a new Postgres dialect.
func New() *Postgres {
	return &Postgres{}
}

// WriteStringLiteral writes an escaped string. No escape characters
// are allowed.
//
// Postgres 9.1+ does not allow any escape
// sequences by default. See http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE
// In short, all backslashes are treated literally not as escape sequences.
func (pd *Postgres) WriteStringLiteral(buf common.BufferWriter, val string) {
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

// WriteIdentifier writes escaped identifier.
func (pd *Postgres) WriteIdentifier(buf common.BufferWriter, ident string) {
	if ident == "" {
		panic("Identifier is empty string")
	}
	if ident == "*" {
		buf.WriteString(ident)
		return
	}

	buf.WriteRune('"')
	buf.WriteString(ident)
	buf.WriteRune('"')
}

// WriteFormattedTime formats t into a format postgres understands.
// Taken with gratitude from pq: https://github.com/lib/pq/blob/b269bd035a727d6c1081f76e7a239a1b00674c40/encode.go#L403
func (pd *Postgres) WriteFormattedTime(buf common.BufferWriter, t time.Time) {
	buf.WriteRune('\'')
	defer buf.WriteRune('\'')
	// XXX: This doesn't currently deal with infinity values

	// Need to send dates before 0001 A.D. with " BC" suffix, instead of the
	// minus sign preferred by Go.
	// Beware, "0000" in ISO is "1 BC", "-0001" is "2 BC" and so on
	bc := false
	if t.Year() <= 0 {
		// flip year sign, and add 1, e.g: "0" will be "1", and "-10" will be "11"
		t = t.AddDate((-t.Year())*2+1, 0, 0)
		bc = true
	}
	buf.WriteString(t.Format(time.RFC3339Nano))

	_, offset := t.Zone()
	offset = offset % 60
	if offset != 0 {
		// RFC3339Nano already printed the minus sign
		if offset < 0 {
			offset = -offset
		}

		buf.WriteRune(':')
		if offset < 10 {
			buf.WriteRune('0')
		}
		buf.WriteString(strconv.FormatInt(int64(offset), 10))
	}

	if bc {
		buf.WriteString(" BC")
	}
}
