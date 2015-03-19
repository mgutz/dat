package runner

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectDocRow(t *testing.T) {
	assert := assert.New(t)
	type Post struct {
		ID    int
		Title string
	}

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var person Person
	err := conn.
		SelectDoc("id", "name").
		From("people").
		Where("id = $1", 1).
		Load("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)

	assert.Equal(2, len(person.Posts))
	assert.Equal("Day 1", person.Posts[0].Title)
	assert.Equal("Day 2", person.Posts[1].Title)
}

func TestSelectDocNil(t *testing.T) {
	assert := assert.New(t)
	type Post struct {
		ID    int
		Title string
	}

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var person Person
	err := conn.
		SelectDoc("id", "name").
		From("people").
		Where("id = $1", 1000).
		Load("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		QueryStruct(&person)
	assert.Equal(sql.ErrNoRows, err)
}

func TestSelectDocRows(t *testing.T) {
	assert := assert.New(t)
	type Post struct {
		ID    int
		Title string
	}

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var people []*Person
	err := conn.
		SelectDoc("id", "name").
		SetIsInterpolated(true).
		From("people").
		Where("id in $1", []int{1, 2}).
		Load("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		QueryStructs(&people)

	assert.NoError(err)

	person := people[0]
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)

	assert.Equal(2, len(person.Posts))
	assert.Equal("Day 1", person.Posts[0].Title)
	assert.Equal("Day 2", person.Posts[1].Title)

	person = people[1]
	assert.Equal("John", person.Name)
	assert.Equal(2, person.ID)

	assert.Equal(2, len(person.Posts))
	assert.Equal("Apple", person.Posts[0].Title)
	assert.Equal("Orange", person.Posts[1].Title)
}

func TestSelectDocRowsNil(t *testing.T) {
	assert := assert.New(t)
	type Post struct {
		ID    int
		Title string
	}

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var people []*Person
	err := conn.
		SelectDoc("id", "name").
		SetIsInterpolated(true).
		From("people").
		Where("id in $1", []int{2000, 2001}).
		Load("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		QueryStructs(&people)
	assert.Equal(sql.ErrNoRows, err)
}

// Not efficient but it's doable
func TestSelectDoc(t *testing.T) {
	assert := assert.New(t)

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var person Person
	err := conn.
		SelectDoc("id", "name").
		From("people").
		Where("id = $1", 1).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)
}
