package kvs

import logxi "github.com/mgutz/logxi/v1"

var logger logxi.Logger

func init() {
	logger = logxi.New("dat.cache")
}
