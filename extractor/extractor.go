package extractor

type Extractor interface {
	loadAPIKey()
	findContractSource(string) string
}
