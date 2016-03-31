package api

// global variable to share it between main and the HTTP handler
import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // We are using postgres
)

var db *gorm.DB

// GetDB is a accessor for a shared db object
func GetDB() *gorm.DB {
	if db == nil {
		db = connectToDatabase()
	}
	return db
}

func verifyDatabaseConnection(db *gorm.DB) {
	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(false)

	log.Info("Connected to database")
}

func databaseConnectionString() (dbstring string) {
	dbstring = os.Getenv("DATABASE_URL")
	log.WithField("connectionString", dbstring).Info("Attempting to connect to database")
	return
}

func connectToDatabase() *gorm.DB {
	db, err := gorm.Open("postgres", databaseConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	verifyDatabaseConnection(&db)
	return &db
}
