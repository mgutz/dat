package dat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lib/pq"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestNullStringFrom(t *testing.T) {
	v := "foo"
	n := NullStringFrom(v)

	assert.True(t, n.Valid)
	assert.Equal(t, n.String, v)
}

func TestNullFloat64From(t *testing.T) {
	v := 42.2
	n := NullFloat64From(v)

	assert.True(t, n.Valid)
	assert.Equal(t, n.Float64, v)
}

func TestNullInt64From(t *testing.T) {
	v := int64(400)
	n := NullInt64From(v)

	assert.True(t, n.Valid)
	assert.Equal(t, n.Int64, v)
}

func TestNullTimeFrom(t *testing.T) {
	v := time.Now()
	n := NullTimeFrom(v)

	assert.True(t, n.Valid)
	assert.Equal(t, n.Time, v)
}

func TestNullBoolFrom(t *testing.T) {
	v := false
	n := NullBoolFrom(v)

	assert.True(t, n.Valid)
	assert.Equal(t, n.Bool, v)
}

func TestInvalidNullTime(t *testing.T) {
	n := NullTime{pq.NullTime{Valid: false}}

	assert.False(t, n.Valid)
	var when time.Time
	assert.Equal(t, n.Time, when)
}

func TestJSONFromString(t *testing.T) {
	type foo struct {
		Jason   JSON `json:"jason"`
		NoValue JSON `json:"noValue"`
	}

	j := foo{Jason: JSONFromString(`{"fruit":"apple"}`)}
	b, err := json.Marshal(j)
	if err != nil {
		t.Error("Expected struct with JSON string fields to marshal", err)
	}
	if string(b) != `{"jason":{"fruit":"apple"},"noValue":null}` {
		t.Error("Expected JSON to defaul to null", string(b))
	}
}

func TestNullMarshalling(t *testing.T) {
	type nully struct {
		Int  NullInt64  `json:"int"`
		Intp *NullInt64 `json:"intp"`
		Intv NullInt64  `json:"intv"`

		Time  NullTime  `json:"time"`
		Timep *NullTime `json:"timep"`

		Jason  JSON  `json:"jason"`
		Jasonp *JSON `json:"jasonp"`
	}

	a := nully{Intv: NullInt64From(100)}

	b, err := json.Marshal(a)
	if err != nil {
		t.Error("Expected struct with null fields to marshal", err)
	}
	if string(b) != `{"int":null,"intp":null,"intv":100,"time":null,"timep":null,"jason":null,"jasonp":null}` {
		t.Error("Expected nulltypes to defaul to null", string(b), err)
	}
}
