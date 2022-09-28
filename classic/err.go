package classic

import "fmt"

// Error with information about the line number
// where the error was found while parsing.
type Error struct {
	Msg  string
	Line uint
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Line, e.Msg)
}

func NewError(msg string, line uint) *Error {
	return &Error{Msg: msg, Line: line}
}
