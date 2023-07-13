make_swagger:
	@swag init --quiet --output swagger --generalInfo http/server.go
run: make_swagger
	@go run -race ./cmd/ecommerce/
build: make_swagger
	@go build .
test:
	@go test ./...
make_migrate:
	go build -o migrate ./cmd/goose
