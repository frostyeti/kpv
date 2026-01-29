package errors

import (
	stderrors "errors"
	"fmt"
)

type Cause interface {
	Cause() error
}

type UserError interface {
	UserError() string
}

type err struct {
	message string
	code    string
	cause   error
}

type userErr struct {
	err
	userMessage string
}

func ToUserError(e error) UserError {
	if ue, ok := e.(*userErr); ok {
		return ue
	}

	return &userErr{
		err: err{
			message: e.Error(),
			code:    "user-error",
			cause:   e,
		},
		userMessage: e.Error(),
	}
}

func (e *err) Error() string {
	return e.message
}

func (ue *userErr) UserError() string {
	return ue.userMessage
}

func (ue *userErr) Error() string {
	return ue.message
}

func New(message string) error {
	return &err{
		message: message,
		code:    "error",
	}
}

func NewUserError(message, userMessage string) error {
	return &userErr{
		err: err{
			message: message,
			code:    "user-error",
		},
		userMessage: userMessage,
	}
}

func WithUserMessage(e error, userMessage string) error {
	if e == nil {
		return nil
	}

	if ue, ok := e.(*userErr); ok {
		ue.userMessage = userMessage
		return ue
	}

	cause := e
	if ce, ok := e.(*err); ok {
		cause = ce.cause
	}

	return &userErr{
		err: err{
			message: e.Error(),
			code:    "user-error",
			cause:   cause,
		},
		userMessage: userMessage,
	}
}

func WithUserMessagef(e error, format string, args ...interface{}) error {
	if e == nil {
		return nil
	}

	userMessage := fmt.Sprintf(format, args...)

	if ue, ok := e.(*userErr); ok {
		ue.userMessage = userMessage
		return ue
	}

	cause := e
	if ce, ok := e.(*err); ok {
		cause = ce.cause
	}

	return &userErr{
		err: err{
			message: e.Error(),
			code:    "user-error",
			cause:   cause,
		},
		userMessage: userMessage,
	}
}

func WithCause(e error, cause error) error {
	if e == nil {
		return nil
	}

	if ue, ok := e.(*userErr); ok {
		ue.cause = cause
		return ue
	}

	if e, ok := e.(*err); ok {
		e.cause = cause
		return e
	}

	return &err{
		message: e.Error(),
		code:    "error",
		cause:   cause,
	}
}

func WithCode(e error, code string) error {
	if e == nil {
		return nil
	}

	if e, ok := e.(*err); ok {
		e.code = code
		return e
	}

	return &err{
		message: e.Error(),
		code:    code,
	}
}

func (e *err) Cause() error {
	if e.cause != nil {
		return e.cause
	}
	return stderrors.New(e.message)
}

func (e *err) Code() string {
	if e.code != "" {
		return e.code
	}
	if e.cause != nil {
		if cause, ok := e.cause.(*err); ok {
			return cause.Code()
		}
	}
	return "error"
}

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(err error, target interface{}) bool { return stderrors.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}
