package config

import "os"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	DatabaseURL       string
	BunExecutorURL    string
	PythonExecutorURL string
	Port              string
}

// Load reads configuration from environment variables, falling back to sensible
// defaults for local development.
func Load() Config {
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://runeforge:runeforge@localhost:5432/runeforge"),
		BunExecutorURL:    getEnv("BUN_EXECUTOR_URL", "http://localhost:8081"),
		PythonExecutorURL: getEnv("PYTHON_EXECUTOR_URL", "http://localhost:8082"),
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
