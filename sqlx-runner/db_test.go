package runner

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestVersion(t *testing.T) {
	// require at least 9.3+ for testing
	assert.True(t, testDB.Version > 90300)
}
