// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httperror

import (
	"fmt"
	"net/http"
)

// Error enriches the original go error type with
// 	DisplayMsg the human readable display message
// 	HTTPStatus the HTTP status code
// 	Err the internal error
type Error struct {
	DisplayMsg string `json:"display_message"`
	HTTPStatus int    `json:"status"`
	Err        error  `json:"-"`
}

// New returns a new error
func New(httpStatus int, displayMsg string, err error) *Error {
	return &Error{DisplayMsg: displayMsg, Err: err, HTTPStatus: httpStatus}
}

// PermissionDenied returns a permission denied message with code 403.
// The following display message is returned: "You are not allowed to [action] the [subject]."
func PermissionDenied(action, subject string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusForbidden,
		Err:        err,
		DisplayMsg: fmt.Sprintf("You are not allowed to %s the %s.", action, subject),
	}
}

// NotFound returns a not found message with code 404.
// The following display message is returned: "The [res] was not found."
func NotFound(res string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusNotFound,
		Err:        err,
		DisplayMsg: fmt.Sprintf("The %s was not found.", res),
	}
}

// ValueTooLong returns the following display message with code 422.
// Display message: "The value of [param] is too long. Maximum [nchars] characters are allowed."
func ValueTooLong(param string, nchars int) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        fmt.Errorf("the value of field %s is too long (%d characters are allowed.)", param, nchars),
		DisplayMsg: fmt.Sprintf("The value of %s is too long. Maximum %d characters are allowed.", param, nchars),
	}
}

// InternalServerError returns a internal server error message with code 500.
// Display message: "An internal server error occurred."
func InternalServerError(err error) *Error {
	return &Error{
		HTTPStatus: http.StatusInternalServerError,
		DisplayMsg: "An internal server error occurred.",
		Err:        err,
	}
}

// ParameterMissing returns a parameter missing message with code 422.
// Display message: "The parameter [param] is invalid."
func ParameterMissing(param string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        err,
		DisplayMsg: fmt.Sprintf("The parameter %s is invalid.", param),
	}
}

// ValueRequired returns a value required message with code 422.
// Display message: "Please fill out the field [param]."
func ValueRequired(param string) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        fmt.Errorf("the value of field %s was empty", param),
		DisplayMsg: fmt.Sprintf("Please fill out the field %s.", param),
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("code=[%d], error=[%s], displayMsg=[%s]", e.HTTPStatus, e.Err.Error(), e.DisplayMsg)
}

func (e *Error) Unwrap() error {
	return e.Err
}
