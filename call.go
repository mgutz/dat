package dat

// CallBuilder is a store procedure call builder.
type CallBuilder struct {
	Execer

	args           []interface{}
	isInterpolated bool
	sproc          string
}

// NewCallBuilder creates a new CallBuilder for the given sproc name and args.
func NewCallBuilder(sproc string, args ...interface{}) *CallBuilder {
	if sproc == "" {
		logger.Error("Invalid sproc name", "name", sproc)
		return nil
	}
	return &CallBuilder{sproc: sproc, args: args}
}

// ToSQL serializes CallBuilder to a SQL string returning
// valid SQL with placeholders an a slice of query arguments.
func (b *CallBuilder) ToSQL() (string, []interface{}) {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	buf.WriteString("SELECT * FROM ")
	buf.WriteString(b.sproc)

	length := len(b.args)
	if length > 0 {
		buildPlaceholders(buf, 1, length)
		return buf.String(), b.args
	}
	buf.WriteString("()")
	return buf.String(), nil
}
