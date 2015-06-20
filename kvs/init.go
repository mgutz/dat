package kvs

import (
	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func init() {
	logger = log.New("dat.cache")
}
