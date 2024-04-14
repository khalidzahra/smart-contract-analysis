package util

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
)

var fp *excelize.File
var currRow = 1

func CreateFileIfNotExists() {
	filePath := "out_data/total_debt.xlsx"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fp = excelize.NewFile()
		fp.SaveAs(filePath)
		fp, err = excelize.OpenFile(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
	} else {
		fp, err = excelize.OpenFile(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
	}
}

func ExportTotalDebtToExcel(contractDeployer, contractName string, debt []int) {
	CreateFileIfNotExists()
	sheetName := "Sheet1"

	cell, _ := excelize.CoordinatesToCellName(1, currRow)
	fp.SetCellValue(sheetName, cell, contractDeployer)
	cell, _ = excelize.CoordinatesToCellName(2, currRow)
	fp.SetCellValue(sheetName, cell, contractName)

	for i, debtAmt := range debt {
		cell, _ = excelize.CoordinatesToCellName(3+i, currRow)
		fp.SetCellValue(sheetName, cell, debtAmt)
	}

	currRow++
}

func SaveExcelFile(filePath string) {
	err := fp.SaveAs(filePath)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
}

func ExportCommentsToExcel(contractName string, contractVersion int, comments []string) {
	f := excelize.NewFile()

	sheetName := "Sheet1"

	fmt.Println()

	for i, comment := range comments {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		f.SetCellValue(sheetName, cell, comment)
	}

	outDir := fmt.Sprintf("out_data/contracts/%s", contractName)
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
