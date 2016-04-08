package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dklassen/chamba/api"
)

var (
	usage = fmt.Sprintf(`Usage: %s COMMAND
The flags available are a subset of the POSIX ones, but should behave similarly.
Valid commands:
 nuke  - Nuke the database and migrate back to ground zero
 migrate - Migrate the database to the latest schema
`, os.Args[0])
)

func nuke() {
	db := api.GetDB()
	db.LogMode(true)
	db.DropTableIfExists(
		&api.User{},
		&api.AuthToken{},
		&api.Address{},
		&api.Farm{})
	migrate()
}

func migrate() {
	startedAt := time.Now()
	log.Info("Starting database migration")
	db := api.GetDB()
	// TODO(dana) :: implement how to automagically discover these additional
	// structs perhaps have those structs inherit an interface or something?
	// TODO(dana) :: Add command line options
	db.LogMode(true)

	db.AutoMigrate(
		&api.User{},
		&api.AuthToken{},
		&api.Address{},
		&api.Farm{})

	finishedAt := time.Now()
	duration := finishedAt.Sub(startedAt)
	log.WithFields(log.Fields{
		"started_at":  startedAt,
		"finished_at": finishedAt,
		"took":        duration.Seconds()}).Info("Finished database migration")
}

// chamba-database command is a cmd for controlling the database
// at the moment we simply run migrations on the database that is specified by
// DATABASE_URL.
func main() {
	if len(os.Args) == 1 {
		log.Fatal(usage)
	}

	command := os.Args[1]
	switch command {
	case "nuke":
		nuke()
	case "migrate":
		migrate()
	default:
		log.Fatal(usage)
	}
}
