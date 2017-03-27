package kvs

import "github.com/mgutz/logxi"

var logger logxi.Logger

func init() {
	logger = logxi.New("dat.cache")
}
