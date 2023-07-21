swagger:
	@swag init --quiet --output swagger --generalInfo http/http.go
run: swagger
	@go run -race ./cmd/ecommerce/
build: swagger
	@go build ./cmd/ecommerce
test:
	@go test ./...
migrate:
	go build -o migrate ./cmd/goose
