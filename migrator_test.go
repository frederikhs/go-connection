package connection

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMigrator(t *testing.T) {
	t.Parallel()
	config := givenLaunchedPostgresContainerAndConfig(t)

	connection := config.Connect()
	t.Cleanup(func() {
		_ = connection.Rollback()
	})

	type Account struct {
		AccountId        int       `db:"account_id"`
		AccountName      string    `db:"account_name"`
		AccountCreatedAt time.Time `db:"account_created_at"`
	}

	assertAccountsLength := func(length int) {
		var accounts []Account
		err := connection.Select(&accounts, "SELECT * FROM test.account")
		assert.NoError(t, err)
		assert.Len(t, accounts, length)
	}

	migrator, err := connection.GetMigrator("file://_testing/migrations")
	assert.NoError(t, err)

	err = migrator.Up()
	assert.NotErrorIs(t, err, migrate.ErrNoChange)

	err = connection.Begin()
	assert.NoError(t, err)

	_, err = connection.Exec("INSERT INTO test.account (account_name) VALUES ($1), ($2)", "test d", "test e")
	assert.NoError(t, err)

	// 3 accounts come from migrations file, 2 from test
	assertAccountsLength(5)

	err = connection.Rollback()
	assert.NoError(t, err)

	assertAccountsLength(3)

	err = migrator.Down()
	assert.NoError(t, err)
}
