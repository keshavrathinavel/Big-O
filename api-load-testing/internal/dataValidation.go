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
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Printf("Total Correct Reads: %d\n", totalCorrectReads)
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
		fmt.Printf("Got status code: %d\n", res.StatusCode)
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

func (vw *ValidationWorker) startValidation(inputCh chan []string) {
	defer vw.wg.Done()

	fmt.Printf("Validation worker with id %d started\n", vw.id)

	var count int32
	var incorrectDataCount int32
	var missingDataCount int32

	for writes := range inputCh {
		for _, line := range writes {
			kvPair, err := decodePayload(line)
			if err != nil {
				continue
			}

			requestUrl := fmt.Sprintf("%v/%v", vw.serverIps[count%int32(len(vw.serverIps))], kvPair.LocationId)
			data, err := vw.makeGetRequest(requestUrl)

			if err != nil {
				if errors.Is(err, ErrHttpNotFound) {
					fmt.Printf("Missing data for location ID: %v\n", kvPair.LocationId)
					fmt.Println(kvPair)
					missingDataCount++
				} else {
					log.Printf("Error while sending GET request: %v", err)
				}
				continue
			}

			if !isDataValid(kvPair, data) {
				// TODO: do something with inconsistent data, write to file maybe ?
				incorrectDataCount++
			}
			count++
		}
	}
	if count > 0 {
		log.Printf(
			"Validation worker ID %d completed requests: %d Correct, %d Incorrect, %d Not Found", vw.id, count-incorrectDataCount, incorrectDataCount, missingDataCount)
	}

	atomic.AddInt32(&totalCorrectReads, count)
	atomic.AddInt32(&totalIncorrectReads, incorrectDataCount)
	atomic.AddInt32(&totalMissingDataReads, missingDataCount)
}
