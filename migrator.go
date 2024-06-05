package connection

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (conn *Conn) getMigratorDriver() (database.Driver, error) {
	return postgres.WithInstance(conn.db.DB, &postgres.Config{})
}

func (conn *Conn) GetMigrator(migrationPath string) (*migrate.Migrate, error) {
	driver, err := conn.getMigratorDriver()
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
	if err != nil {
		return nil, err
	}

	return m, err
}
