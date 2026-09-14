package config

import "os"

type Config struct {
	Port           string
	DataFile       string
	FrontendOrigin string
}

func Load() Config {
	return Config{
		Port:           get("PORT", "8080"),
		DataFile:       get("DATA_FILE", "data/playlists.json"),
		FrontendOrigin: get("FRONTEND_ORIGIN", "*"),
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
