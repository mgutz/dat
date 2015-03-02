package runner

import (
	"os"

	"github.com/mgutz/logxi/v1"
)

var logger log.Logger

func init() {
	logger = log.New(os.Stdout, "dat:sqlx")
}
