package api

import (
	"errors"
	"fmt"
)

var (
	ErrFailureStatus = errors.New("returned status is false")
)

// ErrCall is returned when an API call returned an error.
type ErrCall struct {
	ErrorType string
}

func (e ErrCall) Error() string {
	return fmt.Sprintf("error type %s was returned by the server", e.ErrorType)
}

var _ error = (*ErrCall)(nil)

// ErrUnexpectedStatus is returned when an API call returned an unexpected
// status code.
type ErrUnexpectedStatus struct {
	StatusCode int
}

func (e ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code %d returned by the server", e.StatusCode)
}

var _ error = (*ErrUnexpectedStatus)(nil)
