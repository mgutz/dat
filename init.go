package dat

import (
	"fmt"
	"strconv"
)

// Events is the event receiver.
var Events EventReceiver

// Strict tells dat to raise errors
var Strict = false

// Whether to enable interpolation
var EnableInterpolation = false

// SetVerbose sets the verbosity of logging which defaults to none
func SetVerbose(verbose bool) {
	if verbose {
		Events = NewLogEventReceiver("[dat ]")
	} else {
		Events = &NullEventReceiver{}
	}
}

// SetStrict sets strict value
func SetStrict(strict bool) {
	Strict = strict
}

// maxLookup is the max lookup index for predefined lookup tables
const maxLookup = 100

// atoiTab holds "0" => 0, "1" .. 1 ... "99" -> 99 to avoid using strconv.Atoi()
var atoiTab = make(map[string]int, maxLookup)

// itoaTab holds 0 => "0", 1 => "1" ... n => "n" to avod strconv.Itoa
var itoaTab = make(map[int]string, maxLookup)

// placeholdersTab holds $0, $1 ... $n to avoid using  "$" + strconv.FormatInt()
var placeholderTab = make([]string, maxLookup)

// equalsPlaceholderTab " = $1"
var equalsPlaceholderTab = make([]string, maxLookup)

// inPlaceholderTab " IN $1"
var inPlaceholderTab = make([]string, maxLookup)

func init() {
	SetVerbose(false)

	// There is a performance cost related to using ordinal placeholders.
	// Using '?' placeholders is much more efficient but not eye friendly
	// when coding non-trivial queries.
	//
	// Most of the cost is incurred when converting between integers and
	// strings. These lookup tables hardcode the values for up to 100 args
	// which should cover most queries. Anything over maxLookup defaults to
	// using strconv.FormatInt.
	for i := 0; i < maxLookup; i++ {
		placeholderTab[i] = fmt.Sprintf("$%d", i)
		inPlaceholderTab[i] = fmt.Sprintf(" IN $%d", i)
		equalsPlaceholderTab[i] = fmt.Sprintf(" = $%d", i)
		atoiTab[strconv.Itoa(i)] = i
		itoaTab[i] = strconv.Itoa(i)
	}
}
