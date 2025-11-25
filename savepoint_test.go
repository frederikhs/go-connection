package connection

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSavePoints(t *testing.T) {
	t.Parallel()
	config := givenLaunchedPostgresContainerAndConfig(t)

	t.Run("TestEnableAndStartATransactionWithSavePoint", func(t *testing.T) {
		testEnableAndStartATransactionWithSavePoint(t, config.Connect())
	})

	t.Run("TestEnableDisableNestedTransactions", func(t *testing.T) {
		testEnableDisableNestedTransactions(t, config.Connect())
	})

	t.Run("TestEnableNestedTransactionAndWriteToSavePoint", func(t *testing.T) {
		testEnableNestedTransactionAndWriteToSavePoint(t, config.Connect())
	})

	t.Run("TestEnableNestedTransactionAndRollbackToSavePointBeforeCommitting", func(t *testing.T) {
		testEnableNestedTransactionAndRollbackToSavePointBeforeCommitting(t, config.Connect())
	})

	t.Run("TestRollbackAll", func(t *testing.T) {
		testRollbackAll(t, config.Connect())
	})
}

func testEnableAndStartATransactionWithSavePoint(t *testing.T, connection *Conn) {
	err := connection.EnableNestedTransactions()
	require.NoError(t, err)

	err = connection.Begin()
	assert.NoError(t, err)
	assert.Empty(t, connection.savePoints)

	err = connection.Begin()
	assert.NoError(t, err)
	assert.NotEmpty(t, connection.savePoints)
}

func testEnableDisableNestedTransactions(t *testing.T, connection *Conn) {
	err := connection.Begin()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)

	err = connection.Begin()
	require.ErrorIs(t, err, ErrTransactionAlreadyStarted)
	require.Empty(t, connection.savePoints)

	err = connection.EnableNestedTransactions()
	require.NoError(t, err)

	err = connection.EnableNestedTransactions()
	require.ErrorIs(t, err, ErrSavePointsAlreadyEnabled)

	err = connection.Begin()
	require.NoError(t, err)
	require.NotEmpty(t, connection.savePoints)

	err = connection.DisableNestedTransactions()
	require.ErrorIs(t, err, ErrSavePointsStillNotReleased)

	err = connection.Rollback()
	require.NoError(t, err)

	err = connection.DisableNestedTransactions()
	require.NoError(t, err)

	err = connection.DisableNestedTransactions()
	require.ErrorIs(t, err, ErrSavePointsNotEnabled)

	err = connection.Begin()
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrTransactionAlreadyStarted)
	require.Empty(t, connection.savePoints)
}

func testEnableNestedTransactionAndWriteToSavePoint(t *testing.T, connection *Conn) {
	err := connection.EnableNestedTransactions()
	require.NoError(t, err)

	_, err = connection.db.Exec("CREATE TABLE table_1 (id INT)")
	require.NoError(t, err)

	defer func() {
		err = connection.RollbackAll()
		require.NoError(t, err)

		_, err = connection.db.Exec("DROP TABLE table_1")
		require.NoError(t, err)
	}()

	err = connection.Begin()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)
	_, err = connection.Exec("INSERT INTO table_1 VALUES (1)")
	require.NoError(t, err)

	err = connection.Begin()
	require.NoError(t, err)
	require.NotEmpty(t, connection.savePoints)
	_, err = connection.Exec("INSERT INTO table_1 VALUES (2)")
	require.NoError(t, err)

	var v []int
	err = connection.Select(&v, "SELECT id FROM table_1")
	require.NoError(t, err)
	require.Len(t, v, 2)

	err = connection.Commit()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)
	require.NotNil(t, connection.tx)

	err = connection.Commit()
	require.NoError(t, err)
	require.Nil(t, connection.tx)
}

func testEnableNestedTransactionAndRollbackToSavePointBeforeCommitting(t *testing.T, connection *Conn) {
	err := connection.EnableNestedTransactions()
	require.NoError(t, err)

	_, err = connection.db.Exec("CREATE TABLE table_1 (id INT)")
	require.NoError(t, err)

	defer func() {
		err = connection.RollbackAll()
		require.NoError(t, err)

		_, err = connection.db.Exec("DROP TABLE table_1")
		require.NoError(t, err)
	}()

	err = connection.Begin()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)
	_, err = connection.Exec("INSERT INTO table_1 VALUES (6)")
	require.NoError(t, err)

	err = connection.Begin()
	require.NoError(t, err)
	require.NotEmpty(t, connection.savePoints)
	_, err = connection.Exec("INSERT INTO table_1 VALUES (7)")
	require.NoError(t, err)

	err = connection.Rollback()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)
	require.NotNil(t, connection.tx)

	var v []int
	err = connection.Select(&v, "SELECT id FROM table_1")
	require.NoError(t, err)
	require.Len(t, v, 1)
	require.Equal(t, v[0], 6)

	err = connection.Rollback()
	require.NoError(t, err)
	require.Empty(t, connection.savePoints)
	require.Nil(t, connection.tx)
}

func testRollbackAll(t *testing.T, connection *Conn) {
	err := connection.EnableNestedTransactions()
	require.NoError(t, err)

	err = connection.Begin()
	require.NoError(t, err)

	err = connection.Begin()
	require.NoError(t, err)

	err = connection.Begin()
	require.NoError(t, err)

	err = connection.RollbackAll()
	require.NoError(t, err)

	err = connection.Rollback()
	require.ErrorIs(t, err, ErrTransactionNotStarted)
}
