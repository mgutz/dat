package runner

import (
	"github.com/mgutz/logxi/v1"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/postgres"
)

var logger log.Logger

func init() {
	dat.Dialect = postgres.New()
	logger = log.New("dat:sqlx")
}
