package logging

import (
	"log"
	"os"
)

var Logger *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
