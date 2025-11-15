package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	dbx  *sqlx.DB
	once sync.Once
)

// DBX is a singleton that gets always the same instance of sqlx.DB
// in a both thread and nil safe fashion.
func DBX() *sqlx.DB {
	once.Do(func() {
		maxOpenConnections := viper.GetInt("db-max-open-connections")
		maxIdleTime := viper.GetInt("db-max-idle-time")

		log.Printf("Init DB (%s), max open connections: %d, max idle time (min): %d",
			Environment, maxOpenConnections, maxIdleTime)

		connString := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
			viper.GetString("db.user"),
			viper.GetString("db.password"),
			viper.GetString("db.host"),
			viper.GetInt("db.port"),
			viper.GetString("db.name"),
			viper.GetString("db.ssl-mode"))

		var err error
		dbx, err = sqlx.Connect("postgres", connString)
		if err != nil {
			log.Fatal("couldn't establish a DB connection. Bailing out. err: ", err.Error())
		}

		dbx.SetMaxOpenConns(maxOpenConnections)
		dbx.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Minute)
	})

	return dbx
}
