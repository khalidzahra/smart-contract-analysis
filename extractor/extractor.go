package extractor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type RequestParams map[string]string

type ExtractorProperties struct {
	EtherscanKey string
	ApiURL       string
}

type Extractor interface {
	FindContractSource(string) (string, string, error)
	FindDeployerAddress(string) (string, error)
}

type DefaultExtractor struct {
	properties ExtractorProperties
}

type ExtractorResponse interface {
	IsSuccessful() bool
}

type ContractSourceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []struct {
		SourceCode       string `json:"SourceCode"`
		ABI              string `json:"ABI"`
		ContractName     string `json:"ContractName"`
		CompilerVersion  string `json:"CompilerVersion"`
		OptimizationUsed string `json:"OptimizationUsed"`
		Runs             string `json:"Runs"`
		ConstructorArgs  string `json:"ConstructorArguments"`
		EVMVersion       string `json:"EVMVersion"`
		Library          string `json:"Library"`
		LicenseType      string `json:"LicenseType"`
		Proxy            string `json:"Proxy"`
		Implementation   string `json:"Implementation"`
	} `json:"result"`
}

type ContractDeployerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []struct {
		ContractAddress string `json:"contractAddress"`
		ContractCreator string `json:"contractCreator"`
		TxHash          string `json:"txHash"`
	} `json:"result"`
}

func ExecuteRequest(requestURL string, params RequestParams, extractorRes ExtractorResponse) error {
	payload := url.Values{}

	for k, v := range params {
		payload.Set(k, v)
	}

	r, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(extractorRes); err != nil {
		return err
	}
	return nil
}

func (res *ContractSourceResponse) IsSuccessful() bool {
	return res.Message == "1"
}

func (res *ContractDeployerResponse) IsSuccessful() bool {
	return res.Message == "1"
}

func (extractor *DefaultExtractor) FindContractSource(contractAddress string) (string, string, error) {

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

func (extractor *DefaultExtractor) FindDeployerAddress(contractAddress string) (string, error) {
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
