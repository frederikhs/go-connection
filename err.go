package connection

import (
	"errors"
	"fmt"
)

var (
	ErrTransactionNotStarted = errors.New("transaction not started")
	ErrCommit                = fmt.Errorf("cannot commit, %w", ErrTransactionNotStarted)
	ErrRollback              = fmt.Errorf("cannot rollback, %w", ErrTransactionNotStarted)
)
