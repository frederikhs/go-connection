package connection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type containerOptions struct {
	ctx        context.Context
	dbUser     string
	dbPassword string
	dbName     string
	sslMode    SSLMode
	searchPath *string
}

func defaultContainerOptions() containerOptions {
	return containerOptions{
		ctx:        context.Background(),
		dbUser:     "aaa",
		dbPassword: "bbb",
		dbName:     "ccc",
		sslMode:    SSLModeDisable,
		searchPath: nil,
	}
}

func givenLaunchedPostgresContainerAndConfig(opt containerOptions, t *testing.T) *Config {
	postgresContainer := launchPostgresContainer(opt, t)

	t.Cleanup(func() {
		err := postgresContainer.Terminate(opt.ctx)
		assert.NoError(t, err)
	})

	containerPort, err := postgresContainer.MappedPort(opt.ctx, "5432/tcp")
	assert.NoError(t, err)

	host, err := postgresContainer.Host(opt.ctx)
	assert.NoError(t, err)

	return NewConfig(opt.dbUser, opt.dbPassword, host, containerPort.Port(), opt.dbName, SSLModeDisable, opt.searchPath)
}

func launchPostgresContainer(opt containerOptions, t *testing.T) testcontainers.Container {
	postgresContainer, err := postgres.RunContainer(opt.ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(opt.dbName),
		postgres.WithUsername(opt.dbUser),
		postgres.WithPassword(opt.dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	assert.NoError(t, err)

	return postgresContainer
}

func TestTransactions(t *testing.T) {
	t.Parallel()
	connection := givenLaunchedPostgresContainerAndConfig(defaultContainerOptions(), t).Connect()

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

	t.Run("TestSearchPath", func(t *testing.T) {
		testSearchPath(t)
	})

	t.Run("TestSearchPathDefault", func(t *testing.T) {
		testSearchPathDefault(t)
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

func testSearchPath(t *testing.T) {
	opt := defaultContainerOptions()
	path := "test"
	opt.searchPath = &path
	conn := givenLaunchedPostgresContainerAndConfig(opt, t).Connect()

	t.Cleanup(func() { assert.NoError(t, conn.db.Close()) })

	var actual string
	err := conn.db.QueryRow("SHOW search_path").Scan(&actual)
	assert.NoError(t, err)
	assert.Equal(t, path, actual)
}

func testSearchPathDefault(t *testing.T) {
	opt := defaultContainerOptions()
	conn := givenLaunchedPostgresContainerAndConfig(opt, t).Connect()

	t.Cleanup(func() { assert.NoError(t, conn.db.Close()) })

	var actual string
	err := conn.db.QueryRow("SHOW search_path").Scan(&actual)
	assert.NoError(t, err)
	assert.Equal(t, "public", actual)
}
