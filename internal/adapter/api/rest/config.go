package rest

// Config конфиг рест сервера.
type Config struct {
	Address   string `env:"REST_ADDRESS"`
	SSLEnable bool   `env:"SSL_ENABLE"`
}
