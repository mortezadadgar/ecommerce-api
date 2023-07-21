# Ecommerce API
This is not a serious project by any means, just experimenting with golang to
get the best structures out of it; can be used as a template for further rest
api developments as well.

## Features
- [X] pluggable database implementation.
- [X] REST API written with chi router.
- [X] Stateful token authentication.
- [X] Automatic database migrations on deployment.
- [X] PostgresSQL database provider.
- [X] integration tests using dockertest.
- [X] swagger UI.
- [X] Dockerized.


## How to run locally
this application do not offer a auto migrations in every single run as it's
considered bad practice, in order to run migrartions a binary of goose should be build.
to run migrations locally:
```shell
$ make migrate
$ ./migrate up
```

to start the application on a port specified by `ADDRESS` in `.env` file (by
default `:8080`)
if there's a `make` command:
```shell
$ make run
```
note: `run` target by default starts the application using `-race` flag so it
might be a little slower; to use this application without any overhead you can
run the following code snippet.

if there's no `make` command:
```shell
go run ./cmd/ecommerce
```

## Deploy
personally I have not deployed this api yet.
in order to deploy it you have to export the password used by postgres:
```shell
export POSTGRES_PASSWORD="whatever"
```
and eventually:
```shell
docker compose up
```

## Roadmap
- [ ] imporve authentication - there's no scope or a admin user concepts.
- [ ] cache products results.
- [ ] write remaining integration tests for postgres.
- [ ] mongodb provider.
- [ ] grpc?
- [ ] frontend?
