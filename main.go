package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	test_address := "0x180012500db77132f3da5d00de0e96ef614697e5"
	extractor := extractor.CreateMainNetExtractor()
	name, source, err := extractor.FindContractSource(test_address)

	if err != nil {
		panic(err)
	} else {
		os.WriteFile(fmt.Sprintf("./%s.sol", name), []byte(source), 0644)
	}

	deployer, err := extractor.FindDeployerAddress(test_address)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(deployer)
	}

	time.Sleep(5 * time.Second)

	transactions, err := extractor.FindAllTransactions(deployer)
	if err != nil {
		panic(err)
	} else {
		for _, transaction := range transactions {
			fmt.Printf("Transaction from %s to %s\n", transaction.From, transaction.To)
		}
	}
}
