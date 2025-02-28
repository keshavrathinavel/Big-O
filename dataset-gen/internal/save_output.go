package internal

import (
	"fmt"
	"log"
	"os"
)

func saveData(data []byte, goroutineNumber int) {
	if len(data) > 0 {
		fileName := fmt.Sprintf("output/output-{%d}.json", goroutineNumber)
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error during output file creation: %v", err)
			return
		}
		file.Write(data)
	}
}
