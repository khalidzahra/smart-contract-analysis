package extractor

import (
	"fmt"
	"os"
	"time"
)

type runnable func(string)

type MainNetExtractor struct {
	DefaultExtractor
	InputPath string
}

func CreateMainNetExtractor() Extractor {
	extractor := &MainNetExtractor{InputPath: os.Getenv("INPUT_DATASET_PATH")}
	extractor.properties.ApiURL = os.Getenv("MAIN_NET_URL")
	extractor.properties.EtherscanKey = os.Getenv("ETHERSCAN_API_KEY")
	return extractor
}

func (extractor *MainNetExtractor) MatchContracts(address string) {
	deployer, err := extractor.FindDeployerAddress(address)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	transactions, err := extractor.FindAllTransactions(deployer)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	creationTrans := extractor.FindCreationTransactions(transactions)
	for i, transaction := range creationTrans {
		name, source, err := extractor.FindContractSource(transaction.ContractAddress)
		if err != nil {
			panic(err)
		} else {
			os.WriteFile(fmt.Sprintf("./%s/mainnet/%s_V%d.sol", extractor.InputPath, name, i), []byte(source), 0644)
		}
		time.Sleep(5 * time.Second)
	}
}

func (extractor *MainNetExtractor) TraverseDataset(fn runnable) {

}
