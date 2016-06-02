package runner

import (
	"testing"
	"time"

	"github.com/mgutz/jo/v1"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

type TimeoutPerson struct {
	Name string `db:"name" json:"name"`
	Age  int    `db:"age" json:"age"`
	NA   string `db:"na" json:"na"`
}

func TestTimeoutExec(t *testing.T) {
	_, err := testDB.SQL("SELECT pg_sleep(1)").Timeout(10 * time.Millisecond).Exec()
	assert.Equal(t, err, dat.ErrTimedout)

	// test no timeout
	result, err := testDB.SQL("SELECT 0").Timeout(3 * time.Second).Exec()
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.NoError(t, err)
}

func TestTimeoutScalar(t *testing.T) {
	var s string
	var n int
	err := testDB.SQL("SELECT pg_sleep(2) as sleep, 1 as k;").Timeout(10*time.Millisecond).QueryScalar(&s, &n)
	assert.Equal(t, dat.ErrTimedout, err)

	// test no timeout
	err = testDB.SQL("SELECT 1 as k").Timeout(10 * time.Millisecond).QueryScalar(&n)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestTimeoutSlice(t *testing.T) {
	var s string
	var arr []int
	err := testDB.SQL("SELECT pg_sleep(2) as sleep, 1 as k;").Timeout(10*time.Millisecond).QueryScalar(&s, &arr)
	assert.Equal(t, dat.ErrTimedout, err)

	// test no timeout
	err = testDB.SQL("SELECT * FROM generate_series(1, 3)").Timeout(1 * time.Second).QuerySlice(&arr)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, arr)
}

func TestTimeoutStruct(t *testing.T) {
	var person TimeoutPerson

	err := testDB.SQL("SELECT pg_sleep(2) as na, 'timeout' as name;").Timeout(10 * time.Millisecond).QueryStruct(&person)
	assert.Equal(t, dat.ErrTimedout, err)

	// test no timeout
	err = testDB.SQL("SELECT 'john' as name, 10 as age").Timeout(1 * time.Second).QueryStruct(&person)
	assert.NoError(t, err)
	assert.Equal(t, "john", person.Name)
	assert.Equal(t, 10, person.Age)
}

func TestTimeoutStructs(t *testing.T) {
	var people []TimeoutPerson

	err := testDB.SQL("SELECT pg_sleep(2) as na, 'timeout' as name;").Timeout(10 * time.Millisecond).QueryStructs(&people)
	assert.Equal(t, dat.ErrTimedout, err)

	// test no timeout
	err = testDB.SQL("SELECT 'john' as name, 10 as age UNION ALL SELECT 'jane' as name, 11 as age").Timeout(1 * time.Second).QueryStructs(&people)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(people))
	assert.Equal(t, "john", people[0].Name)
	assert.Equal(t, "jane", people[1].Name)
}

func TestTimeoutObject(t *testing.T) {
	var person jo.Object

	err := testDB.SQL("SELECT pg_sleep(2) as na, 'timeout' as name").Timeout(10 * time.Millisecond).QueryObject(&person)
	assert.Equal(t, dat.ErrTimedout, err)

	// test no timeout
	err = testDB.SQL("SELECT 'john' as name, 10 as age LIMIT 1").Timeout(1 * time.Second).QueryObject(&person)
	assert.NoError(t, err)
	assert.Equal(t, "john", person.AsString("[0].name"))
	assert.Equal(t, 10, person.AsInt("[0].age"))
}

func TestTimeoutJSON(t *testing.T) {
	b, err := testDB.SQL("SELECT pg_sleep(2) as na, 'timeout' as name").Timeout(10 * time.Millisecond).QueryJSON()
	assert.Equal(t, dat.ErrTimedout, err)
	assert.Equal(t, []byte(nil), b)

	// test no timeout
	b, err = testDB.SQL("SELECT 'john' as name, 10 as age LIMIT 1").Timeout(1 * time.Second).QueryJSON()
	assert.NoError(t, err)
	obj, _ := jo.NewFromBytes(b)
	assert.Equal(t, "john", obj.AsString("[0].name"))
	assert.Equal(t, 10, obj.AsInt("[0].age"))
}
