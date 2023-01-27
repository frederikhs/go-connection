# go-connection

[![Release](https://img.shields.io/github/v/release/frederikhs/go-connection.svg)](https://godoc.org/github.com/frederikhs/go-connection)
[![GoDoc](https://godoc.org/github.com/frederikhs/go-connection?status.svg)](https://godoc.org/github.com/frederikhs/go-connection)
[![Quality](https://goreportcard.com/badge/github.com/frederikhs/go-connection)](https://goreportcard.com/report/github.com/frederikhs/go-connection)
[![Test](https://github.com/frederikhs/go-connection/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/frederikhs/go-connection/actions/workflows/test.yml)

An opinionated small PostgreSQL client library based on [jmoiron/sqlx](https://github.com/jmoiron/sqlx)

### about
A postgres client with support for enforcing transactions and allowing the user to utilize transactions with the same interface as when not in a transaction.

### usage

`ConnectFromEnv` assumes environment variables:

```bash
DB_USER
DB_PASS
DB_HOST
DB_PORT
DB_DATABASE
```

to exist.

```go
package main

import (
    "github.com/frederikhs/go-connection"
)

type City struct {
    Name string
    County string
}

func main()  {
    conn := connection.ConnectFromEnv()

    var cities []City
    err := conn.Select(&cities, "SELECT name, country FROM cities WHERE country = $1", "Denmark")
    if err != nil {
       panic(err)
    }
	
    ...
}
```
