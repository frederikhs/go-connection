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

type SSLMode string

const (
	SSLModeDisable    SSLMode = "disable"
	SSLModeRequire    SSLMode = "require"
	SSLModeVerifyFull SSLMode = "verify-full"
	SSLModeVerifyCA   SSLMode = "verify-ca"
)

type Config struct {
	User              string
	Pass              string
	Host              string
	Port              string
	Database          string
	Mode              SSLMode
	Logger            *log.Logger
	ConnectionTimeout time.Duration
}

func ConnectFromEnv() *Conn {
	return NewConfigFromEnv().Connect()
}

func NewConfigFromEnv() *Config {
	return &Config{
		User:     os.Getenv("DB_USER"),
		Pass:     os.Getenv("DB_PASS"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Database: os.Getenv("DB_DATABASE"),
	}
}

func NewConfig(user, pass, host, port, database string, mode SSLMode) *Config {
	return &Config{
		User:     user,
		Pass:     pass,
		Host:     host,
		Port:     port,
		Database: database,
		Mode:     mode,
	}
}

func (config *Config) Connect() *Conn {
	var mode SSLMode

	if config.Mode == "" {
		mode = SSLModeDisable
	} else {
		mode = config.Mode
	}

	if config.ConnectionTimeout == time.Duration(0) {
		config.ConnectionTimeout = time.Second * 5
	}

	if config.Logger == nil {
		config.Logger = log.Default()
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s search_path=public", config.Host, config.Port, config.User, config.Database, config.Pass, mode))
	if err != nil {
		config.Logger.Println("unable to connect to database:", err)
		time.Sleep(time.Second * 5)
		return config.Connect()
	}

	if db.Ping() != nil {
		config.Logger.Println("unable to connect to database:", err)
		time.Sleep(time.Second * 5)
		return config.Connect()
	}

	config.Logger.Println("connected to database")

	return &Conn{db: db}
}

type Conn struct {
	db *sqlx.DB
	tx *sqlx.Tx
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

func (conn *Conn) NamedQuery(query string, args any) (*sqlx.Rows, error) {
	if conn.tx == nil {
		return conn.NamedQuery(query, args)
	}

	return conn.tx.NamedQuery(query, args)
}

func (conn *Conn) Exec(query string, args ...any) (sql.Result, error) {
	if conn.tx == nil {
		panic(ErrTransactionNotStarted)
	}

	return conn.tx.Exec(query, args...)
}

func (conn *Conn) NamedExec(query string, arg any) (sql.Result, error) {
	if conn.tx == nil {
		panic(ErrTransactionNotStarted)
	}

	return conn.tx.NamedExec(query, arg)
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
	if conn.tx != nil {
		return ErrTransactionAlreadyStarted
	}

	tx, err := conn.db.Beginx()
	if err != nil {
		return err
	}

	conn.tx = tx

	return nil
}

func (conn *Conn) Commit() error {
	if conn.tx == nil {
		return ErrTransactionNotStarted
	}

	err := conn.tx.Commit()
	conn.tx = nil

	return err
}

func (conn *Conn) Rollback() error {
	if conn.tx == nil {
		return ErrTransactionNotStarted
	}

	err := conn.tx.Rollback()
	conn.tx = nil

	return err
}
