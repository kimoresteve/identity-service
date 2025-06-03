package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}
}

func GetDBConnection() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&multiStatements=true", username, password, host, port, dbname, "utf8")

	Db, err := sql.Open("mysql", dbURI)
	if err != nil {

		panic(err)
	}

	err = Db.Ping()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		panic(err)
	}

	return Db

}
