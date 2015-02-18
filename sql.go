package dat

// RawBuilder builds SQL from raw SQL.
type RawBuilder struct {
	sql  string
	args []interface{}
}

// ToSql implements builder interface
func (b *RawBuilder) ToSQL() (string, []interface{}) {
	return b.sql, b.args
}

// MustString interpolates this builders sql.
func (b *RawBuilder) Interpolate() (string, error) {
	return interpolate(b)
}

// MustString interpolates this builders sql.
func (b *RawBuilder) MustInterpolate() string {
	return mustInterpolate(b)
}
