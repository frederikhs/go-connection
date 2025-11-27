package connection

import (
	"embed"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
	"io/fs"
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
		require.NoError(t, err)
		require.Len(t, accounts, length)
	}

	migrator, err := connection.GetMigrator("file://_testing/migrations")
	require.NoError(t, err)

	err = migrator.Up()
	require.NotErrorIs(t, err, migrate.ErrNoChange)

	err = connection.Begin()
	require.NoError(t, err)

	_, err = connection.Exec("INSERT INTO test.account (account_name) VALUES ($1), ($2)", "test d", "test e")
	require.NoError(t, err)

	// 3 accounts come from migrations file, 2 from test
	assertAccountsLength(5)

	err = connection.Rollback()
	require.NoError(t, err)

	assertAccountsLength(3)

	err = migrator.Down()
	require.NoError(t, err)
}

//go:embed _testing/migrations/*.sql
var MigrationsFS embed.FS

func TestEmbeddedFileSystemAsMigrationsPath(t *testing.T) {
	sub, err := fs.Sub(MigrationsFS, "_testing/migrations")
	require.NoError(t, err)

	t.Parallel()
	config := givenLaunchedPostgresContainerAndConfig(t)
	connection := config.Connect()

	migrator, err := connection.GetMigratorFromFs(sub)
	require.NoError(t, err)

	err = migrator.Up()
	require.NotErrorIs(t, err, migrate.ErrNoChange)

	driver, err := connection.getMigratorDriver()
	require.NoError(t, err)

	version, dirty, err := driver.Version()
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, 1, version)
}
