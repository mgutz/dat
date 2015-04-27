package runner

import (
	"gopkg.in/mgutz/dat.v1/v1"
	"gopkg.in/mgutz/dat.v1/v1/postgres"
	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func init() {
	dat.Dialect = postgres.New()
	logger = log.New("dat:sqlx")
}
