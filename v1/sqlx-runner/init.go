package runner

import (
	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/postgres"
	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func init() {
	dat.Dialect = postgres.New()
	logger = log.New("dat:sqlx")
}
