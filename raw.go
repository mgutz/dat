package dat

// RawBuilder builds SQL from raw SQL.
type RawBuilder struct {
	Execer

	sql  string
	args []interface{}
}

// NewRawBuilder creates a new RawBuilder for the given SQL string and arguments
func NewRawBuilder(sql string, args ...interface{}) *RawBuilder {
	return &RawBuilder{sql: sql, args: args}
}

// ToSQL implements builder interface
func (b *RawBuilder) ToSQL() (string, []interface{}) {
	return b.sql, b.args
}

// Interpolate interpolates this builder's SQL.
func (b *RawBuilder) Interpolate() (string, []interface{}, error) {
	return interpolate(b)
}

// MustInterpolate interpolates this builder's SQL.
func (b *RawBuilder) MustInterpolate() (string, []interface{}) {
	return mustInterpolate(b)
}
