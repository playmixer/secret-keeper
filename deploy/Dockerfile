FROM golang:1.22.3 as build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server/ ./cmd/server/
COPY docs ./docs
COPY internal ./internal
COPY pkg ./pkg

# Build
WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cmd/server

FROM ubuntu:latest

WORKDIR /app

COPY --from=build /app/cmd/server/server /server

EXPOSE 8080 8443

# Run
CMD ["/server"]