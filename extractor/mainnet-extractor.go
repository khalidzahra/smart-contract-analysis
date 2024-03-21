package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/khalidzahra/smart-contract-analysis/logging"
)

type runnable func(string)

type MainNetExtractor struct {
	DefaultExtractor
	InputPath string
	OutPath   string
}

func CreateMainNetExtractor() *MainNetExtractor {
	extractor := &MainNetExtractor{
		InputPath: os.Getenv("INPUT_DATASET_PATH"),
		OutPath:   os.Getenv("OUTPUT_DATASET_PATH"),
	}
	extractor.properties.ApiURL = os.Getenv("MAIN_NET_URL")
	extractor.properties.EtherscanKey = os.Getenv("ETHERSCAN_API_KEY")
	return extractor
}

func (extractor *MainNetExtractor) MatchContracts(address string) {
	logging.Logger.Printf("Finding properties for contract with address %s", address)
	ogProps, err := extractor.FindContractProperties(address)
	if err != nil {
		time.Sleep(5 * time.Second)
		return
	}

	time.Sleep(5 * time.Second)

	logging.Logger.Printf("Finding deployer for contract with address %s", address)
	deployer, err := extractor.FindDeployerAddress(address)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	logging.Logger.Printf("Finding all transactions for address %s", deployer)
	transactions, err := extractor.FindAllTransactions(deployer)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	logging.Logger.Printf("Finding creation transactions for %s", deployer)
	creationTrans := extractor.FindCreationTransactions(transactions)
	version := 0
	for _, transaction := range creationTrans {
		logging.Logger.Printf("Finding properties for contract with address %s", transaction.ContractAddress)
		foundProps, err := extractor.FindContractProperties(transaction.ContractAddress)

		if err != nil || foundProps.ContractName != ogProps.ContractName {
			fmt.Println(foundProps.ContractName, ogProps.ContractName)
			time.Sleep(5 * time.Second)
			continue
		}

		time.Sleep(5 * time.Second)

		name, source, err := extractor.FindContractSource(transaction.ContractAddress)
		if err != nil {
			panic(err)
		} else {
			if len(source) > 0 {
				outPath := fmt.Sprintf("%s/mainnet/%s_V%d.sol", extractor.OutPath, name, version)
				version++
				if err := os.MkdirAll(filepath.Dir(outPath), 0770); err != nil {
					panic(err)
				}
				os.WriteFile(outPath, []byte(source), 0644)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (extractor *MainNetExtractor) TraverseDataset(fn runnable) {
	err := filepath.Walk(extractor.InputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sol") {
			meta := strings.Split(info.Name(), "_")
			address := fmt.Sprintf("0x%s", meta[0])
			fn(address)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
