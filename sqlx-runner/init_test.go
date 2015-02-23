package runner

import "github.com/mgutz/dat"

func init() {
	testConn = NewConnection(realDb(), "postgres")
	dat.SetVerbose(false)
	dat.Strict = true
}
