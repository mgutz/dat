package dat

import (
	"fmt"
	"strconv"

	"github.com/mgutz/logxi"
)

var logger logxi.Logger

// Strict tells dat to raise errors
var Strict = false

// EnableInterpolation enables or disable interpolation
var EnableInterpolation = false

// maxLookup is the max lookup index for predefined lookup tables
const maxLookup = 100

// atoiTab holds "0" => 0, "1" .. 1 ... "99" -> 99 to avoid using strconv.Atoi()
var atoiTab = make(map[string]int, maxLookup)

// itoaTab holds [0]=="0", [1]=="1", ... [n]=="n". To avoid strconv.Itoa
var itoaTab = make([]string, maxLookup)

// placeholdersTab holds $0, $1 ... $n to avoid using  "$" + strconv.FormatInt()
var placeholderTab = make([]string, maxLookup)

// equalsPlaceholderTab " = $1"
var equalsPlaceholderTab = make([]string, maxLookup)

// inPlaceholderTab " IN $1"
var inPlaceholderTab = make([]string, maxLookup)

var identifierTab = make([]string, maxLookup)

func init() {
	// There are performance costs with using ordinal placeholders.
	// '?' placeholders are much more efficient but not eye friendly
	// when coding non-trivial queries.
	//
	// Most of the cost is incurred when converting between integers and
	// strings. These lookup tables hardcode the values for up to maxLookup args
	// which should cover most queries. Anything over maxLookup defaults to
	// using strconv.FormatInt.
	for i := 0; i < maxLookup; i++ {
		placeholderTab[i] = fmt.Sprintf("$%d", i)
		inPlaceholderTab[i] = fmt.Sprintf(" IN $%d", i)
		equalsPlaceholderTab[i] = fmt.Sprintf(" = $%d", i)
		atoiTab[strconv.Itoa(i)] = i
		itoaTab[i] = strconv.Itoa(i)
		identifierTab[i] = fmt.Sprintf("dat%d", i)
	}

	logger = logxi.New("dat")
}
