package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dklassen/chamba/api"
)

// chamba-database command is a cmd for controlling the database
// at the moment we simply run migrations on the database that is specified by
// DATABASE_URL.
func main() {
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
	duration := finishedAt.sub(startedAt)
	log.Info("Finished database migration")
	log.WithFields(log.Fields{
		"started_at":  startedAt,
		"finished_at": finishedAt,
		"took":        duration.Seconds()}).Info("Finished database migration")
}
