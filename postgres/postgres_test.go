package postgres_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mortezadadgar/ecommerce-api/postgres"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pressly/goose/v3"
)

const (
	PGPassword = "secret"
	PGTestDB   = "testdb"
	PGUser     = "postgres"

	PGVersion = "15.3-alpine"
	PGImage   = "postgres"
)

var (
	// testDB is used for making databases for MustConnectDB().
	testDB *pgxpool.Pool

	// port holds the port returned from dockertest.
	port string
)

func formatDSN(dbName string) string {
	return fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable",
		PGUser,
		PGPassword,
		port,
		dbName,
	)
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: PGImage,
		Tag:        PGVersion,
		Env: []string{
			"POSTGRES_PASSWORD=" + PGPassword,
			"POSTGRES_DB=" + PGTestDB,
			"POSTGRES_USER=" + PGUser,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	port = resource.GetPort("5432/tcp")
	dsn := formatDSN(PGTestDB)

	err = resource.Expire(10)
	if err != nil {
		log.Fatalf("Could not set expire: %s", err)
	}

	pool.MaxWait = 10 * time.Second
	if err = pool.Retry(func() error {
		pg, err := postgres.New(dsn)
		if err != nil {
			return err
		}
		testDB = pg.DB
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	err = pool.Purge(resource)
	if err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

const migrationsDir = "../cmd/goose/migrations"

func runMigrations(dsn string) error {
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return err
	}

	err = goose.Up(db, migrationsDir)
	if err != nil {
		return err
	}

	return nil
}

func newTestDB(t *testing.T, dbName string) *pgxpool.Pool {
	t.Helper()

	query := `
	CREATE DATABASE ` + dbName + `
	`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := testDB.Exec(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	dsn := formatDSN(dbName)

	pg, err := postgres.New(dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = runMigrations(dsn)
	if err != nil {
		t.Fatal(err)
	}

	return pg.DB
}
