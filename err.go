package connection

import (
	"errors"
)

var (
	ErrTransactionNotStarted     = errors.New("transaction not started")
	ErrTransactionAlreadyStarted = errors.New("transaction already started")
)
