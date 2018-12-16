// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httperror

import (
	"fmt"
	"net/http"
)

//Error enriches the original go error type with
//DisplayMsg the description for the displaying message; shown to the user
//HTTPStatus is returned as response code in the middleware.AppHandler
//Error the error if available which is logged internally
type Error struct {
	DisplayMsg string `json:"display_message"`
	HTTPStatus int    `json:"status"`
	Err        error  `json:"-"`
}

//New convenient function to returns a new error
func New(httpStatus int, displayMsg string, err error) *Error {
	return &Error{DisplayMsg: displayMsg, Err: err, HTTPStatus: httpStatus}
}

//PermissionDenied returns a permission denied message with code 403 to the user on a specific action and subject.
//The following display message is returned: "You are not allowed to [action] the [subject]."
func PermissionDenied(action, subject string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusForbidden,
		Err:        err,
		DisplayMsg: fmt.Sprintf("You are not allowed to %s the %s.", action, subject),
	}
}

//NotFound returns a not found message with code 404 on a resource.
//The following display message is returned: "The %s was not found."
func NotFound(res string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusNotFound,
		Err:        err,
		DisplayMsg: fmt.Sprintf("The %s was not found.", res),
	}
}

//ValueTooLong returns the following display message with code 422.
//Display message: "The value of [param] is too long. Maximum %d characters are allowed."
func ValueTooLong(param string, nChars int) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        fmt.Errorf("the value of field %s is too long (%d characters are allowed.)", param, nChars),
		DisplayMsg: fmt.Sprintf("The value of %s is too long. Maximum %d characters are allowed.", param, nChars),
	}
}

//InternalServerError returns a internal server error message with code 500.
//Display message: "An internal server error occured."
func InternalServerError(err error) *Error {
	return &Error{
		HTTPStatus: http.StatusInternalServerError,
		DisplayMsg: "An internal server error occured.",
		Err:        err,
	}
}

//ParameterMissing returns a parameter missing message with code 422.
//Display message: "The parameter [param] is invalid."
func ParameterMissing(param string, err error) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        err,
		DisplayMsg: fmt.Sprintf("The parameter %s is invalid.", param),
	}
}

//ValueRequired returns a value required message with code 422.
//Display message: "Please fill out the field [param]."
func ValueRequired(param string) *Error {
	return &Error{
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        fmt.Errorf("the value of field %s was empty", param),
		DisplayMsg: fmt.Sprintf("Please fill out the field %s.", param),
	}
}

func Equals(a error, b error) bool {
	v, ok := a.(*Error)
	v2, ok2 := b.(*Error)

	if ok && ok2 {
		return v.Err == v2.Err
	} else if !ok && !ok2 {
		return v == v2
	} else if ok && !ok2 {
		return v.Err == b
	} else if !ok && ok2 {
		return a == v2.Err
	}

	return false
}

func (e Error) Error() string {
	return fmt.Sprintf("code=[%d], error=[%s], displayMsg=[%s]", e.HTTPStatus, e.Err.Error(), e.DisplayMsg)
}
