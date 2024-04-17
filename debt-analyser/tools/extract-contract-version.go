package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:] // Skip the program path

	// Check if any arguments were passed
	if len(args) == 0 {
		fmt.Println("No arguments were provided.")
		return
	}

	// Specify the file name you want to read
	fileName := args[0]

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close() // Make sure to close the file when you're done

	// Create a new Scanner for the file
	scanner := bufio.NewScanner(file)

	// Loop through all lines of the file
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "/") || strings.HasPrefix(line, "*") {
			continue
		}
		version := strings.Split(line, " ")[2][1:]
		version = version[:len(version)-1]
		fmt.Println(version)
		// TODO run solc commands
		break
	}

	// Check for errors during Scan. End of file is expected and not reported by Scan as an error.
	if err := scanner.Err(); err != nil {
		log.Fatalf("error during file scan: %s", err)
	}
}
