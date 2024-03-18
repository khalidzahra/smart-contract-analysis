package extractor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	payload := url.Values{}
	payload.Set("module", "contract")
	payload.Set("action", "getsourcecode")
	payload.Set("address", contractAddress)
	payload.Set("usernapikeyame", extractor.properties.EtherscanKey)

	r, err := http.NewRequest(http.MethodPost, extractor.properties.ApiURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return "", "", err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(r)
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
