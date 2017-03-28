package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestVersion(t *testing.T) {
	// require at least 9.3+ for testing
	assert.True(t, testDB.Version > 90300)
}

func TestSkipVersionLookup(t *testing.T) {
	dat.SkipVersionDetection = true
	versionTestDB := NewDB(sqlDB, "postgres")
	assert.True(t, versionTestDB.Version == 0)
	dat.SkipVersionDetection = false
}
