package main

import (
	"log"
	"os"
	"path/filepath"

	"repo-doc/cmd"

	"github.com/joho/godotenv"
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	envPath := filepath.Join(cwd, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found in project root")
	}
}

func main() {
	cmd.Execute()
}
