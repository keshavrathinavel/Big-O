package utils

import (
	"log"
	"os"
)

var numKeys int
var buffer []byte
var stopGeneration bool

func Generate(totalKeys int) {
	characters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	file, err := os.Create("combinations.txt")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	stopGeneration = false
	numKeys = 0

	generateCombinations(characters, 6, []byte{}, file, totalKeys)

	// Write remaining buffer (if any)
	if len(buffer) > 0 {
		file.Write(buffer)
	}
}

func generateCombinations(characters string, length int, current []byte, file *os.File, n int) {
	if stopGeneration {
		return
	}

	if len(current) == length {
		buffer = append(buffer, current...)
		buffer = append(buffer, '\n')
		numKeys++

		if numKeys >= n {
			stopGeneration = true
			return
		}

		// Write if buffer exceeds 64KB
		if len(buffer) > 64000 {
			_, err := file.Write(buffer)
			if err != nil {
				log.Fatalf("Error writing to file: %v", err)
			}
			buffer = buffer[:0] // Reset buffer without reallocation
		}
		return
	}

	for i := 0; i < len(characters) && !stopGeneration; i++ {
		current = append(current, characters[i])
		generateCombinations(characters, length, current, file, n)
		current = current[:len(current)-1] // Backtrack
	}
}
