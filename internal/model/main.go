package model

import (
	"github.com/jmoiron/sqlx"
)

var (
	Environment string
)

// InitDB instantiates a DB connection
func InitDB() *sqlx.DB {
	return DBX()
}
