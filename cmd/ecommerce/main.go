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
	if err != nil {
		log.Fatal(err)
	}

	pg, err := postgres.New()
	if err != nil {
		log.Fatal(err)
	}

	server := http.New(pg)

	server.UsersStore = postgres.NewUserStore(pg.DB)
	server.ProductsStore = postgres.NewProductStore(pg.DB)
	server.CategoriesStore = postgres.NewCategoryStore(pg.DB)
	server.TokensStore = postgres.NewTokenStore(pg.DB)
	server.CartsStore = postgres.NewCartStore(pg.DB)
	server.SearchStore = postgres.NewSearchStore(pg.DB)

	server.Start()

	// wait for user signal
	<-registerSignalNotify()

	err = closeMain(server, pg)
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

func closeMain(server *http.Server, store http.Store) error {
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
