package dat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mgutz/dat/v1/common"
)

// NameMapping is the routine to use when mapping column names to struct properties
var NameMapping = camelCaseToSnakeCase

func camelCaseToSnakeCase(name string) string {
	var buf bytes.Buffer

	firstTime := true
	for _, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if firstTime {
				firstTime = false
			} else {
				buf.WriteRune('_')
			}
			chr -= ('A' - 'a')
		}
		buf.WriteRune(chr)
	}

	return buf.String()
}

func camelCaseToSnakeCaseID(name string) string {
	// handle the common ID idiom
	if name == "ID" {
		return "id"
	}
	var buf bytes.Buffer

	lenName := len(name)
	writeID := false
	if lenName > 2 {
		writeID = name[lenName-2:lenName-1] == "I" && name[lenName-1:lenName] == "D"
	}

	firstTime := true
	for i, chr := range name {
		if writeID && i == lenName-2 {
			buf.WriteString("_id")
			break
		}
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if firstTime {
				firstTime = false
			} else {
				buf.WriteRune('_')
			}
			chr -= ('A' - 'a')
		}
		buf.WriteRune(chr)
	}

	return buf.String()
}

func writeInt64(buf common.BufferWriter, n int64) {
	if 0 <= n && n < maxLookup {
		buf.WriteString(itoaTab[int(n)])
	} else {
		buf.WriteString(strconv.FormatInt(n, 10))
	}
}

func writeUint64(buf common.BufferWriter, n uint64) {
	if n <= maxLookup {
		buf.WriteString(itoaTab[int(n)])
	} else {
		buf.WriteString(strconv.FormatUint(n, 10))
	}
}

func writePlaceholder(buf common.BufferWriter, pos int) {
	if pos < maxLookup {
		buf.WriteString(placeholderTab[pos])
	} else {
		buf.WriteRune('$')
		buf.WriteString(strconv.Itoa(pos))
	}
}

func writePlaceholder64(buf common.BufferWriter, pos int64) {
	if pos < maxLookup {
		buf.WriteString(placeholderTab[pos])
	} else {
		buf.WriteRune('$')
		buf.WriteString(strconv.FormatInt(pos, 10))
	}
}

// SQLMapFromReader creates a SQL map from an io.Reader.
//
// This string
//
//		`
//		--@selectUsers
//		SELECT * FROM users;
//
//		--@selectAccounts
//		SELECT * FROM accounts;
//		`
//
//		returns map[string]string{
//			"selectUsers": "SELECT * FROM users;",
//			"selectACcounts": "SELECT * FROM accounts;",
//		}
func SQLMapFromReader(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	var buf bytes.Buffer
	var key string
	var text string
	result := map[string]string{}
	collect := false
	for scanner.Scan() {
		text = scanner.Text()
		if strings.HasPrefix(text, "--@") {
			if key != "" {
				result[key] = buf.String()
			}
			key = text[3:]
			collect = true
			buf.Reset()
			continue
		}
		if collect {
			buf.WriteString(text)
			buf.WriteRune('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if key != "" && buf.Len() > 0 {
		result[key] = buf.String()
	}

	if collect {
		return result, nil
	}
	return nil, nil
}

// SQLMapFromFile loads a file containing special markers and loads
// the SQL statements into a map.
func SQLMapFromFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return SQLMapFromReader(file)
}

// SQLMapFromString creates a map of strings from s.
func SQLMapFromString(s string) (map[string]string, error) {
	buf := bytes.NewBufferString(s)
	return SQLMapFromReader(buf)
}

var goRe = regexp.MustCompile(`(?m)^GO$`)

// SQLSliceFromString converts a multiline string marked by `^GO$`
// into a slice of SQL statements.
//
// This string
//
//		SELECT *
//		FROM users;
//		GO
//		SELECT *
//		FROM accounts;
//
//		returns []string{"SELECT *\nFROM users;", "SELECT *\nFROM accounts"}
func SQLSliceFromString(s string) ([]string, error) {
	sli := goRe.Split(s, -1)
	return sli, nil
}

// SQLSliceFromFile reads a file's text then passes it to
// SQLSliceFromString.
func SQLSliceFromFile(filename string) ([]string, error) {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return SQLSliceFromString(string(text))
}

// getIdentifier
func getIdentifier(i *int) string {
	*i++
	idx := *i
	if idx < maxLookup {
		return identifierTab[idx]
	}
	return fmt.Sprintf("dat%d", idx)
}
