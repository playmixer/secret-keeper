go run -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git show --oneline -s)'" ./cmd/client/client.go