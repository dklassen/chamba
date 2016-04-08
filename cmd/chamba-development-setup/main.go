package main

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func executeCommand(command string, ignore bool) string {
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.Output()

	if err != nil && ignore == false {
		log.Fatal(err.Error())
	}
	log.Println(string(out))
	return string(out)
}

func up() {
	log.Println("Bringing up Postgres container...")
	executeCommand("docker build --tag chamba-postgres --rm $GOPATH/src/github.com/dklassen/chamba/cmd/chamba-development-setup/", false)
	executeCommand("docker rm /chamba-postgres", true) // TODO:: FIgure out why we have to remove the existing killed container name
	executeCommand("docker run -p 5432:5432 --name chamba-postgres -d chamba-postgres", false)
	dockerIP := executeCommand("docker-machine ip default", true)
	dockerIP = strings.Replace(dockerIP, "\n", "", -1)
	log.Printf("To connect to database export DATABASE_URL=postgres://chamba_user@%s:5432/chamba?sslmode=disable\n", dockerIP)
}

func down() {
	log.Println("Killing Postgres container....")
	executeCommand("docker kill chamba-postgres", false)
}

func main() {
	actions := os.Args[1:]
	var action string
	if len(actions) == 1 {
		log.Println(actions)
		action = actions[0]
	}

	switch action {
	case "up":
		up()
	case "down":
		down()
	default:
		log.Fatal("Up or Down those are the only god dam options")
	}

}
