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
	VuId        int
	NumRequests int
	ServerIPs   [7]string
	Wg          *sync.WaitGroup
}

type ResponseDetails struct {
	statusCode int
	duration   float64
	url        string
}

const kvPairBufferSize = 128 * 1000

var client *http.Client

func initHttpClient() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	client = &http.Client{Transport: tr}
}

func sendRequest(url string, locationId string, data []byte) (ResponseDetails, error) {
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
		return ResponseDetails{}, err
	}

	duration := time.Since(startTime).Seconds()
	// requestCounter.WithLabelValues(url).Inc()
	// requestDuration.Observe(duration)

	// if res.StatusCode >= 400 {
	// 	requestErrors.WithLabelValues(url, strconv.Itoa(res.StatusCode)).Inc()
	// }
	return ResponseDetails{
		statusCode: res.StatusCode,
		duration:   duration,
		url:        url,
	}, nil
}

func decodePayload(line string) KeyValuePair {
	kvPair := KeyValuePair{}
	json.Unmarshal([]byte(line), &kvPair)
	return kvPair
}

func (vu VirtualUser) StartLoadTest(inputChannel <-chan []string, acceptedWritesCh chan<- []byte) {
	defer vu.Wg.Done()

	initHttpClient()

	log.Printf("Virtual user %v starting load test\n", vu.VuId)

	vu.sendRequests(inputChannel, acceptedWritesCh)

	log.Printf("Virtual user %v load testing done\n", vu.VuId)
}

func recordRequest(responseDetails ResponseDetails, kvPairLine string, kvPairBuffer *bytes.Buffer, acceptedWritesCh chan<- []byte) {
	requestCounter.WithLabelValues(responseDetails.url).Inc()
	requestDuration.Observe(responseDetails.duration)

	if responseDetails.statusCode >= 400 {
		requestErrors.WithLabelValues(responseDetails.url, strconv.Itoa(responseDetails.statusCode)).Inc()
	}
	kvPairBuffer.WriteString(kvPairLine)
	kvPairBuffer.WriteByte('\n')

	if kvPairBuffer.Len() >= kvPairBufferSize {
		acceptedWritesCh <- kvPairBuffer.Bytes()
		kvPairBuffer.Reset()
	}
}

func (vu VirtualUser) sendRequests(inputChannel <-chan []string, acceptedWritesCh chan<- []byte) {
	var numSentRequests int
	kvPairBuffer := bytes.NewBuffer(make([]byte, 0, kvPairBufferSize))

	defer func() {
		if kvPairBuffer.Len() > 0 {
			acceptedWritesCh <- kvPairBuffer.Bytes()
		}
	}()

	bar := progressbar.Default(int64(vu.NumRequests), "Progress")
	for inputLines := range inputChannel {
		for _, line := range inputLines {
			kvPair := decodePayload(line)
			body, err := json.Marshal(kvPair.Data)
			if err != nil {
				log.Fatalf("Error while marshalling JSON body: %v", err)
			}
			serverUrl := vu.ServerIPs[numSentRequests%len(vu.ServerIPs)]
			res, err := sendRequest(serverUrl, kvPair.LocationId, body)
			if err == nil {
				recordRequest(res, line, kvPairBuffer, acceptedWritesCh)
			}
			bar.Add(1)
			numSentRequests += 1
			if numSentRequests >= vu.NumRequests {
				return
			}
		}
	}
}
