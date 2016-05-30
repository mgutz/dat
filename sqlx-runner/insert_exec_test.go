package runner

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/syreclabs/dat"
	"github.com/syreclabs/dat/common"
	"github.com/syreclabs/dat/postgres"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestInsertKeywordColumnName(t *testing.T) {
	// Insert a column whose name is reserved
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	res, err := s.
		InsertInto("people").
		Columns("name", "key").
		Values("Barack", "44").
		Exec()

	assert.NoError(t, err)
	assert.EqualValues(t, res.RowsAffected, 1)
}

func TestInsertDoubleDollarQuote(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	expected := common.RandomString(16)
	var str string
	err := s.
		InsertInto("people").
		Columns("name", "key").
		Values("test", expected).
		Returning("key").
		QueryScalar(&str)
	assert.NoError(t, err)
	assert.Equal(t, expected, str)

	// ensure the tag cannot be escaped by user
	oldDollarTag := postgres.GetPgDollarTag()
	expected = common.RandomString(1024) + "'" + oldDollarTag
	builder := s.
		InsertInto("people").
		Columns("name", "key").
		Values("test", expected).
		Returning("key")

	sql, _, _ := builder.SetIsInterpolated(true).Interpolate()
	assert.NotEqual(t, oldDollarTag, postgres.GetPgDollarTag())
	assert.True(t, strings.Contains(sql, postgres.GetPgDollarTag()))

	builder.QueryScalar(&str)
	assert.NoError(t, err)
	assert.Equal(t, expected, str)
}

func TestInsertDefault(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var str string
	err := s.
		InsertInto("people").Columns("name", "foo").
		Values("Barack", "fool").
		Returning("foo").
		QueryScalar(&str)
	assert.NoError(t, err)
	assert.Equal(t, str, "fool")

	dat.EnableInterpolation = true
	err = s.
		Update("people").
		Set("foo", dat.DEFAULT).
		Returning("foo").
		QueryScalar(&str)
	dat.EnableInterpolation = false
	assert.NoError(t, err)
	assert.Equal(t, str, "bar")
}

func TestInsertReal(t *testing.T) {
	// Insert by specifying values
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "obama0@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	// Insert by specifying a record (ptr to struct)
	person := Person{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama1@whitehouse.gov"

	err = s.
		InsertInto("people").
		Columns("name", "email").
		Record(&person).
		Returning("id", "created_at").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.NotEqual(t, person.CreatedAt, dat.NullTime{})

	// Insert by specifying a record (struct)
	person2 := Person{Name: "Barack"}
	person2.Email.Valid = true
	person2.Email.String = "obama2@whitehouse.gov"
	err = s.
		InsertInto("people").Columns("name", "email").
		Record(person2).
		Returning("id").
		QueryStruct(&person2)
	assert.NoError(t, err)
	assert.True(t, person2.ID > 0)
	assert.NotEqual(t, person.ID, person2.ID)
}

func TestInsertMultipleRecords(t *testing.T) {
	assert := assert.New(t)

	s := beginTxWithFixtures()
	defer s.AutoRollback()

	res, err := s.
		InsertInto("people").
		Columns("name", "email").
		Values("apple", "apple@fruits.local").
		Values("orange", "orange@fruits.local").
		Values("pear", "pear@fruits.local").
		Exec()
	assert.NoError(err)
	assert.EqualValues(res.RowsAffected, 3)

	person1 := Person{Name: "john_timr"}
	person2 := Person{Name: "jane_timr"}

	res, err = s.InsertInto("people").
		Columns("name", "email").
		Record(&person1).
		Record(&person2).
		Exec()
	assert.NoError(err)
	n := res.RowsAffected
	assert.EqualValues(n, 2)

	people := []Person{}
	err = s.
		Select("name").
		From("people").
		Where("name like $1", "%_timr").
		QueryStructs(&people)
	assert.NoError(err)
	assert.Equal(len(people), 2)

	n = 0
	for _, person := range people {
		if person.Name == "john_timr" {
			n++
		}
		if person.Name == "jane_timr" {
			n++
		}
	}
	assert.EqualValues(n, 2)
}

func TestInsertWhitelist(t *testing.T) {
	// Insert by specifying a record (struct)
	person2 := Person{Name: "Barack"}
	person2.Email.Valid = true
	person2.Email.String = "obama2@whitehouse.gov"
	var email sql.NullString
	var name string
	var id int64
	err := testDB.
		InsertInto("people").
		Whitelist("name").
		Record(person2).
		Returning("id", "name", "email").
		QueryScalar(&id, &name, &email)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.False(t, email.Valid)
	assert.Equal(t, name, "Barack")
}

func TestInsertBlacklist(t *testing.T) {
	// type Person struct {
	// 	ID        int64          `db:"id"`
	// 	Name      string         `db:"name"`
	// 	Foo       string         `db:"foo"`
	// 	Email     dat.NullString `db:"email"`
	// 	Key       dat.NullString `db:"key"`
	// 	Doc       dat.NullString `db:"doc"`
	// 	CreatedAt dat.NullTime   `db:"created_at"`
	// }

	// Insert by specifying a record (struct)
	person2 := Person{Name: "Barack"}
	person2.Email.Valid = true
	person2.Email.String = "obama2@whitehouse.gov"
	var email sql.NullString
	var name string
	var id int64

	err := testDB.
		InsertInto("people").
		Blacklist("id", "foo", "email", "key", "doc", "created_at").
		Record(person2).
		Returning("id", "name", "email").
		QueryScalar(&id, &name, &email)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.False(t, email.Valid)
	assert.Equal(t, name, "Barack")
}

func TestInsertBytes(t *testing.T) {
	b := []byte{0, 0, 0}
	var image []byte
	var id int32
	sql := `
		INSERT INTO people (name, image)
		VALUES ($1, $2)
		RETURNING id, image
	`
	dat.EnableInterpolation = true
	err := testDB.SQL(sql, "foo", b).QueryScalar(&id, &image)
	assert.NoError(t, err)
	assert.Exactly(t, b, image)
	dat.EnableInterpolation = false
}
