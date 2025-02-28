package internal

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func ReadChunksFromFile(file *os.File, chunkSize int, numKeys int) <-chan []string {
	reader := bufio.NewReader(file)
	ch := make(chan []string, 4)
	buffer := make([]byte, chunkSize)
	var numKeysVisited int
	go func() {
		defer close(ch)
		for numKeysVisited < numKeys {
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
			if numKeysVisited+len(splitData) > numKeys {
				splitData = splitData[:numKeys-numKeysVisited]
			}
			numKeysVisited += len(splitData)
			ch <- splitData
		}
	}()

	return ch
}
