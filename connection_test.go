package connection

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

func givenLaunchedPostgresContainerAndConfig(t *testing.T) *Config {
	ctx := context.Background()

	dbName := "aaa"
	dbUser := "bbb"
	dbPassword := "ccc"

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err = postgresContainer.Terminate(ctx)
		assert.NoError(t, err)
	})

	containerPort, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	assert.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	assert.NoError(t, err)

	return NewConfig(dbUser, dbPassword, host, containerPort.Port(), dbName, SSLModeDisable)
}

func TestTransactions(t *testing.T) {
	t.Parallel()
	connection := givenLaunchedPostgresContainerAndConfig(t).Connect()

	t.Run("TestConnectCanBeginAndRollbackTransaction", func(t *testing.T) {
		testConnectCanBeginAndRollbackTransaction(t, connection)
	})

	t.Run("TestCommit", func(t *testing.T) {
		testCommit(t, connection)
	})

	t.Run("TestRollback", func(t *testing.T) {
		testRollback(t, connection)
	})

	t.Run("TestBegin", func(t *testing.T) {
		testBegin(t, connection)
	})
}

func testConnectCanBeginAndRollbackTransaction(t *testing.T, connection *Conn) {
	assert.NotNil(t, connection.db)
	assert.Nil(t, connection.tx)

	err := connection.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, connection.tx)

	err = connection.Rollback()
	assert.NoError(t, err)
	assert.Nil(t, connection.tx)
}

func testCommit(t *testing.T, connection *Conn) {
	err := connection.Commit()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionNotStarted)

	err = connection.Begin()
	assert.NoError(t, err)

	_, err = connection.Exec("CREATE TABLE test (id varchar)")
	assert.NoError(t, err)

	_, err = connection.Exec("INSERT INTO test (id) VALUES ($1), ($2), ($3)", "1", "3", "5")
	assert.NoError(t, err)

	err = connection.Commit()
	assert.NoError(t, err)

	err = connection.Begin()
	assert.NoError(t, err)
	_, err = connection.Exec("DROP TABLE test")
	assert.NoError(t, err)

	err = connection.Commit()
	assert.NoError(t, err)
}

func testRollback(t *testing.T, connection *Conn) {
	err := connection.Rollback()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionNotStarted)
}

func testBegin(t *testing.T, connection *Conn) {
	err := connection.Begin()
	assert.NoError(t, err)

	err = connection.Begin()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionAlreadyStarted)
}
