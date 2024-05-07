package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	sitter "github.com/khalidzahra/parser/go-tree-sitter"
	sol "github.com/khalidzahra/parser/go-tree-sitter/solidity"
)

type CSVFile struct {
	FilePath string
	Content  strings.Builder
}

func CreateCSVFile(filePath string) *CSVFile {
	return &CSVFile{
		FilePath: filePath,
	}
}

func (cf *CSVFile) Append(content string) {
	cf.Content.WriteString(content)
}

func (cf *CSVFile) SaveToFile() error {
	file, err := os.Create(cf.FilePath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close()
	_, err = file.Write([]byte(cf.Content.String()))
	if err != nil {
		return err
	}
	return nil
}

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

var outFile *CSVFile

var debtKeywords = []string{" todo: ", " todo ", " fix ", " fix: ", " fixme ", " fixme: ", "legacy", "deprecated",
	"refactor", "temporary", " temp ", "hack", "workaround",
	"work around", " wip ", "work in progress", "enhancement", "improvement"}

var lang = sol.GetLanguage()

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	startDir := path.Join("../../versioned-smart-contracts", "mainnet")
	err := filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		if info.IsDir() {
			err := readFilesInDirectory(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directories: %v\n", err)
	}
}

func readFilesInDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	var fileVersions []fileVersion
	var debtComments, debtCode []string
	var debtEvolution []int
	foundContract := false

	for _, currFile := range files {
		if !currFile.IsDir() {
			filePath := path.Join(dirPath, currFile.Name())
			foundContract = true
			version, err := ExtractVersion(currFile.Name())
			if err != nil {
				fmt.Printf("Error extracting version for currFile %s: %v\n", currFile.Name(), err)
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
		fmt.Println("Processing currFile:", fv.path)
		localDebtComments, localDebtCode := readFile(fv.path)
		debtScore := 0
		debtScore += len(localDebtComments) - len(debtComments)
		debtComments = append(debtComments, localDebtComments...)
		debtCode = append(debtCode, localDebtCode...)
	}

	if foundContract {
		pathSplit := strings.Split(dirPath, string(filepath.Separator))
		contractName := pathSplit[len(pathSplit)-2]
		contractDeployer := pathSplit[len(pathSplit)-1]
		techDebtPrettified := strings.Trim(strings.Replace(fmt.Sprint(debtEvolution), " ", ",", -1), "[]")
		line := fmt.Sprintf("%s,%s,%s\n", contractDeployer, contractName, techDebtPrettified)
		fmt.Println(line)
	}

	return nil
}

func readFile(filePath string) ([]string, []string) {
	fmt.Println(filePath)

	debtComments := []string{}
	debtCode := []string{}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, nil
	}

	contentStr := string(content)
	contentStr = strings.TrimSpace(contentStr)

	//var comments []string

	// Multi-contract file
	if strings.HasPrefix(contentStr, "{{") && strings.HasSuffix(contentStr, "}}") {
		fmt.Printf("Multi-contract file found: %s\n", filePath)
		contentStr = contentStr[1 : len(contentStr)-1] // Remove single brace
		var contractData ContractData
		err = json.Unmarshal([]byte(contentStr), &contractData)
		if err != nil {
			fmt.Printf("Error unmarshalling contract data: %v\n", err)
			return nil, nil
		}
		for _, v := range contractData.Sources {
			findDebt(v.Content)
		}
	} else {
		fmt.Println("Normal contract")
		findDebt(contentStr)
	}

	//splPath := strings.Split(strings.ReplaceAll(filePath, string(filepath.Separator), "/"), "/")
	//contractNameVersion := splPath[len(splPath)-1]
	//contractVersion, err := ExtractVersion(contractNameVersion)
	// ctrct_name/deployer/ctrct_file
	//contractName := splPath[len(splPath)-3] + "_" + splPath[len(splPath)-2]
	//go util.ExportCommentsToExcel(contractName, contractVersion, comments)
	return debtComments, debtCode
}

func findDebt(contractSource string) ([]string, []string) {
	sourceCode := []byte(contractSource)

	debtComments := make([]string, 0)
	debtCode := make([]string, 0)

	queryPattern := `(
		(comment) @comment
		.
		(function_definition) @func-def
	)`

	n, _ := sitter.ParseCtx(context.Background(), sourceCode, lang)
	query, _ := sitter.NewQuery([]byte(queryPattern), lang)

	qc := sitter.NewQueryCursor()
	qc.Exec(query, n)

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		comment := m.Captures[0].Node.Content(sourceCode)
		fBody := m.Captures[1].Node.Content(sourceCode)

		for _, k := range debtKeywords {
			if strings.Contains(comment, k) {
				debtComments = append(debtComments, comment)
				debtCode = append(debtCode, fBody)
				fmt.Println(comment)
				fmt.Println(fBody)
			}
		}
	}

	return debtComments, debtCode
}

func ExtractVersion(filename string) (int, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return 0, fmt.Errorf("filename format not recognized")
	}
	versionPart := parts[len(parts)-1] // Get the last part which should be "V%version%.sol"
	versionStr := strings.TrimPrefix(versionPart, "V")
	versionStr = strings.TrimSuffix(versionStr, ".sol")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("invalid version number")
	}

	return version, nil
}
