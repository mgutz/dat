package dat

import (
	"fmt"
	"strconv"
)

// Events is the event receiver.
//var Events = NewLogEventReceiver("[dat ]")
var Events = &NullEventReceiver{}

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
	for i := 0; i < maxLookup; i++ {
		placeholderTab[i] = fmt.Sprintf("$%d", i)
		inPlaceholderTab[i] = fmt.Sprintf(" IN $%d", i)
		equalsPlaceholderTab[i] = fmt.Sprintf(" = $%d", i)
		atoiTab[strconv.Itoa(i)] = i
		itoaTab[i] = strconv.Itoa(i)
	}
}
