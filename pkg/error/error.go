package error

import (
	"errors"
	"fmt"
)

var ErrItemNotFound = errors.New("item cloud not be found")

type ErrorDetails interface {
	Error() string
	Type() error
}

type errDetails struct {
	errType error
	details interface{}
}

func NewErrDetails(err error, details ...interface{}) ErrorDetails {
	return &errDetails{
		errType: err,
		details: details,
	}
}

func (err *errDetails) Error() string {
	return fmt.Sprintf("%v: %v", err.errType, err.details)
}

func (err *errDetails) Type() error {
	return err.errType
}
