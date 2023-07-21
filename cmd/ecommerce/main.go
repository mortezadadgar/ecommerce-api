// Package main ties all application dependencies and execute it.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mortezadadgar/ecommerce-api/http"
	"github.com/mortezadadgar/ecommerce-api/postgres"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	pg, err := postgres.New(os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	server := http.New(pg)

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}

	// wait for user signal
	<-registerSignalNotify()

	err = closeMain(server, &pg)
	if err != nil {
		log.Println("failed to close program, exiting now...")
		os.Exit(1)
	}
}

func registerSignalNotify() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	return sig
}

type store interface {
	Close() error
}

type server interface {
	Start() error
	Close() error
}

func closeMain(server server, store store) error {
	err := server.Close()
	if err != nil {
		return err
	}

	err = store.Close()
	if err != nil {
		return err
	}

	return nil
}
