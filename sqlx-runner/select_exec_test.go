package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestSelectQueryStructs(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var people []Person
	err := s.
		Select("id", "name", "email").
		From("people").
		OrderBy("id ASC").
		QueryStructs(&people)

	assert.NoError(t, err)
	assert.Equal(t, len(people), 2)

	// Make sure that the Ids are set. It's possible (maybe?) that different DBs set ids differently so
	// don't assume they're 1 and 2.
	assert.True(t, people[0].ID > 0)
	assert.True(t, people[1].ID > people[0].ID)

	assert.Equal(t, people[0].Name, "Jonathan")
	assert.True(t, people[0].Email.Valid)
	assert.Equal(t, people[0].Email.String, "jonathan@acme.com")
	assert.Equal(t, people[1].Name, "Dmitri")
	assert.True(t, people[1].Email.Valid)
	assert.Equal(t, people[1].Email.String, "zavorotni@jadius.com")

	// TODO: test map
}

func TestSelectQueryStruct(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	// Found:
	var person Person
	err := s.
		Select("id", "name", "email").
		From("people").
		Where("email = $1", "jonathan@acme.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.Equal(t, person.Name, "Jonathan")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "jonathan@acme.com")

	// Not found:
	var person2 Person
	err = s.
		Select("id", "name", "email").
		From("people").Where("email = $1", "dontexist@acme.com").
		QueryStruct(&person2)
	assert.Contains(t, err.Error(), "no rows")
}

func TestSelectBySqlQueryStructs(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var people []*Person
	dat.EnableInterpolation = true
	err := s.
		SQL("SELECT name FROM people WHERE email IN $1", []string{"jonathan@acme.com"}).
		QueryStructs(&people)
	dat.EnableInterpolation = false

	assert.NoError(t, err)
	assert.Equal(t, len(people), 1)
	assert.Equal(t, people[0].Name, "Jonathan")
	assert.Equal(t, people[0].ID, 0)              // not set
	assert.Equal(t, people[0].Email.Valid, false) // not set
	assert.Equal(t, people[0].Email.String, "")   // not set
}

func TestSelectQueryScalar(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var name string
	err := s.
		Select("name").
		From("people").
		Where("email = 'jonathan@acme.com'").
		QueryScalar(&name)

	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var id int64
	err = s.Select("id").From("people").Limit(1).QueryScalar(&id)

	assert.NoError(t, err)
	assert.True(t, id > 0)
}

func TestSelectQuerySlice(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var names []string
	err := s.Select("name").From("people").QuerySlice(&names)

	assert.NoError(t, err)
	assert.Equal(t, len(names), 2)
	assert.Equal(t, names, []string{"Jonathan", "Dmitri"})

	var ids []int64
	err = s.Select("id").From("people").Limit(1).QuerySlice(&ids)

	assert.NoError(t, err)
	assert.Equal(t, len(ids), 1)
	assert.Equal(t, ids, []int64{1})
}

func TestScalar(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var name string
	err := s.Select("name").From("people").Where("email = 'jonathan@acme.com'").QueryScalar(&name)
	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var count int64
	err = s.Select("COUNT(*)").From("people").QueryScalar(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)
}

func TestSelectExpr(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var name string
	scope := dat.Expr("email = $1", "jonathan@acme.com")
	err := s.Select("name").From("people").Where(scope).QueryScalar(&name)
	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var count int64
	err = s.Select("COUNT(*)").From("people").QueryScalar(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)
}

func TestSelectScope(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var id int
	err := s.
		InsertInto("people").
		Columns("name").
		Values("mgutz").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	var postID int
	err = s.
		InsertInto("posts").
		Columns("title", "state", "user_id").
		Values("my post", "published", id).
		Returning("id").
		QueryScalar(&postID)
	assert.NoError(t, err)
	assert.True(t, postID > 0)

	publishedByUser := dat.NewScope(
		`
		INNER JOIN people P on (P.id = :TABLE.user_id)
		WHERE
			P.name = :user AND
			:TABLE.state = 'published' AND
			:TABLE.deleted_at IS NULL
		`,
		dat.M{"user": ""},
	)
	var posts []*Post
	err = s.
		Select("posts.*").
		From("posts").
		Scope(publishedByUser, dat.M{"user": "mgutz"}).
		QueryStructs(&posts)
	assert.NoError(t, err)
	assert.Equal(t, posts[0].Title, "my post")
}

// Series of tests that test mapping struct fields to columns
