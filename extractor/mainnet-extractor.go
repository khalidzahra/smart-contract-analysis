package extractor

import (
	"os"
)

type MainNetExtractor struct {
	DefaultExtractor
}

func CreateMainNetExtractor() Extractor {
	extractor := &MainNetExtractor{}
	extractor.properties.ApiURL = os.Getenv("MAIN_NET_URL")
	extractor.properties.EtherscanKey = os.Getenv("ETHERSCAN_API_KEY")
	return extractor
}
