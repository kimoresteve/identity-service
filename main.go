package main

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // MySQL driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kimoresteve/identity-service/app/controllers"
	"github.com/kimoresteve/identity-service/app/database"
	subroute "github.com/kimoresteve/identity-service/app/routes"
	_ "github.com/kimoresteve/identity-service/docs"
	"log"
	"path/filepath"
	"runtime"
)

func main() {

	dbInstance := database.GetDBConnection()
	driver, err := mysql.WithInstance(dbInstance, &mysql.Config{})
	if err != nil {

		panic(err)
	}

	migrationPath := filepath.Join(GetRootPath(), "migrations")
	//log.Printf("Looking for migrations in: %s", migrationPath)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"mysql",
		driver,
	)
	if err != nil {
		log.Printf("migration setup error %s ", err.Error())
	}

	// Handle "no change" gracefully
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("migration error %s ", err.Error())
	} else if err == migrate.ErrNoChange {
		log.Println("No database changes needed")
	} else {
		log.Println("Migration completed successfully")

	}

	router := &subroute.App{}

	router.Controller = &controllers.Controller{
		DB: dbInstance,
	}

	router.Initialize()
	router.Run()

	//http.Handle("/swagger/", http.StripPrefix("/swagger/", httpSwagger.WrapHandler))

	log.Println("This message won't appear if server runs successfully")

}

func GetRootPath() string {

	_, b, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(b), "./")
}
