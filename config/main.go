package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/log"
)

var (
	JWT_SECRET string
)

func loadDB(require bool) {
	uri := os.Getenv("POSTGRES_URI")
	if uri == "" {
		if require {
			panic("Postgres URI (POSTGRES_URI) not defined, but required")
		}

		return
	}

	db.LoadPool(uri)
}

// Returns cleanup
func LoadBasics() func () error {
	godotenv.Load(".env")

	logCleanup, err := log.Init(os.Getenv("DISCORD_PREFIX"), os.Getenv("DISCORD_URL"))
	if err != nil {
		panic("Can't load log: " + err.Error())
	}

	JWT_SECRET = os.Getenv("JWT_SECRET")
	if JWT_SECRET == "" {
		panic("JWT_SECRET is absolutely required!")
	}

	loadDB(true)

	return func() error {
		err := logCleanup()
		fmt.Println("Bad Log Cleanup", err)
		db.Close()

		return err
	}
}

func LoadForCLI(withDB bool) func () {
	godotenv.Load(".env")

	loadDB(withDB)

	return func () {
		db.Close()
	}
}

func LoadForTests() {
	loadDB(false)
}
