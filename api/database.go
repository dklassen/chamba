package api

// global variable to share it between main and the HTTP handler
import (
	"fmt"
	"log"

	_ "github.com/dklassen/chamba/Godeps/_workspace/src/github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

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

func databaseConnectionString() string {
	mysqlConfig := GetMysqlConfigs()
	// NOTE:: You are going to run into this in the future
	// related to https://github.com/go-sql-driver/mysql/issues/9
	// gist a table that has datetime fields can't be read as a time.Time useing Scan()
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database)
}

func connectToDatabase() *gorm.DB {
	db, err := gorm.Open("mysql", databaseConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	verifyDatabaseConnection(&db)
	// FIXME:: DO not automigrate
	db.AutoMigrate(&User{}, &AuthToken{})
	return &db
}
