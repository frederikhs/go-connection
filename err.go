package connection

import (
	"errors"
)

var (
	ErrTransactionNotStarted      = errors.New("transaction not started")
	ErrTransactionAlreadyStarted  = errors.New("transaction already started")
	ErrSavePointsNotEnabled       = errors.New("savepoints not enabled")
	ErrSavePointsAlreadyEnabled   = errors.New("savepoints are already enabled")
	ErrSavePointsStillNotReleased = errors.New("savepoints still not released")
)
