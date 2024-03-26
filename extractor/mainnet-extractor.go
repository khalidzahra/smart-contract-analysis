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

	version := 0
	for _, transaction := range transactions {
		if transaction.To == deployer { // Skip self transactions
			continue
		}

		logging.Logger.Printf("Transaction Finding properties for contract with address %s", transaction.To)
		var targetAddress string

		if transaction.To == "" {
			targetAddress = transaction.ContractAddress
		} else {
			targetAddress = transaction.To
		}

		foundProps, err := extractor.FindContractProperties(targetAddress)

		if err != nil {
			logging.Logger.Fatal(err)
			continue
		}

		if foundProps.SourceCode == "" { // Not a contract or sourcecode not verified
			continue
		}

		// contractABI, err := abi.JSON(strings.NewReader(string(foundProps.ABI)))
		// if err != nil {
		// 	logging.Logger.Fatal(err)
		// 	return
		// }

		// // Decode input data using contract's ABI
		// method, err := contractABI.MethodById([]byte(transaction.Input))
		// if err == nil {
		// 	logging.Logger.Println("Method Name:", method.Name)
		// 	return
		// }

		if foundProps.ContractName != ogProps.ContractName { // Only interested in contracts with the same name
			continue
		}

		logging.Logger.Printf("Contract Name %s", foundProps.ContractName)
		logging.Logger.Printf("Finding source for contract with address %s", targetAddress)
		name, source, err := extractor.FindContractSource(targetAddress)

		if err != nil {
			logging.Logger.Fatal(err)

		} else {
			if len(source) > 0 {
				outPath := fmt.Sprintf("%s/mainnet/%s/%s_%s_V%d.sol", extractor.OutPath, name, targetAddress, name, version)
				version++
				if err := os.MkdirAll(filepath.Dir(outPath), 0770); err != nil {
					logging.Logger.Fatal(err)
				}
				os.WriteFile(outPath, []byte(source), 0644)
			}
		}
	}
}

func (extractor *MainNetExtractor) TraverseDataset(fn runnable) {
	err := filepath.Walk(extractor.InputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logging.Logger.Fatal(err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sol") {
			meta := strings.Split(info.Name(), "_")
			address := fmt.Sprintf("0x%s", meta[0])
			fn(address)

			if err := os.Remove(path); err != nil {
				logging.Logger.Fatal(err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		logging.Logger.Fatal(err)
	}
}
