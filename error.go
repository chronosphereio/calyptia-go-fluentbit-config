package fluentbitconfig

import (
	"errors"
	"fmt"
)

var ErrMissingName = errors.New("missing name property")

// LinedError with information about the line number
// where the error was found while parsing.
type LinedError struct {
	Msg  string
	Line uint
}

func (e *LinedError) Error() string {
	return fmt.Sprintf("%d: %s", e.Line, e.Msg)
}

func NewLinedError(msg string, line uint) *LinedError {
	return &LinedError{Msg: msg, Line: line}
}

func WrapLinedError(err error, line uint) *LinedError {
	return &LinedError{Msg: err.Error(), Line: line}
}

type UnknownPluginError struct {
	Kind SectionKind
	Name string
}

func NewUnknownPluginError(kind SectionKind, name string) *UnknownPluginError {
	return &UnknownPluginError{Kind: kind, Name: name}
}

func (e *UnknownPluginError) Error() string {
	return fmt.Sprintf("%s: unknown plugin %q", e.Kind, e.Name)
}
