package main

import (
	"database/sql"
	"embed"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		return
	}

	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	command := args[0]
	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[2:]...)
	}

	err = goose.Run(command, db, "migrations", arguments...)
	if err != nil {
		log.Fatal(err)
	}
}
