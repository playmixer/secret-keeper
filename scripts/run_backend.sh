swag init -o ./docs -g ./internal/adapter/api/rest/rest.go
swag fmt
go run ./cmd/server/server.go