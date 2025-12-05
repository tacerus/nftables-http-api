package nftables

import (
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrUnknownFamily = errors.New("Unknown family")
)

type nftError struct {
	Op  string
	Err error
}

func (e *nftError) Unwrap() error {
	return e.Err
}

func (e *nftError) Error() string {
	return fmt.Sprintf("%s => %s", e.Op, e.Err)
}

func newNftError(op string, inner error) error {
	e := &nftError{
		Op:  op,
		Err: inner,
	}

	slog.Debug(e.Error())

	return e
}
