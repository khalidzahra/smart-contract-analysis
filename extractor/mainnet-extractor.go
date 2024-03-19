package extractor

import (
	"fmt"
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

	params := make(RequestParams)
	params["module"] = "contract"
	params["action"] = "getsourcecode"
	params["address"] = contractAddress
	params["userapikey"] = extractor.properties.EtherscanKey

	resBody := &ContractSourceResponse{}
	err := ExecuteRequest(extractor.properties.ApiURL, params, resBody)

	if err != nil {
		return "", "", err
	}

	if resBody.IsSuccessful() {
		return "", "", fmt.Errorf(resBody.Message)
	}

	if len(resBody.Result) == 0 {
		return "", "", fmt.Errorf("address has no contract source")
	}

	return resBody.Result[0].ContractName, resBody.Result[0].SourceCode, nil
}

func (extractor *MainNetExtractor) FindDeployerAddress(contractAddress string) (string, error) {
	params := make(RequestParams)
	params["module"] = "contract"
	params["action"] = "getcontractcreation"
	params["contractaddresses"] = contractAddress
	params["userapikey"] = extractor.properties.EtherscanKey

	resBody := &ContractDeployerResponse{}
	err := ExecuteRequest(extractor.properties.ApiURL, params, resBody)

	if err != nil {
		return "", err
	}

	if resBody.IsSuccessful() {
		return "", fmt.Errorf(resBody.Message)
	}

	if len(resBody.Result) == 0 {
		return "", fmt.Errorf("address has no contract source")
	}

	return resBody.Result[0].ContractCreator, nil
}
