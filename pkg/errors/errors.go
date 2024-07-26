package errors

import (
	"errors"
	"fmt"
)

func Wrap(e error, s string) error {
	errStr := fmt.Sprintf("%s - %s", e, s)
	return errors.New(errStr)
}

func New(s string) error {
	return errors.New(s)
}
