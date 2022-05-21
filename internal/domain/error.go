package domain

import (
	"fmt"
	"net/http"
)

type Error interface {
	Error() string
	// ClientMsg returns messages clients should know
	ClientMsg() string
}

// InternalError is used when process goes wrong
type InternalError struct {
	clientMsg string
	err       error
}

func NewInternalError(clientMsg string, err error) Error {
	if err, ok := err.(Error); ok {
		return err
	}
	return InternalError{
		clientMsg: clientMsg,
		err:       err,
	}
}

func (e InternalError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e InternalError) ClientMsg() string {
	if e.clientMsg == "" {
		return e.Error()
	}
	return e.clientMsg
}

// ExternalError is used when calling external service goes wrong
type ExternalError struct {
	// Custom status code
	statusCode *int
	clientMsg  string
	err        error
}

func NewExternalError(clientMsg string, statusCode *int, err error) Error {
	if err, ok := err.(Error); ok {
		return err
	}
	return ExternalError{
		clientMsg:  clientMsg,
		statusCode: statusCode,
		err:        err,
	}
}

func (e ExternalError) Error() string {
	var msg string
	if e.statusCode != nil {
		msg += fmt.Sprintf("%v: ", *e.statusCode)
	}

	if e.err != nil {
		msg += fmt.Sprintf("%s: ", e.err.Error())
	}
	return msg
}

func (e ExternalError) ClientMsg() string {
	if e.clientMsg == "" {
		return e.Error()
	}
	return e.clientMsg
}
func (e ExternalError) StatusCode() int {
	if e.statusCode != nil {
		return *e.statusCode
	}
	return http.StatusInternalServerError
}

// ResourceNotFoundError is used when resources cannot be found
type ResourceNotFoundError struct {
	clientMsg string
	err       error
}

func NewResourceNotFoundError(clientMsg string, err error) Error {
	if err, ok := err.(Error); ok {
		return err
	}
	return ResourceNotFoundError{
		clientMsg: clientMsg,
		err:       err,
	}
}

func (e ResourceNotFoundError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e ResourceNotFoundError) ClientMsg() string {
	if e.clientMsg == "" {
		return e.Error()
	}
	return e.clientMsg
}

// ParameterError is used when parameters are missing or invalid
type ParameterError struct {
	clientMsg string
	err       error
}

func NewParameterError(clientMsg string, err error) Error {
	if err, ok := err.(Error); ok {
		return err
	}
	return ParameterError{
		clientMsg: clientMsg,
		err:       err,
	}
}

func (e ParameterError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e ParameterError) ClientMsg() string {
	if e.clientMsg == "" {
		return e.Error()
	}
	return e.clientMsg
}
