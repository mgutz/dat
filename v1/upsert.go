package dat

import (
	"reflect"

	"github.com/mgutz/str"
)

// UpsertBuilder contains the clauses for an INSERT statement
type UpsertBuilder struct {
	Execer

	isInterpolated bool
	table          string
	cols           []string
	isBlacklist    bool
	vals           []interface{}
	record         interface{}
	returnings     []string
	whereFragments []*whereFragment
}

// NewUpsertBuilder creates a new UpsertBuilder for the given table.
func NewUpsertBuilder(table string) *UpsertBuilder {
	if table == "" {
		logger.Error("Insect requires a table name.")
		return nil
	}
	return &UpsertBuilder{table: table, isInterpolated: EnableInterpolation}
}

// Columns appends columns to insert in the statement
func (b *UpsertBuilder) Columns(columns ...string) *UpsertBuilder {
	return b.Whitelist(columns...)
}

// Blacklist defines a blacklist of columns and should only be used
// in conjunction with Record.
func (b *UpsertBuilder) Blacklist(columns ...string) *UpsertBuilder {
	b.isBlacklist = true
	b.cols = columns
	return b
}

// Whitelist defines a whitelist of columns to be inserted. To
// specify all columsn of a record use "*".
func (b *UpsertBuilder) Whitelist(columns ...string) *UpsertBuilder {
	b.cols = columns
	return b
}

// Values appends a set of values to the statement
func (b *UpsertBuilder) Values(vals ...interface{}) *UpsertBuilder {
	b.vals = vals
	return b
}

// Record pulls in values to match Columns from the record
func (b *UpsertBuilder) Record(record interface{}) *UpsertBuilder {
	b.record = record
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *UpsertBuilder) Returning(columns ...string) *UpsertBuilder {
	b.returnings = columns
	return b
}

// ToSQL serialized the UpsertBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *UpsertBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}
	lenCols := len(b.cols)
	if lenCols == 0 {
		panic("no columns specified")
	}
	if len(b.vals) == 0 && b.record == nil {
		panic("no values or records specified")
	}

	if b.record == nil && b.cols[0] == "*" {
		panic(`"*" can only be used in conjunction with Record`)
	}
	if b.record == nil && b.isBlacklist {
		panic(`Blacklist can only be used in conjunction with Record`)
	}
	// build where clause from columns and values
	if len(b.whereFragments) == 0 {
		panic("where clause required for upsert")
	}

	// reflect fields removing blacklisted columns
	if b.record != nil && b.isBlacklist {
		info := reflectFields(b.record)
		lenFields := len(info.fields)
		cols := []string{}
		for i := 0; i < lenFields; i++ {
			f := info.fields[i]
			if str.SliceContains(b.cols, f.dbName) {
				continue
			}
			cols = append(cols, f.dbName)
		}
		b.cols = cols
	}
	// reflect all fields
	if b.record != nil && b.cols[0] == "*" {
		info := reflectFields(b.record)
		b.cols = info.Columns()
	}

	if len(b.returnings) == 0 {
		b.returnings = b.cols
	}

	/*
					   END GOAL:

				WITH
					new_values (id, field1, field2) AS (
						values (1, 'A', 'X'),
							   (2, 'B', 'Y'),
							   (3, 'C', 'Z')
					),
					upsert as
					(
						update mytable m
							set field1 = nv.field1,
								field2 = nv.field2
						FROM new_values nv
						WHERE m.id = nv.id
						RETURNING m.*
					)
				INSERT INTO mytable (id, field1, field2)
				SELECT id, field1, field2
				FROM new_values
				WHERE NOT EXISTS (SELECT 1
				                  FROM upsert up
				                  WHERE up.id = new_values.id)






		Upsert("table").
			Columns("name", "email").
			Values("mario", "mario@barc.com").
			Where("name = $1", "mario").
			Returning("id", "name", "email")
	*/
	if b.record != nil {
		ind := reflect.Indirect(reflect.ValueOf(b.record))
		var err error
		b.vals, err = ValuesFor(ind.Type(), ind, b.cols)
		if err != nil {
			panic(err.Error())
		}
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)

	/*
		WITH
			upd as (
				update people
				set name = $1,
					email = $2
				WHERE name = $3
				RETURNING id, name, email
			),
			ins AS (
				INSERT INTO people (name, email)
				SELECT $1, $2
				WHERE NOT EXISTS (SELECT 1 FROM upd)
				RETURNING id, name, email
			)
		SELECT * FROM upd
		UNION ALL
		SELECT * FROM ins
	*/

	// TODO refactor this, no need to call update
	// builder, just need a few more helper functions
	var args []interface{}

	buf.WriteString("WITH upd AS ( ")

	ub := NewUpdateBuilder(b.table)
	for i, col := range b.cols {
		ub.Set(col, b.vals[i])
	}
	ub.whereFragments = b.whereFragments
	ub.returnings = b.returnings
	updateSQL, args := ub.ToSQL()
	buf.WriteString(updateSQL)

	buf.WriteString("), ins AS (")

	buf.WriteString(" INSERT INTO ")
	writeIdentifier(buf, b.table)
	buf.WriteString("(")
	writeIdentifiers(buf, b.cols, ",")
	buf.WriteString(") SELECT ")

	writePlaceholders(buf, len(b.vals), ",", 1)

	buf.WriteString(" WHERE NOT EXISTS (SELECT 1 FROM upd) RETURNING ")
	writeIdentifiers(buf, b.returnings, ",")

	buf.WriteString(") SELECT * FROM ins UNION ALL SELECT * FROM upd")

	return buf.String(), args
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *UpsertBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *UpsertBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}
