package main

import (
	"api_load_testing/internal"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func shuffleServerIps(ips [7]string) [7]string {
	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})
	return ips
}

func loadTest(numRequestsPerVU int, numVUs int) {
	fileData, err := os.ReadFile("config.yaml")

	if err != nil {
		log.Fatalf("Error while reading config file: %v", err)
	}
	serverIps := internal.ReadConfig(fileData)
	checkServerHealth(serverIps)
	var wg sync.WaitGroup

	ch := internal.ReadKeyValuePairs("output", 64*1000)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Println("Error while serving metrics endpoint:", err)
		}
	}()
	for i := 0; i < numVUs; i++ {
		wg.Add(1)
		vu := internal.VirtualUser{
			VuId:         i,
			NumRequests:  numRequestsPerVU,
			InputChannel: ch,
			ServerIPs:    shuffleServerIps(serverIps),
			Wg:           &wg,
		}
		go vu.LoadTest()
	}
	wg.Wait()
	fmt.Println("\nWaiting 10s for prometheus to scrape metrics...")
	time.Sleep(10 * time.Second)
	fmt.Println("\nWaiting 10s for prometheus to scrape metrics...DONE")
	fmt.Println(`
	
Instructions to submit result:

1. Visit http://localhost:3000/d/befi36fr71atca/bigo-monitoring
2. In the Reqs/Sec Graph, select the portion of the graph post-request rampup and pre-request ramp down (basically the first highest peak and the last highest peak). This can be done by left clicking and dragging the mouse across the two points.
3. Capture a screenshot containing the graphs in the dashboard.`)

}

func checkServerHealth(serverIPs [7]string) {
	for _, ip := range serverIPs {
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

func main() {

	rootCmd := &cobra.Command{
		Use: "load_test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	var numVUs int
	var numRequestsPerVU int

	rootCmd.PersistentFlags().IntVarP(&numVUs, "vus", "", 0, "Number of virtual users to simulate")
	rootCmd.PersistentFlags().IntVarP(&numRequestsPerVU, "reqs", "", 0, "Number of requests per virtual user")

	if err := rootCmd.Execute(); err != nil {
		log.Println("CLI error:", err)
		os.Exit(1)
	}

	if numVUs == 0 || numRequestsPerVU == 0 {
		fmt.Println("Flags missing")
		rootCmd.Help()
		os.Exit(1)
	}
	loadTest(numRequestsPerVU, numVUs)
}
