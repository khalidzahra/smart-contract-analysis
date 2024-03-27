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
	extractor.ApiURL = os.Getenv("MAIN_NET_URL")
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
	addressSet := make(map[string]bool)

	// Load already known addresses for contract if previously encountered
	dirPath := fmt.Sprintf("%s/mainnet/%s", extractor.OutPath, ogProps.ContractName)
	_, err = os.Stat(dirPath)
	if err == nil {
		dirEntries, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Println("Error opening directory:", err)
			return
		}

		for _, entry := range dirEntries {
			fileName := entry.Name()
			parts := strings.Split(fileName, "_")
			if len(parts) < 2 {
				continue
			}
			loadedAddress := parts[0]
			addressSet[loadedAddress] = true
			version++
		}
	}

	for _, transaction := range transactions {
		_, okTo := addressSet[transaction.To]
		_, okCtrctAddr := addressSet[transaction.ContractAddress]
		if transaction.To == deployer || (okTo && okCtrctAddr) { // Skip self transactions and transactions to already known contracts
			logging.Logger.Printf("Skipping, already seen %s", transaction.To)
			continue
		}

		addressSet[transaction.To] = true              // Store viewed contract address
		addressSet[transaction.ContractAddress] = true // Store viewed contract address

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
