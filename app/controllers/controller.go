package controllers

import (
	"database/sql"
)

type Controller struct {
	DB *sql.DB // or whatever your database type is
	// Add other dependencies as needed (mailer, logger, etc.)
}
