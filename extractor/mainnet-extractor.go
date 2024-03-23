package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		DefaultExtractor: *CreateDefaultExtractor(),
		InputPath:        os.Getenv("INPUT_DATASET_PATH"),
		OutPath:          os.Getenv("OUTPUT_DATASET_PATH"),
	}
	extractor.properties.ApiURL = os.Getenv("MAIN_NET_URL")
	extractor.properties.EtherscanKey = os.Getenv("ETHERSCAN_API_KEY")
	return extractor
}

func (extractor *MainNetExtractor) MatchContracts(address string) {
	logging.Logger.Printf("Finding properties for contract with address %s", address)
	ogProps, err := extractor.FindContractProperties(address)
	if err != nil {
		logging.Logger.Fatal(err)
		return
	}

	logging.Logger.Printf("Finding deployer for contract with address %s", address)
	deployer, err := extractor.FindDeployerAddress(address)
	if err != nil {
		logging.Logger.Fatal(err)
		return
	}

	logging.Logger.Printf("Finding all transactions for address %s", deployer)
	transactions, err := extractor.FindAllTransactions(deployer)
	if err != nil {
		logging.Logger.Fatal(err)
		return
	}

	logging.Logger.Printf("Finding creation transactions for %s", deployer)
	creationTrans := extractor.FindCreationTransactions(transactions)
	version := 0
	for _, transaction := range creationTrans {
		logging.Logger.Printf("Finding properties for contract with address %s", transaction.ContractAddress)
		foundProps, err := extractor.FindContractProperties(transaction.ContractAddress)

		if err != nil {
			logging.Logger.Fatal(err)
			continue
		}

		if foundProps.ContractName != ogProps.ContractName { // Only interested in contracts with the same name
			continue
		}

		name, source, err := extractor.FindContractSource(transaction.ContractAddress)
		if err != nil {
			panic(err)
		} else {
			if len(source) > 0 {
				outPath := fmt.Sprintf("%s/mainnet/%s/%s_%s_V%d.sol", extractor.OutPath, name, transaction.ContractAddress, name, version)
				version++
				if err := os.MkdirAll(filepath.Dir(outPath), 0770); err != nil {
					panic(err)
				}
				os.WriteFile(outPath, []byte(source), 0644)
			}
		}
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
