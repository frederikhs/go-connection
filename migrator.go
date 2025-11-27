package connection

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"io/fs"
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

func (conn *Conn) GetMigratorFromFs(fs fs.FS) (*migrate.Migrate, error) {
	src, err := iofs.New(fs, ".")
	if err != nil {
		return nil, err
	}

	driver, err := conn.getMigratorDriver()
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return nil, err
	}

	return m, err
}
