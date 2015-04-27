package dat

// RawBuilder builds SQL from raw SQL.
type RawBuilder struct {
	Execer

	isInterpolated bool
	sql            string
	args           []interface{}
}

// NewRawBuilder creates a new RawBuilder for the given SQL string and arguments
func NewRawBuilder(sql string, args ...interface{}) *RawBuilder {
	return &RawBuilder{sql: sql, args: args, isInterpolated: EnableInterpolation}
}

// ToSQL implements builder interface
func (b *RawBuilder) ToSQL() (string, []interface{}) {
	return b.sql, b.args
}
