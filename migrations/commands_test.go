package migrations

import (
	"fmt"
	"testing"
)

func TestConsole(t *testing.T) {
	err := Console()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
