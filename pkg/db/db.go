package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Conn *gorm.DB

func Init(DBUser string, DBPassword string, DBHost string, DBName string) error {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", DBHost, DBUser, DBPassword, DBName)

	var err error

	Conn, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database init error")
	}

	return nil
}

func GetDBConn() *gorm.DB {
	return Conn
}
