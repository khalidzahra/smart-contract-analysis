package util

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
)

func ExportCommentsToExcel(contractName string, contractVersion int, comments []string) {
	f := excelize.NewFile()

	sheetName := "Sheet1"

	fmt.Println()

	for i, comment := range comments {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		f.SetCellValue(sheetName, cell, comment)
	}

	outDir := fmt.Sprintf("debt_data/contracts/%s", contractName)
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	err = f.SaveAs(fmt.Sprintf("%s/%d.xlsx", outDir, contractVersion))
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Printf("Saved debt comments for %s.\n", contractName)
}

func GetVersionIndex(name string) int {
	for i := range len(name) {
		if name[len(name)-1-i] == '_' {
			return len(name) - i
		}
	}
	return -1 // should never happen
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
