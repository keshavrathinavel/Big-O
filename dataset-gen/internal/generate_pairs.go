package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"sync"

	"github.com/google/uuid"
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

func New() (LocationData, error) {
	uuidValue, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("Error while generating UUID: %v", err)
		return LocationData{}, err
	}
	return LocationData{
		Id:              uuidValue,
		SeismicActivity: rand.Float32(),
		TemperatureC:    rand.Float32(),
		RadiationLevel:  rand.Float32(),
	}, nil
}

func GenerateKeyValuePairs(keysChannel <-chan []string, goroutineNumber int, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := fmt.Sprintf("output/output-%d.json", goroutineNumber)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error during output file creation: %v", err)
	}
	defer file.Close()
	for chunk := range keysChannel {
		var buffer bytes.Buffer

		for _, value := range chunk {
			locationId := fmt.Sprintf("PAND-%s", value)
			locationData, err := New()
			if err != nil {
				log.Fatalf("Error during dataset generation: %v", err)
			}

			pair := KeyValuePair{
				LocationId: locationId,
				Data:       locationData,
			}
			b, err := json.Marshal(pair)

			if err != nil {
				log.Fatalf("Error during JSON marshal: %v", err)
			}
			buffer.Write(b)
			buffer.WriteByte('\n')
		}

		file.Write(buffer.Bytes())
	}
}
