package runner

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/mgutz/jo/v1"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
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
	err := testDB.
		SelectDoc("id", "name").
		Many("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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

	var person Person

	posts := dat.SelectDoc("id", "title").
		Many("comments", `SELECT * FROM comments WHERE comments.id = posts.id`).
		From("posts").
		Where("user_id = people.id")

	err := testDB.
		SelectDoc("id", "name").
		Many("posts", posts).
		From("people").
		Where("id = $1", 1).
		SetIsInterpolated(true).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(int64(1), person.ID)

	assert.Equal("A very good day", person.Posts[0].Comments[0].Comment)
	assert.Equal("Yum. Apple pie.", person.Posts[1].Comments[0].Comment)
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
	err := testDB.
		SelectDoc("id", "name").
		Many("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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
	err := testDB.
		SelectDoc("id", "name").
		Many("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
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
	err := testDB.
		SelectDoc("id", "name").
		Many("posts", `SELECT id, title FROM posts WHERE user_id = people.id`).
		From("people").
		Where("id in $1", []int{2000, 2001}).
		SetIsInterpolated(true).
		QueryStructs(&people)
	assert.Equal(sql.ErrNoRows, err)
}

func TestSelectDoc(t *testing.T) {
	assert := assert.New(t)

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var person Person
	err := testDB.
		SelectDoc("id", "name").
		From("people").
		Where("id = $1", 1).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)
}

func TestSelectDocDistinctOn(t *testing.T) {
	assert := assert.New(t)

	type Person struct {
		ID    int
		Name  string
		Posts []*Post
	}

	var person Person
	err := testDB.
		SelectDoc("id", "name").
		DistinctOn("id").
		From("people").
		Where("id = $1", 1).
		QueryStruct(&person)

	assert.NoError(err)
	assert.Equal("Mario", person.Name)
	assert.Equal(1, person.ID)
}

func TestSelectQueryEmbeddedJSON(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

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
		One("user", `select 42 as id`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post)

	assert.NoError(t, err)
	assert.Equal(t, 1, post.ID)
	assert.EqualValues(t, 42, post.User.ID)

	var post2 PostEmbedded2
	err = s.SelectDoc("id", "state").
		One("user", `select 42 as id`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post2)

	assert.NoError(t, err)
	assert.Equal(t, 1, post2.ID)
	assert.EqualValues(t, 42, post2.User.ID)
}

func TestSelectDocOneNoRows(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

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
		One("user", `select * from people where id = 1232345`).
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post)

	assert.NoError(t, err)
	assert.Equal(t, 1, post.ID)
	assert.Nil(t, post.User)
}

func TestSelectDocDate(t *testing.T) {
	var comments []*Comment

	err := testDB.SelectDoc("id", "comment", `created_at as "CreatedAt"`).
		From("comments").
		QueryStructs(&comments)

	assert.NoError(t, err)
	assert.True(t, comments[0].CreatedAt.Valid)
	assert.True(t, comments[1].CreatedAt.Valid)
}

func TestSelectDocBytes(t *testing.T) {
	b, err := testDB.SelectDoc("id", "comment").
		From("comments").
		OrderBy("id").
		QueryJSON()

	assert.NoError(t, err)

	var comments jo.Object
	err = json.Unmarshal(b, &comments)
	assert.NoError(t, err)

	assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
	assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
}

func TestSelectDocObject(t *testing.T) {
	var comments jo.Object
	err := testDB.SelectDoc("id", "comment").
		From("comments").
		OrderBy("id").
		QueryObject(&comments)

	assert.NoError(t, err)
	assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
	assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
}

func TestSelectDocForObject(t *testing.T) {
	var comments jo.Object
	err := testDB.SelectDoc("id", "comment").
		From("comments").
		OrderBy("id").
		For("UPDATE").
		QueryObject(&comments)

	assert.NoError(t, err)
	assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
	assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
}
