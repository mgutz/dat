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

// Interpolate interpolates this builder's SQL.
func (b *RawBuilder) Interpolate() (string, []interface{}, error) {
	return interpolate(b)
}

// IsInterpolated determines if this builder will interpolate when
// Interpolate() is called.
func (b *RawBuilder) IsInterpolated() bool {
	return b.isInterpolated
}

// SetIsInterpolated sets whether this builder should interpolate.
func (b *RawBuilder) SetIsInterpolated(enable bool) *RawBuilder {
	b.isInterpolated = enable
	return b
}
