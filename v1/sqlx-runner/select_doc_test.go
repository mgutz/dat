package runner

import (
	"database/sql"
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
		As("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		From("people").
		Where("id = $1", 1).
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
		As("comments", `SELECT * FROM comments WHERE comments.id = posts.id`).
		From("posts").
		Where("user_id = people.id")

	err := conn.
		SelectDoc("id", "name").
		As("posts", posts).
		From("people").
		Where("id = $1", 1).
		SetIsInterpolated(true).
		QueryStruct(&obj)

	assert.NoError(err)
	assert.Equal("Mario", obj.AsString("name"))
	assert.Equal(1, obj.AsInt64("id"))

	assert.Equal("A very good day", obj.AsString("posts[0].comments[0].comment"))
	assert.Equal("Yum. Apple pie.", obj.AsString("posts[1].comments[0].comment"))
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
		As("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		From("people").
		Where("id = $1", 1000).
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
		As("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		From("people").
		Where("id in $1", []int{1, 2}).
		SetIsInterpolated(true).
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
		As("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		From("people").
		Where("id in $1", []int{2000, 2001}).
		SetIsInterpolated(true).
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

func TestSelectQueryEmbeddedJSON(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.AutoCommit()

	type PostEmbedded struct {
		ID    int    `db:"id"`
		State string `db:"state"`
		User  struct {
			ID int64
		}
	}

	type User struct {
		ID int64
	}

	type PostEmbedded2 struct {
		ID    int    `db:"id"`
		State string `db:"state"`
		User  *User
	}

	var post PostEmbedded

	err := s.SelectDoc("id", "state").
		HasOne("user", `select 42 as id`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post)

	assert.NoError(t, err)
	assert.Equal(t, 1, post.ID)
	assert.Equal(t, 42, post.User.ID)

	var post2 PostEmbedded2
	err = s.SelectDoc("id", "state").
		HasOne("user", `select 42 as id`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post2)

	assert.NoError(t, err)
	assert.Equal(t, 1, post2.ID)
	assert.Equal(t, 42, post2.User.ID)
}

func TestSelectDocHasOneNoRows(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.AutoCommit()

	type User struct {
		ID int64
	}

	type PostEmbedded struct {
		ID    int    `db:"id"`
		State string `db:"state"`
		User  *User
	}

	var post PostEmbedded
	err := s.SelectDoc("id", "state").
		HasOne("user", `select * from people where id = 1232345`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post)

	assert.NoError(t, err)
	assert.Equal(t, 1, post.ID)
	assert.Nil(t, post.User)
}
