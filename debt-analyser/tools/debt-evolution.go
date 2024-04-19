package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

// compareFiles compares two .xlsx files and returns the count of added and removed strings.
func compareFiles(file1, file2 string) (int, int, map[string]int, map[string]int) {
	wb1, err := xlsx.OpenFile(file1)
	if err != nil {
		panic(err) // Simplified error handling
	}
	wb2, err := xlsx.OpenFile(file2)
	if err != nil {
		panic(err)
	}

	// Assuming data is in the first sheet and in the first column
	ws1 := wb1.Sheets[0]
	ws2 := wb2.Sheets[0]

	count1 := make(map[string]int)
	count2 := make(map[string]int)

	for _, row := range ws1.Rows {
		if len(row.Cells) > 0 {
			val := row.Cells[0].String()
			if val != "" {
				count1[val]++
			}
		}
	}

	for _, row := range ws2.Rows {
		if len(row.Cells) > 0 {
			val := row.Cells[0].String()
			if val != "" {
				count2[val]++
			}
		}
	}

	added := 0
	removed := 0

	for key, val := range count2 {
		if _, exists := count1[key]; !exists {
			added += val
		} else {
			diff := val - count1[key]
			if diff > 0 {
				added += diff
			}
		}
	}

	for key, val := range count1 {
		if _, exists := count2[key]; !exists {
			removed += val
		} else {
			diff := val - count2[key]
			if diff > 0 {
				removed += diff
			}
		}
	}

	return added, removed, count1, count2
}

// processDirectory processes each directory and its files.
func processDirectory(rootDir string) {
	remDict := make(map[string][2]int)
	var dirs []string

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != rootDir {
			dirs = append(dirs, path)
		}
		return nil
	})

	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.xlsx"))
		if err != nil {
			panic(err) // Simplified error handling
		}
		sort.Slice(files, func(i, j int) bool {
			numI, _ := strconv.Atoi(strings.Split(filepath.Base(files[i]), ".")[0])
			numJ, _ := strconv.Atoi(strings.Split(filepath.Base(files[j]), ".")[0])
			return numI < numJ
		})

		for i := 0; i < len(files)-1; i++ {
			file1 := files[i]
			file2 := files[i+1]
			added, removed, count1, _ := compareFiles(file1, file2)
			if i == 0 {
				remDict[dir] = [2]int{len(count1), 0}
			}
			current := remDict[dir]
			remDict[dir] = [2]int{current[0] + added, current[1] + removed}
		}
	}

	removedInstances := 0
	for _, v := range remDict {
		if v[1] > 0 {
			removedInstances++
		}
	}

	fmt.Printf("Occurrence of debt removal: %f%%\n", float64(removedInstances)*100/float64(len(remDict)))
}

func main() {
	rootDir := "../debt_data/contracts"
	processDirectory(rootDir)
}
