package internal

import (
	"log"
	"os"
	"path/filepath"
)

func pre() error {
	path := filepath.Join(".", "tracking")
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

func StartTrackingWrites(ch <-chan []byte) {
	err := pre()
	if err != nil {
		log.Fatalf("Errow while pre call in tracking: %v", err)
	}

	go func() {
		file, err := os.OpenFile("tracking/accepted_writes.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error while opening accepted writes file: %v", err)
		}
		defer file.Close()

		file.Truncate(0)
		file.Seek(0, 0)
		for data := range ch {
			file.Write(data)
		}
	}()
}
