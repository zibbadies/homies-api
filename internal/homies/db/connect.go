package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/zibbadies/homies/internal/homies/config"
	_ "github.com/lib/pq"
)

// Internal Variables
var (
	db *sql.DB
)

func ConnectDatabase() error {
    psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	newdb, err := sql.Open("postgres", psqlconn)
	db = newdb;
	if (err != nil) {
		return err;
	}

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(10 * time.Minute)

	err = db.Ping()
	if (err != nil) {
		return err;
	}

	return nil;
}

func CheckConnection() error {
	return db.Ping();
}
