package connection

import (
	"errors"
	"os"
	"testing"
)

func givenDatabaseConnectionFromEnv() *Conn {
	_ = os.Setenv("DB_USER", "test")
	_ = os.Setenv("DB_PASS", "test")
	_ = os.Setenv("DB_DATABASE", "test")
	_ = os.Setenv("DB_HOST", "localhost")
	_ = os.Setenv("DB_PORT", "3671")

	return ConnectFromEnv()
}

func TestConnectCanBeginAndRollbackTransaction(t *testing.T) {
	connection := givenDatabaseConnectionFromEnv()

	if connection.db == nil {
		t.Fatalf("database property nil")
	}

	if connection.tx != nil {
		t.Fatalf("transaction property not nil")
	}

	err := connection.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	if connection.tx == nil {
		t.Fatalf("expected transaction not to be nil when transaction has begun")
	}

	err = connection.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}
}

func TestCommit(t *testing.T) {
	connection := givenDatabaseConnectionFromEnv()

	err := connection.Commit()
	if err == nil || !errors.Is(err, ErrTransactionNotStarted) {
		t.Fatalf("expected commit to fail when no transaction has begun")
	}

	_ = connection.Begin()

	_, err = connection.Exec("CREATE TABLE test (id varchar)")
	if err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	_, err = connection.Exec("INSERT INTO test (id) VALUES ($1), ($2), ($3)", "1", "3", "5")
	if err != nil {
		t.Fatalf("failed to insert values into testing table: %v", err)
	}

	err = connection.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	_ = connection.Begin()
	_, err = connection.Exec("DROP TABLE test")
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	err = connection.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}
}

func TestRollback(t *testing.T) {
	connection := givenDatabaseConnectionFromEnv()

	err := connection.Rollback()
	if err == nil || !errors.Is(err, ErrTransactionNotStarted) {
		t.Fatalf("expected rollback to fail when no transaction has begun")
	}
}

func TestBegin(t *testing.T) {
	connection := givenDatabaseConnectionFromEnv()

	err := connection.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	err = connection.Begin()
	if err == nil || !errors.Is(err, ErrTransactionAlreadyStarted) {
		t.Fatalf("expected begin to fail when transaction has already begun")
	}
}
