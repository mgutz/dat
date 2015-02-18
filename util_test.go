package dat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnakeCase(t *testing.T) {
	assert.Equal(t, camelCaseToSnakeCase("ID"), "id")
	assert.Equal(t, camelCaseToSnakeCase("a"), "a")
	assert.Equal(t, camelCaseToSnakeCase("SomeId"), "some_id")
	assert.Equal(t, camelCaseToSnakeCase("SomeID"), "some_i_d")
}
