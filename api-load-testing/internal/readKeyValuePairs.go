package internal

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadKeyValuePairs(folderPath string, chunkSize int) <-chan []string {
	ch := make(chan []string, 5)
	go func() {
		defer close(ch)
		err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatalf("Error while accessing path inside dir: %v", err)
			}
			file, err := os.Open(path)
			readChunksFromFile(file, ch, chunkSize)

			return nil
		})
		if err != nil {
			log.Fatalf("Error while walking dir: %v", err)
			return
		}
	}()
	return ch
}

func readChunksFromFile(file *os.File, ch chan []string, chunkSize int) {
	reader := bufio.NewReader(file)
	buffer := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil {
			log.Fatal("Error while reading file")
			break
		}

		readBuffer := buffer[:n]
		data := string(readBuffer)

		lastCharacter := data[n-1]
		if lastCharacter != 10 {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalln("Error while reading file")
				break
			}
			data += line
		}
		splitData := strings.Split(data, "\n")
		if splitData[len(splitData)-1] == "" {
			splitData = splitData[:len(splitData)-1]
		}
		ch <- splitData
	}
}
