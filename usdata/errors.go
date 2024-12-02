package usdata

import (
	"errors"
)

type (
	ErrorCode   int
	Apierr      error
	StringToErr func(string) APIError
)

type APIError struct {
	StatusCode ErrorCode `"json:"status_code"`
	Msg        string    `"json:"msg"`
}

func EzError(code ErrorCode) StringToErr {
	return func(s string) APIError {
		new := errors.New(s)
		return APIError{
			StatusCode: code,
			Msg:        new.Error(),
		}
	}
}

var (
	DreamsOfOhio = EzError(405)(" skill issues ")
)

// PARTIAL FUNC
func CreateError(code ErrorCode) func(Apierr) APIError {
	return func(msg Apierr) APIError {
		return APIError{
			StatusCode: code,
			Msg:        msg.Error(),
		}
	}
}

var (
	userErrs      = CreateError(400)
	internalErrs  = CreateError(500)
	WrongPass     = userErrs(WRONG_PASSWORD)
	InvalidMethod = userErrs(INVALID_METHOD)
	BadInput      = userErrs(BAD_INPUT)
	BadHeaders    = userErrs(BAD_HEADERS)
	AuthFail      = userErrs(AUTH_FAIL)
	InternalError = internalErrs(INTERNAL_ERROR)
)

var (
	WRONG_PASSWORD Apierr = (errors.New("Incorrect password"))
	INVALID_NAME   Apierr = (errors.New("Incorrect username"))
	INVALID_METHOD Apierr = (errors.New("Invalid Request Method"))
	INVALID_JSON   Apierr = (errors.New("Invalid Json"))
	BAD_INPUT      Apierr = (errors.New("Invalid input"))
	BAD_HEADERS    Apierr = (errors.New("Invalid header detected"))
	AUTH_FAIL      Apierr = (errors.New("Failed to validate header token"))
	//
	ANT_REQ_FAILED Apierr = (errors.New("Anthropic request/response error"))
	INTERNAL_ERROR Apierr = (errors.New("Internal server error"))
)
