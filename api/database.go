package api

// global variable to share it between main and the HTTP handler
import (
	"log"
	"os"

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
}

func connectToDatabase() *gorm.DB {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	verifyDatabaseConnection(&db)
	return &db
}
