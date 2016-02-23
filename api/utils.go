package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var (
	goenv         string
	configuration *Configuration
)

// MysqlConfig struct holds the connection information for setting
// up
type MysqlConfig struct {
	Username string
	Password string
	Database string
	Host     string
	Port     string
}

// Configuration struct holds all the information related to
// properly configuring the application
type Configuration struct {
	MysqlConfig MysqlConfig
}

func environment() string {
	if goenv == "" {
		goenv = os.Getenv("GOENV")
	}

	if goenv == "" {
		log.Fatal("GOENV environment variable is not set. Please set and re-launch")
	}
	return goenv
}

func getConfigFile() (file *os.File) {
	fileName := fmt.Sprintf("%s/src/github.com/dklassen/chamba/config/sources.%s.json", os.Getenv("GOPATH"), environment())
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("The config file %s for the environment %s was not found", fileName, environment())
	}
	return
}

// GetMysqlConfigs load the configuration file and return the Mysql config information
// for the environment
func GetMysqlConfigs() MysqlConfig {
	configuration := loadConfiguration()
	return configuration.MysqlConfig
}

func checkConfigIsNotEmpty(config Configuration) {
	empty := Configuration{}
	if config == empty {
		log.Fatal("No configuration loaded")
	}
}

func loadConfiguration() Configuration {
	if configuration == nil {
		file := getConfigFile()
		decoder := json.NewDecoder(file)
		configuration = &Configuration{}
		err := decoder.Decode(&configuration)
		if err != nil {
			log.Fatal(err)
		}

		// I don't know what to do should we bail hard or not?
		// i feel like we should bail hard
		checkConfigIsNotEmpty(*configuration)
	}

	return *configuration
}
