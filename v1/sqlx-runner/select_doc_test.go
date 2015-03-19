package runner

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/mgutz/dat/v1"
	"github.com/mgutz/jo/v1"
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
		Embed("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)

	assert.Equal(2, len(person.Posts))
	assert.Equal("Day 1", person.Posts[0].Title)
	assert.Equal("Day 2", person.Posts[1].Title)
}

func TestSelectDocNested(t *testing.T) {
	assert := assert.New(t)

	var obj jo.Object

	posts := dat.SelectDoc("id", "title").
		From("posts").
		Where("user_id = people.id").
		Embed("comments", `SELECT * FROM comments WHERE comments.id = posts.id`)

	err := conn.
		SelectDoc("id", "name").
		SetIsInterpolated(true).
		From("people").
		Where("id = $1", 1).
		Embed("posts", posts).
		QueryStruct(&obj)

	assert.NoError(err)
	assert.Equal("Mario", obj.AsString("name"))
	assert.Equal(1, obj.AsInt64("id"))

	fmt.Println("DBG: obj", obj.Stringify())

	// assert.Equal(2, len(person.Posts))
	// assert.Equal("Day 1", person.Posts[0].Title)
	// assert.Equal("Day 2", person.Posts[1].Title)
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
		Embed("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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
		Embed("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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
		Embed("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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
