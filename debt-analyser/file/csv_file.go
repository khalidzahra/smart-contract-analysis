package file

import (
	"log"
	"os"
	"strings"
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
