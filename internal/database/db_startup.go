package database

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Database struct {
	Client *sqlx.DB
}

func NewDatabase() (*Database, error) {
	log.Info("Establishing new database connection")

	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	//connectionString := "root:Rahuliv@2002@tcp(127.0.0.1:3306)/studentdb?parseTime=true"

	db, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		log.Errorf("Database connection failed: %v", err)
		return nil, fmt.Errorf("could not connect to the database: %w", err)
	}

	return &Database{
		Client: db,
	}, nil
}

func (d *Database) Ping(ctx context.Context) error {
	if err := d.Client.DB.PingContext(ctx); err != nil {
		log.Errorf("Database ping failed: %v", err)
		return err
	}
	log.Info("Database connection is active")
	return nil
}
