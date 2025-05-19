package config

import (
	"log"
	"os"
)

type Config struct {
	FrontendOrigin string
	RedisAddr      string
	DBHost         string
	DBUser         string
	DBPass         string
	DBName         string
	CertFile       string
	KeyFile        string
	Port           string
}

func Load() *Config {
	cfg := &Config{
		FrontendOrigin: getenv("FRONTEND_ORIGIN", "*"),
		RedisAddr:      getenv("REDIS_ADDR", "localhost:6379"),
		DBHost:         getenv("DB_HOST", "localhost"),
		DBUser:         getenv("DB_USER", "chat"),
		DBPass:         getenv("DB_PASSWORD", "chat"),
		DBName:         getenv("DB_NAME", "chat"),
		CertFile:       getenv("TLS_CERT", "cert.pem"),
		KeyFile:        getenv("TLS_KEY", "key.pem"),
		Port:           getenv("PORT", "8443"),
	}
	log.Printf("[config] %+v\n", cfg)
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
