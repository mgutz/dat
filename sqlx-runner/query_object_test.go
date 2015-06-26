package runner

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"github.com/mgutz/jo/v1"
)

func TestQueryObject(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var people jo.Object
	err := s.
		Select("id", "name", "email").
		From("people").
		OrderBy("id ASC").
		QueryObject(&people)

	assert.NoError(t, err)
	assert.Equal(t, len(people.AsSlice(".")), 6)

	// Make sure that the Ids are set. It's possible (maybe?) that different DBs set ids differently so
	// don't assume they're 1 and 2.
	assert.True(t, people.MustInt64("[0].id") > 0)
	assert.True(t, people.MustInt64("[1].id") > people.MustInt64("[0].id"))

	mario, _ := people.At("[0]")
	john, _ := people.At("[1]")
	assert.Equal(t, mario.MustString("name"), "Mario")
	assert.Equal(t, mario.MustString("email"), "mario@acme.com")
	assert.Equal(t, john.MustString("name"), "John")
	assert.Equal(t, john.MustString("email"), "john@acme.com")
}
