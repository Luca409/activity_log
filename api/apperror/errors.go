package apperror

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type TimeoutError struct {
	wrappedError error
	timeoutLimit time.Duration
}

func NewTimeoutError(wrappedError error, timeoutLimit time.Duration) *TimeoutError {
	return &TimeoutError{
		wrappedError: wrappedError,
		timeoutLimit: timeoutLimit,
	}
}

func (te *TimeoutError) Error() string {
	return fmt.Sprintf("timeout: %s, error: %s", te.timeoutLimit, te.wrappedError.Error())
}

func (te *TimeoutError) Unwrap() error {
	return te.wrappedError
}

func IsTimeoutError(err error) bool {
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

type UserConfusedError struct {
	wrappedError error
}

func NewUserConfusedError(wrappedError error) *UserConfusedError {
	return &UserConfusedError{
		wrappedError: wrappedError,
	}
}

func (uce *UserConfusedError) Error() string {
	return uce.wrappedError.Error()
}

func (uce *UserConfusedError) Unwrap() error {
	return uce.wrappedError
}

func IsUserConfusedError(err error) bool {
	var userConfusedErr *UserConfusedError
	return errors.As(err, &userConfusedErr)
}

type NotFoundError struct {
	wrappedError error
}

func NewNotFoundError(wrappedError error) *NotFoundError {
	return &NotFoundError{
		wrappedError: wrappedError,
	}
}

func (nfe *NotFoundError) Error() string {
	return nfe.wrappedError.Error()
}

func (nfe *NotFoundError) Unwrap() error {
	return nfe.wrappedError
}

func IsNotFoundError(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr) || errors.Is(err, os.ErrNotExist)
}
