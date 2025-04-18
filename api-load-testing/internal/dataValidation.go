package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type LocationResponseData struct {
	LocationData
	locationId        string `json:"location_id"`
	modificationCount int    `json:"modification_count"`
}

type ValidationWorker struct {
	id        int
	serverIps [7]string
	client    *http.Client
	wg        *sync.WaitGroup
}

var totalCorrectReads int32
var totalIncorrectReads int32
var totalMissingDataReads int32
var ErrHttpNotFound = errors.New("location ID not found")

func ValidateData(serverIps [7]string) {
	writesToCheckCh := make(chan []string)

	file, err := os.Open("tracking/accepted_writes.txt")
	if err != nil {
		log.Fatalf("Error while opening accepted writes file: %v", err)
	}
	go func() {
		defer close(writesToCheckCh)
		readChunksFromFile(file, writesToCheckCh, 8*1000)
	}()
	var wg sync.WaitGroup

	go func() {
		fmt.Print("\n\nValidation Metrics(Updating every 1 second)\n\n")
		for {
			printValidationMetrics()
			time.Sleep(1 * time.Second)
			clearLines(5)
		}
	}()

	for i := range 4 {
		wg.Add(1)
		worker := ValidationWorker{
			id:        i,
			serverIps: serverIps,
			wg:        &wg,
		}
		worker.initClient()
		go worker.startValidation(writesToCheckCh)
	}
	wg.Wait()
	fmt.Printf("\nTotal Correct Reads: %d\n", totalCorrectReads)
	fmt.Printf("Total Incorrect Reads: %d\n", totalIncorrectReads)
	fmt.Printf("Total Missing Data Reads: %d\n", totalMissingDataReads)
	fmt.Println("----------------------------------------------------------------------------------")
}

func (vw *ValidationWorker) initClient() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	vw.client = &http.Client{Transport: tr}
}

func (vw *ValidationWorker) makeGetRequest(url string) (LocationResponseData, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return LocationResponseData{}, fmt.Errorf("error while creating request: %w", err)
	}

	res, err := vw.client.Do(req)
	if err != nil {
		return LocationResponseData{}, fmt.Errorf("error while sending request: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return LocationResponseData{}, ErrHttpNotFound
	}

	if err != nil {
		return LocationResponseData{}, fmt.Errorf("error while reading response body: %w", err)
	}
	var responseData LocationResponseData
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return LocationResponseData{}, fmt.Errorf("error while unmarshalling JSON: %w", err)
	}
	return responseData, nil
}

func isDataValid(expectedKV KeyValuePair, actualData LocationResponseData) bool {
	return expectedKV.Data == actualData.LocationData
}

func printValidationMetrics() {
	fmt.Println("+----------------+----------------+----------------+----------------+")
	fmt.Println("|  Total Reads   | Correct Reads  | Incorrect Reads| Data not found |")
	fmt.Println("+----------------+----------------+----------------+----------------+")
	fmt.Printf("| %-14d | %-14d | %-14d | %-14d |\n", totalCorrectReads+totalIncorrectReads+totalMissingDataReads, totalCorrectReads, totalIncorrectReads, totalMissingDataReads)
	fmt.Println("+----------------+----------------+----------------+----------------+")
}

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[1A") // Move cursor up one line
		fmt.Print("\033[2K") // Clear the entire line
	}
}

func (vw *ValidationWorker) startValidation(inputCh chan []string) {
	defer vw.wg.Done()

	for writes := range inputCh {
		var count int32
		var incorrectDataCount int32
		var missingDataCount int32
		for _, line := range writes {
			kvPair, err := decodePayload(line)
			if err != nil {
				continue
			}

			requestUrl := fmt.Sprintf("%v/%v", vw.serverIps[count%int32(len(vw.serverIps))], kvPair.LocationId)
			data, err := vw.makeGetRequest(requestUrl)

			if err != nil {
				if errors.Is(err, ErrHttpNotFound) {
					missingDataCount++
				}
				continue
			}

			if !isDataValid(kvPair, data) {
				// TODO: do something with inconsistent data, write to file maybe ?
				incorrectDataCount++
			}
			count++
		}
		atomic.AddInt32(&totalCorrectReads, count)
		atomic.AddInt32(&totalIncorrectReads, incorrectDataCount)
		atomic.AddInt32(&totalMissingDataReads, missingDataCount)
	}
}
