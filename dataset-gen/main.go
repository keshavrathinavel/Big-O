package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/rohanjnr/bigo/dataset-gen/internal"
	"github.com/spf13/cobra"
)

const keysFileName = "exp/combinations.txt"
const outputDirName = "output"

func pre() error {
	path := filepath.Join(".", outputDirName)
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

func main() {

	err := pre()
	if err != nil {
		log.Fatalf("Error during Pre() call: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "datagen",
		Short: "Generate dataset for BigO 2025",
	}

	var numKeyValuePairs int
	var numGoroutines int
	var chunkSize int

	rootCmd.PersistentFlags().IntVarP(&numKeyValuePairs, "num", "", 5000, "Number of key value pairs")
	rootCmd.PersistentFlags().IntVarP(&numGoroutines, "parallel", "", 1, "Number of goroutines")
	rootCmd.PersistentFlags().IntVarP(&chunkSize, "chunk_size", "", 32*100, "Chunk size to load data")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("CLI error: %v", err)
		return
	}

	file, err := os.Open(keysFileName)
	if err != nil {
		log.Fatalf("Error while opening file: %v", err)
		return
	}

	keysChannel := internal.ReadChunksFromFile(file, chunkSize, numKeyValuePairs)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go internal.GenerateKeyValuePairs(keysChannel, i, &wg)
	}

	wg.Wait()
}
