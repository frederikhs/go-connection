package connection

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

type Config struct {
	User     string
	Pass     string
	Host     string
	Port     string
	Database string
}

func ConnectFromEnv() *Conn {
	config := &Config{
		User:     os.Getenv("DB_USER"),
		Pass:     os.Getenv("DB_PASS"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Database: os.Getenv("DB_DATABASE"),
	}

	return config.Connect()
}

func (config *Config) Connect() *Conn {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", config.Host, config.Port, config.User, config.Database, config.Pass))
	if err != nil {
		log.Println("unable to connect to database:", err)
		time.Sleep(time.Second * 5)
		return config.Connect()
	}

	if db.Ping() != nil {
		log.Println("unable to connect to database:", err)
		time.Sleep(time.Second * 5)
		return config.Connect()
	}

	log.Println("connected to database")

	return &Conn{db: db}
}

type Conn struct {
	db            *sqlx.DB
	tx            *sqlx.Tx
	enableLogging bool
}

type Transactioner interface {
	Begin() error
	Commit() error
	Rollback() error
}

func (conn *Conn) QueryRow(query string, args ...any) *sql.Row {
	if conn.tx == nil {
		panic(ErrTransactionNotStarted)
	}

	return conn.tx.QueryRow(query, args...)
}

func (conn *Conn) Exec(query string, args ...any) (sql.Result, error) {
	if conn.tx == nil {
		panic(ErrTransactionNotStarted)
	}

	return conn.tx.Exec(query, args...)
}

func (conn *Conn) Select(dest interface{}, query string, args ...interface{}) error {
	if conn.tx != nil {
		return conn.tx.Select(dest, query, args...)
	}

	return conn.db.Select(dest, query, args...)
}

func (conn *Conn) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if conn.tx != nil {
		return conn.tx.SelectContext(ctx, dest, query, args...)
	}

	return conn.db.SelectContext(ctx, dest, query, args...)
}

func (conn *Conn) Rebind(s string) string {
	if conn.tx != nil {
		return conn.tx.Rebind(s)
	}

	return conn.db.Rebind(s)
}

func (conn *Conn) Begin() error {
	tx, err := conn.db.Beginx()
	if err != nil {
		return err
	}

	conn.tx = tx

	return nil
}

func (conn *Conn) Commit() error {
	if conn.tx == nil {
		return ErrCommit
	}

	err := conn.tx.Commit()
	conn.tx = nil

	return err
}

func (conn *Conn) Rollback() error {
	if conn.tx == nil {
		return ErrRollback
	}

	err := conn.tx.Rollback()
	conn.tx = nil

	return err
}
