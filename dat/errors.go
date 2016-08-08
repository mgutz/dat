package dat

var (
	// ErrNotUTF8 ...
	ErrNotUTF8 = NewError("invalid UTF-8")
	// ErrInvalidSliceLength ...
	ErrInvalidSliceLength = NewError("length of slice is 0. length must be >= 1")
	// ErrInvalidSliceValue ...
	ErrInvalidSliceValue = NewError("trying to interpolate invalid slice value into query")
	// ErrInvalidValue ...
	ErrInvalidValue = NewError("trying to interpolate invalid value into query")
	// ErrArgumentMismatch ...
	ErrArgumentMismatch = NewError("mismatch between number of $placeholders and arguments")
	// ErrTimedout occurs when a query times out.
	ErrTimedout = NewError("query timed out")
	// ErrInvalidOperation occurs when an invalid operation occurs like cancelling
	// an operation without a procPID.
	ErrInvalidOperation = NewError("invalid operation")
	// ErrDisconnectedExecer is returned when a dat builder is used directly instead of through sqlx-runner
	ErrDisconnectedExecer = NewError("dat builders are disconnected, use sqlx-runner package")
)

// Error are errors returned by Dat.
type Error struct {
	Code    int
	Message string
}

// Error returns the enclosed error message.
func (de *Error) Error() string {
	return de.Message
}

// NewError creates a new dat Error.
func NewError(msg string) error {
	return &Error{Message: msg}
}

// NewDatSQLErr is a shortcut for returning an error when building SQL. Should not
// be used by end users.
func NewDatSQLErr(err error) (string, []interface{}, error) {
	return "", nil, err
}
