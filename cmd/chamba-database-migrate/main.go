package main

import "github.com/dklassen/chamba/api"

// chamba-database command is a cmd for controlling the database
// at the moment we simply run migrations on the database that is specified by
// DATABASE_URL.

func main() {
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
}
