package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/schollz/progressbar/v3"
)

type LocationData struct {
	Id              uuid.UUID `json:"id"`
	SeismicActivity float32   `json:"seismic_activity"`
	TemperatureC    float32   `json:"temperature_c"`
	RadiationLevel  float32   `json:"radiation_level"`
}

type KeyValuePair struct {
	LocationId string       `json:"location_id"`
	Data       LocationData `json:"data"`
}

type VirtualUser struct {
	VuId         int
	NumRequests  int
	InputChannel <-chan []string
	ServerIPs    [7]string
	Wg           *sync.WaitGroup
}

var client *http.Client

func initHttpClient() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	client = &http.Client{Transport: tr}
}

func makeRequest(url string, locationId string, data []byte) {
	serverUrlFormatted := fmt.Sprintf("%s/%s", url, locationId)
	req, err := http.NewRequest(http.MethodPut, serverUrlFormatted, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error while creating request: %v", err)
	}
	startTime := time.Now()
	res, err := client.Do(req)
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	if err != nil {
		log.Printf("Error while sending request: %v", err)
		return
	}

	duration := time.Since(startTime).Seconds()
	requestCounter.WithLabelValues(url).Inc()
	requestDuration.Observe(duration)

	if res.StatusCode >= 400 {
		requestErrors.WithLabelValues(url, strconv.Itoa(res.StatusCode)).Inc()
		// log.Printf("Bad status code: %v", res.StatusCode)
	}
}

func decodePayloadData(line string) KeyValuePair {
	kvPair := KeyValuePair{}
	json.Unmarshal([]byte(line), &kvPair)
	return kvPair
}

func (vu VirtualUser) LoadTest() {
	defer vu.Wg.Done()
	initHttpClient()
	log.Printf("Virtual user %v starting load test\n", vu.VuId)
	s := fmt.Sprintf("VU %d Progress", vu.VuId)

	bar := progressbar.Default(int64(vu.NumRequests), s)

	var numSentRequests int
	for inputLines := range vu.InputChannel {
		for _, line := range inputLines {
			serverIndex := numSentRequests % len(vu.ServerIPs)
			serverUrl := vu.ServerIPs[serverIndex]
			kvPair := decodePayloadData(line)
			body, err := json.Marshal(kvPair.Data)
			if err != nil {
				log.Fatalf("Error while marshalling JSON body: %v", err)
			}
			makeRequest(serverUrl, kvPair.LocationId, body)
			bar.Add(1)
		}
		numSentRequests += len(inputLines)
		if numSentRequests >= vu.NumRequests {
			break
		}
	}
	log.Printf("Virtual user %v load testing done\n", vu.VuId)
}
