package main

// TODO:
//     - users registration and login
//     - authentications (and only allow admin to make changes)
//     - mongodb
//     - cart (add, list, empty, total prices)
//     - order
//     - ...

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

	server.UsersStore = postgres.NewUsersStore(pg.DB)
	server.ProductsStore = postgres.NewProductsStore(pg.DB)
	server.CategoriesStore = postgres.NewCategoriesStore(pg.DB)

	server.Start()

	sig := registerSignalNotify()

	// wait for user signal
	<-sig

	err = close(server, pg)
	if err != nil {
		log.Println("failed to close program, exiting now...")
		os.Exit(1)
	}
}

func registerSignalNotify() chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	return sig
}

func close(server *http.Server, store http.Store) error {
	err := server.Close()
	if err != nil {
		return err
	}

	store.Close()

	return nil
}
