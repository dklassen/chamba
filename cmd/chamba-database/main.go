package main

import (
	"database/sql"
	"flag"
	"fmt"
	api "github.com/dklassen/chamba/api"
	"log"
	"os"
	"os/user"
)

var (
	usage = fmt.Sprintf(`Usage:  Valid commands: 
		-environment: set the environment you are setting the database for 
		-mysql_user: admin user to set up the database on your computer
		-mysql_pass: admin password to setup the database on your computer`)
	environment string
	mysql_user  string
	mysql_pass  string
)

func main() {
	current_user, err := user.Current()
	if err != nil {
		log.Println("Unable to grab default user name")
	}

	flag.StringVar(&environment, "environment", "", "Environment to setup")
	flag.StringVar(&mysql_user, "mysql_user", current_user.Username, "Mysql user for administrator")
	flag.StringVar(&mysql_pass, "mysql_pass", "", "Mysql password for administrator")
	flag.Parse()

	if environment == "" {
		log.Fatal("Require environment flag to be set")
	}

	log.Println("Setting environment:", environment)
	os.Setenv("GOENV", environment)
	connection_string := fmt.Sprintf("%s:%s:@tcp(127.0.0.1:3306)/?parseTime=true", mysql_user, mysql_pass)
	db, err := sql.Open("mysql", connection_string)
	if err != nil {
		log.Fatal("Something is wrong with the database connection", err)
	}
	setupDatabase(*db, environment)

}

func teardownDatabase(db sql.DB, environment string) {

}

func setupDatabase(db sql.DB, environment string) {
	mysqlConfig := api.GetMysqlConfigs()
	password := mysqlConfig.Password
	_, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS chamba_%s", environment))
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON chamba_%s.* TO 'chamba_user'@'localhost' IDENTIFIED BY '%s' WITH GRANT OPTION;", environment, password))
	if err != nil {
		log.Fatal(err)
	}

}
