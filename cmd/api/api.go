package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/xumak-grid/bedrock/http"
)

func main() {
	log := logrus.New()
	checkEnvVar(log)
	server := http.NewServer(log)
	server.Open()
}

// checkEnvVar checks critical environment variables and exits if one is not present
func checkEnvVar(log *logrus.Logger) {
	if os.Getenv("VAULT_ADDR") == "" {
		log.Fatalf("VAULT_ADDR is not set and is required")
	}
	if os.Getenv("VAULT_TOKEN") == "" {
		log.Fatalf("VAULT_TOKEN is not set and is required")
	}
	if os.Getenv("GRID_EXTERNAL_DOMAIN") == "" {
		log.Fatalf("GRID_EXTERNAL_DOMAIN is not set and is required")
	}
}
