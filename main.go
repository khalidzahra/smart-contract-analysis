package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/khalidzahra/smart-contract-analysis/extractor"
)

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	loadEnvVars()

	extractor := extractor.CreateMainNetExtractor()
	extractor.TraverseDataset(extractor.MatchContracts)
}
