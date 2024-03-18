package extractor

type Extractor interface {
	loadAPIKey()
	findContractSource(string) string
}

type ContractSourceResponse struct {
	Status  int    `json:"status"`
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
