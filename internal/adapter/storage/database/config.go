package database

type Config struct {
	DSN string `env:"DATABASE_STRING"`
}
