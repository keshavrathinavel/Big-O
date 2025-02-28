package main

import (
	"fmt"
	"os"
)

func main() {
	// Define the characters to be used (0-9 and A-Z)
	characters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Open a file for writing
	file, err := os.Create("combinations.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Generate all combinations of length 6
	generateCombinations(characters, 6, "", file)
}

// Recursive function to generate combinations
func generateCombinations(characters string, length int, current string, file *os.File) {
	// If the current combination is of the desired length, write it to the file
	if len(current) == length {
		_, err := file.WriteString(current + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		return
	}

	// Recursively build the combination
	for _, char := range characters {
		generateCombinations(characters, length, current+string(char), file)
	}
}
