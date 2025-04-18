package main

import (
	"api_load_testing/internal"
	"log"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

func loadTest(numRequestsPerVU int, numVUs int) {
	serverIps := internal.ReadConfig()
	internal.CheckClusterHealth(serverIps)

	go internal.StartMetricsServer()

	ch := internal.ReadKeyValuePairs("output", 64*1000)

	startVUs(numVUs, numRequestsPerVU, serverIps, ch)
	internal.DisplayPostTestInformation()
}

func startVUs(numVUs int, numRequestsPerVU int, serverIps [7]string, dataInputCh <-chan []string) {
	acceptedWritesCh := make(chan []byte, 10)
	defer close(acceptedWritesCh)

	var wg sync.WaitGroup

	internal.StartTrackingWrites(acceptedWritesCh)

	for i := range numVUs {
		wg.Add(1)
		vu := internal.VirtualUser{
			VuId:        i,
			NumRequests: numRequestsPerVU,
			ServerIPs:   internal.ShuffleServerIps(serverIps),
			Wg:          &wg,
		}
		go vu.StartLoadTest(dataInputCh, acceptedWritesCh)
	}
	wg.Wait()
}

func checkDataIntegrity() {
	serverIps := internal.ReadConfig()
	internal.CheckClusterHealth(serverIps)
	internal.ValidateData(serverIps)
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

	var writeCmd = &cobra.Command{
		Use:   "write",
		Short: "Start the write process",
		Long:  "Start write process to cluster by sending PUT requests",
		Run: func(cmd *cobra.Command, args []string) {
			if numVUs == 0 || numRequestsPerVU == 0 {
				log.Println("Flags missing")
				rootCmd.Help()
				os.Exit(1)
			}
			log.Printf("Calling write commad, num requests per VU: %v, num VUs: %v", numRequestsPerVU, numVUs)
			loadTest(numRequestsPerVU, numVUs)
		},
	}

	writeCmd.PersistentFlags().IntVarP(&numVUs, "vus", "", 0, "Number of virtual users to simulate")
	writeCmd.PersistentFlags().IntVarP(&numRequestsPerVU, "reqs", "", 0, "Number of requests per virtual user")

	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate written data from the cluster using GET requests",
		Run: func(cmd *cobra.Command, args []string) {
			checkDataIntegrity()
		},
	}

	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println("CLI error:", err)
		os.Exit(1)
	}
}
