package extractor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type MainNetExtractor struct {
	properties ExtractorProperties
}

func CreateMainNetExtractor() Extractor {
	extractor := &MainNetExtractor{}
	extractor.properties.ApiURL = os.Getenv("MAIN_NET_URL")
	extractor.properties.EtherscanKey = os.Getenv("ETHERSCAN_API_KEY")
	return extractor
}

func (extractor *MainNetExtractor) FindContractSource(contractAddress string) (string, string, error) {
	url := fmt.Sprintf("%s?module=contract&action=getsourcecode&address=%s&apikey=%s",
		extractor.properties.ApiURL,
		contractAddress,
		extractor.properties.EtherscanKey)

	res, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	var resBody ContractSourceResponse
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return "", "", err
	}

	if resBody.Status != "1" {
		return "", "", fmt.Errorf(resBody.Message)
	}

	if len(resBody.Result) == 0 {
		return "", "", fmt.Errorf("address has no contract source")
	}

	return resBody.Result[0].ContractName, resBody.Result[0].SourceCode, nil
}
