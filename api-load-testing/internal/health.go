package internal

import (
	"log"
	"net/http"
	"time"
)

func CheckClusterHealth(serverIps [7]string) {
	for _, ip := range serverIps {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get(ip + "/health")
		if err != nil {
			log.Fatalf("failed to connect to health endpoint: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Server with addr %s returned status %d instead of 200", ip, resp.StatusCode)
		}
		log.Printf("Server with addr %s is healthy", ip)
	}
}
