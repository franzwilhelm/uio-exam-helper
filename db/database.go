package db

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Default is the connection reference
var Default *gorm.DB

func ConnectPostgres(dialect, connString string) (err error) {
	if Default != nil {
		return nil
	}
	Default, err = gorm.Open(dialect, connString)
	return
}

func init() {
	connString := os.Getenv("DB_CONN_STRING")
	if connString == "" {
		connString = "host=localhost port=5432 sslmode=disable dbname=uio_exam_helper user=uio password=exam-helper"
		log.
			WithField("connString", connString).
			Debug("DB_CONN_STRING env empty, falling back to default")
	}
	if err := ConnectPostgres("postgres", connString); err != nil {
		panic(err.Error())
	}
	log.Info("Successfully connected to the database")
	Default.LogMode(false)
}
