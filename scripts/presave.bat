gofmt .\internal\.. .\pkg\.. .\cmd\..
goimports -local "github.com/playmixer/secret-keeper" -w .\internal\.. .\pkg\.. .\cmd\..
go mod tidy
go test ./...