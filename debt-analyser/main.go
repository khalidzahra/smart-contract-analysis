package main

import (
	"encoding/json"
	"fmt"
	"github.com/khalidzahra/debt-analyzer/util"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ContractData struct {
	Language string `json:"language"`
	Sources  map[string]struct {
		Content string `json:"content"`
	} `json:"sources"`
	Settings struct {
		Optimizer struct {
			Enabled bool `json:"enabled"`
			Runs    int  `json:"runs"`
		} `json:"optimizer"`
		OutputSelection map[string]map[string][]string `json:"outputSelection"`
		Metadata        struct {
			UseLiteralContent bool `json:"useLiteralContent"`
		} `json:"metadata"`
		Libraries map[string]interface{} `json:"libraries"`
	} `json:"settings"`
}

type fileVersion struct {
	path    string
	version int
}

var start bool = true

func main() {
	counter := 0
	startDir := path.Join("versioned-smart-contracts", "mainnet")
	err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			counter++
			err := readFilesInDirectory(path)
			if err != nil {
				return err
			}
			if counter == 5000 {
				util.SaveExcelFile("out_data/total_debt.xlsx")
				panic("DONE")
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directories: %v\n", err)
	}

	util.SaveExcelFile("out_data/total_debt.xlsx")
}

func readFilesInDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	var fileVersions []fileVersion
	var debtEvolution []int
	foundContract := false

	for _, file := range files {
		if !file.IsDir() {
			filePath := path.Join(dirPath, file.Name())
			if !start {
				splPath := strings.Split(filePath, string(filepath.Separator))
				contractName := splPath[len(splPath)-2]
				if contractName == "FB" {
					start = true
				} else {
					fmt.Printf("Skipping contract %s\n", contractName)
				}
			}
			foundContract = true
			version, err := util.ExtractVersion(file.Name())
			if err != nil {
				fmt.Printf("Error extracting version for file %s: %v\n", file.Name(), err)
				continue
			}
			fileVersions = append(fileVersions, fileVersion{path: filePath, version: version})
		}
	}

	// Sort based on version
	sort.Slice(fileVersions, func(i, j int) bool {
		return fileVersions[i].version < fileVersions[j].version
	})

	for _, fv := range fileVersions {
		fmt.Println("Processing file:", fv.path)
		debt := readFile(fv.path, start)
		debtEvolution = append(debtEvolution, debt)
	}

	if foundContract {
		pathSplit := strings.Split(dirPath, string(filepath.Separator))
		contractName := pathSplit[len(pathSplit)-2]
		contractDeployer := pathSplit[len(pathSplit)-1]
		if start {
			util.ExportTotalDebtToExcel(contractDeployer, contractName, debtEvolution)
		}
	}

	return nil
}

func readFile(filePath string, save bool) int {
	totalDebt := 0

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return 0
	}

	contentStr := string(content)
	contentStr = strings.TrimSpace(contentStr)

	var comments []string

	// Multi-contract file
	if strings.HasPrefix(contentStr, "{{") && strings.HasSuffix(contentStr, "}}") {
		fmt.Printf("Multi-contract file found: %s\n", filePath)
		contentStr = contentStr[1 : len(contentStr)-1] // Remove single brace
		var contractData ContractData
		err = json.Unmarshal([]byte(contentStr), &contractData)
		if err != nil {
			fmt.Printf("Error unmarshalling contract data: %v\n", err)
			return 0
		}
		for _, v := range contractData.Sources {
			contractComments := findComments(v.Content)
			comments = append(comments, contractComments...)
			totalDebt += len(contractComments)
		}
	} else {
		fmt.Println("Normal contract")
		comments = findComments(contentStr)
		totalDebt += len(comments)
	}

	splPath := strings.Split(filePath[:len(filePath)-4], string(filepath.Separator))
	contractNameVersion := splPath[len(splPath)-1]
	versionIndex := util.GetVersionIndex(contractNameVersion)
	contractName := contractNameVersion[:versionIndex-1]
	contractVersionStr := contractNameVersion[versionIndex:]
	contractVersion, err := strconv.Atoi(contractVersionStr)
	if save {
		go util.ExportCommentsToExcel(contractName, contractVersion, comments)
	}
	return totalDebt
}

func findComments(content string) []string {
	var debtComments []string

	// Regex for single-line comments
	singleLineRegex, err := regexp.Compile(`//.*`)
	if err != nil {
		fmt.Printf("Error compiling single line regex: %v\n", err)
		return nil
	}

	// Regex for multi-line comments
	multiLineRegex, err := regexp.Compile(`/\*[\s\S]*?\*/`)
	if err != nil {
		fmt.Printf("Error compiling multi line regex: %v\n", err)
		return nil
	}

	// Match all comments in provided contract content
	singleLineComments := singleLineRegex.FindAllString(content, -1)
	multiLineComments := multiLineRegex.FindAllString(content, -1)

	debtKeywords := []string{" todo: ", " todo ", " fix ", " fix: ", " fixme ", " fixme: ", "legacy", "deprecated",
		"refactor", "temporary", " temp ", "hack", "workaround",
		"work around", " wip ", "work in progress", "enhancement", "improvement"}

	// Check single-line comments for TODOs
	for _, comment := range singleLineComments {
		comment = strings.ToLower(comment)
		for _, debtKeyword := range debtKeywords {
			if strings.Contains(comment, debtKeyword) {
				debtComments = append(debtComments, comment)
			}
		}
	}

	// Check multi-line comments for TODOs
	for _, comment := range multiLineComments {
		comment = strings.ToLower(comment)
		for _, debtKeyword := range debtKeywords {
			if strings.Contains(comment, debtKeyword) {
				debtComments = append(debtComments, comment)
			}
		}
	}

	return debtComments
}
