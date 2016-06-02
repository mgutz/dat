package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestUpdateKeywordColumnName(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	// Insert a user with a key
	res, err := s.
		InsertInto("people").
		Columns("name", "email", "key").
		Values("Benjamin", "ben@whitehouse.gov", "6").
		Exec()
	assert.NoError(t, err)

	// Update the key
	res, err = s.Update("people").Set("key", "6-revoked").Where(dat.Eq{"key": "6"}).Exec()
	assert.NoError(t, err)

	// Assert our record was updated (and only our record)
	assert.EqualValues(t, res.RowsAffected, 1)

	var person Person
	err = s.Select("*").From("people").Where(dat.Eq{"email": "ben@whitehouse.gov"}).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Name, "Benjamin")
	assert.Equal(t, person.Key.String, "6-revoked")
}

func TestUpdateReal(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	// Insert a George
	err := s.InsertInto("people").Columns("name", "email").
		Values("George", "george@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)

	// Rename our George to Barack
	_, err = s.Update("people").SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).Where("id = $1", id).Exec()

	assert.NoError(t, err)

	var person Person
	err = s.Select("*").From("people").Where("id = $1", id).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}

func TestUpdateReturningStar(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	// Insert a George
	var insertPerson Person
	err := s.InsertInto("people").Columns("name", "email").
		Values("George", "george@whitehouse.gov").
		Returning("*").
		QueryStruct(&insertPerson)

	assert.NoError(t, err)
	assert.NotEmpty(t, insertPerson.ID)
	assert.Equal(t, insertPerson.Name, "George")
	assert.Equal(t, insertPerson.Email.Valid, true)
	assert.Equal(t, insertPerson.Email.String, "george@whitehouse.gov")

	var updatePerson Person
	err = s.Update("people").
		Set("name", "Barack").
		Set("email", "barack@whitehouse.gov").
		Where("id = $1", insertPerson.ID).
		Returning("*").
		QueryStruct(&updatePerson)
	assert.NoError(t, err)
	assert.Equal(t, insertPerson.ID, updatePerson.ID)
	assert.Equal(t, updatePerson.Name, "Barack")
	assert.Equal(t, updatePerson.Email.Valid, true)
	assert.Equal(t, updatePerson.Email.String, "barack@whitehouse.gov")
}

func TestUpdateWhitelist(t *testing.T) {
	installFixtures()

	// Insert by specifying a record (struct)
	p := Person{Name: "Barack"}
	p.Foo = "bar"
	var foo string
	var name string
	var id int64
	err := testDB.
		InsertInto("people").
		Whitelist("name", "foo").
		Record(p).
		Returning("id", "name", "foo").
		QueryScalar(&id, &name, &foo)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, name, "Barack")
	assert.Equal(t, foo, "bar")

	p2 := Person{Name: "oy"}
	p2.Foo = "bah"
	var name2 string
	var foo2 string
	err = testDB.
		Update("people").
		SetWhitelist(p2, "foo").
		Where("id = $1", id).
		Returning("name", "foo").
		QueryScalar(&name2, &foo2)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, name2, "Barack")
	assert.Equal(t, foo2, "bah")

}

func TestUpdateBlacklist(t *testing.T) {
	installFixtures()

	// Insert by specifying a record (struct)
	p := Person{Name: "Barack"}
	p.Foo = "bar"
	var foo string
	var name string
	var id int64
	err := testDB.
		InsertInto("people").
		Whitelist("name", "foo").
		Record(p).
		Returning("id", "name", "foo").
		QueryScalar(&id, &name, &foo)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, name, "Barack")
	assert.Equal(t, foo, "bar")

	p2 := Person{Name: "oy"}
	p2.Foo = "bah"
	var name2 string
	var foo2 string
	err = testDB.
		Update("people").
		SetBlacklist(p2, "id", "name", "email", "key", "doc", "created_at").
		Where("id = $1", id).
		Returning("name", "foo").
		QueryScalar(&name2, &foo2)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, name2, "Barack")
	assert.Equal(t, foo2, "bah")
}

func TestUpdateScope(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	// Insert a George
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Scope", "scope@foo.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)

	scope := dat.NewScope("WHERE id = :id", dat.M{"id": 1000})

	// Rename our George to Barack
	_, err = s.
		Update("people").
		SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).
		ScopeMap(scope, dat.M{"id": id}).
		Exec()

	assert.NoError(t, err)

	var person Person
	err = s.Select("*").From("people").Where("id = $1", id).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}

func TestUpdateScopeFunc(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	// Insert a George
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Scope", "scope@foo.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)

	scope := `WHERE id = $1`

	// Rename our George to Barack
	_, err = s.
		Update("people").
		SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).
		Scope(scope, id).
		Exec()

	assert.NoError(t, err)

	var person Person
	err = s.Select("*").From("people").Where("id = $1", id).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}
