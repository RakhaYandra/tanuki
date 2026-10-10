// Package errors defines stable exit codes and structured errors.
package errors

// Exit codes per PRD §14.
const (
	CodeOK       = 0
	CodeGeneral  = 1
	CodeConfig   = 2
	CodePrereq   = 3
	CodeConflict = 4
	CodeStartup  = 5
	CodeShutdown = 6
	CodeState    = 7
)

// Error is a structured CLI error with an exit code.
type Error struct {
	Code int
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error { return e.Err }

// New creates an Error with a code and message.
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// Wrap attaches a cause to an Error.
func Wrap(code int, msg string, err error) *Error {
	return &Error{Code: code, Msg: msg, Err: err}
}

// CodeOf extracts the exit code, defaulting to CodeGeneral.
func CodeOf(err error) int {
	if err == nil {
		return CodeOK
	}
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return CodeGeneral
}
