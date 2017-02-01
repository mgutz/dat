package dat

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// UnsafeString is interpolated as an unescaped and unquoted value and should
// only be used to create constants.
type UnsafeString string

// Interpolator is the interface for types which interpolate.
type Interpolator interface {
	Interpolate() (string, error)
}

// Expressioner is an interface that returns raw SQL with possible arguments.
type Expressioner interface {
	Expression() (string, []interface{}, error)
}

// Value implements a valuer for compatibility
func (u UnsafeString) Value() (driver.Value, error) {
	panic("UnsafeStrings and its constants NOW, DEFAULT ... are disabled when EnableInterpolation==false")
}

// DEFAULT SQL value
const DEFAULT = UnsafeString("DEFAULT")

// NOW SQL value
const NOW = UnsafeString("NOW()")

// NullString is a type that can be null or a string
type NullString struct {
	sql.NullString
}

// NullFloat64 is a type that can be null or a float64
type NullFloat64 struct {
	sql.NullFloat64
}

// NullInt64 is a type that can be null or an int
type NullInt64 struct {
	sql.NullInt64
}

// NullTime is a type that can be null or a time
type NullTime struct {
	pq.NullTime
}

// NullBool is a type that can be null or a bool
type NullBool struct {
	sql.NullBool
}

// NullStringFrom creates a valid NullString
func NullStringFrom(v string) NullString {
	return NullString{sql.NullString{String: v, Valid: true}}
}

// NullFloat64From creates a valid NullFloat64
func NullFloat64From(v float64) NullFloat64 {
	return NullFloat64{sql.NullFloat64{Float64: v, Valid: true}}
}

// NullInt64From creates a valid NullInt64
func NullInt64From(v int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: v, Valid: true}}
}

// NullTimeFrom creates a valid NullTime
func NullTimeFrom(v time.Time) NullTime {
	return NullTime{pq.NullTime{Time: v, Valid: true}}
}

// NullBoolFrom creates a valid NullBool
func NullBoolFrom(v bool) NullBool {
	return NullBool{sql.NullBool{Bool: v, Valid: true}}
}

// JSONFromString creates a JSON type from JSON encoded string.
func JSONFromString(encoded string) JSON {
	return []byte(encoded)
}

var nullString = []byte("null")

// MarshalJSON correctly serializes a NullString to JSON
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.String)
		return j, e
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullFloat64 to JSON
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Float64)
		return j, e
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullInt64 to JSON
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Int64)
		return j, e
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullTime to JSON
func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Time)
		return j, e
	}
	return nullString, nil
}

// MarshalJSON correctly serializes a NullBool to JSON
func (n NullBool) MarshalJSON() ([]byte, error) {
	if n.Valid {
		j, e := json.Marshal(n.Bool)
		return j, e
	}
	return nullString, nil
}

// UnmarshalJSON correctly deserializes a NullString from JSON
func (n *NullString) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullInt64 from JSON
func (n *NullInt64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullFloat64 from JSON
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// UnmarshalJSON correctly deserializes a NullTime from JSON
func (n *NullTime) UnmarshalJSON(b []byte) error {
	// scan for null
	if bytes.Equal(b, nullString) {
		return n.Scan(nil)
	}
	// scan for JSON timestamp
	formats := []string{
		// Go
		time.RFC3339Nano,
		// JavaScript JSON.stringify()
		"2006-01-02T15:04:05.000Z",
		// postgres
		"2006-01-02 15:04:05.999999999-07",
	}

	s := string(b)
	s = s[1 : len(s)-1]
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return n.Scan(t)
		}
	}
	return logger.Error("Cannot parse time", "time", s, "formats", formats)
}

// UnmarshalJSON correctly deserializes a NullBool from JSON
func (n *NullBool) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.Scan(s)
}

// JSON is a json.RawMessage, which is a []byte underneath.
// Value() validates the json format in the source, and returns an error if
// the json is not valid.  Scan does no validation.  JSON additionally
// implements `Unmarshal`, which unmarshals the json within to an interface{}
type JSON json.RawMessage

// NewJSON creates a JSON value.
func NewJSON(any interface{}) (*JSON, error) {
	var j JSON
	var err error
	j, err = json.Marshal(any)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// MarshalJSON returns the j as the JSON encoding of j.
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return nullString, nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Interpolate interpolates the value into a string.
func (j JSON) Interpolate() (string, error) {
	return string(j), nil
}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j JSON) Value() (driver.Value, error) {
	var m json.RawMessage
	var err = j.Unmarshal(&m)
	if err != nil {
		return []byte{}, err
	}
	return []byte(j), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *JSON) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case string:
		source = []byte(src.(string))
	case []byte:
		source = src.([]byte)
	default:
		return errors.New("Incompatible type for JSON")
	}
	*j = append((*j)[0:0], source...)
	return nil
}

// Unmarshal unmarshal's the json in j to v, as in json.Unmarshal.
func (j *JSON) Unmarshal(v interface{}) error {
	return json.Unmarshal([]byte(*j), v)
}
