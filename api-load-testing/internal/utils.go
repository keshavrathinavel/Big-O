package internal

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func ShuffleServerIps(ips [7]string) [7]string {
	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})
	return ips
}

func DisplayPostTestInformation() {

	fmt.Println("\nWaiting 10s for prometheus to scrape metrics...")
	time.Sleep(10 * time.Second)

	fmt.Println("\nWaiting 10s for prometheus to scrape metrics...DONE")
	fmt.Println(`

Instructions to submit result:

1. Visit http://localhost:3000/d/befi36fr71atca/bigo-monitoring
2. In the Reqs/Sec Graph, select the portion of the graph post-request rampup and pre-request ramp down (basically the first highest peak and the last highest peak). This can be done by left clicking and dragging the mouse across the two points.
3. Capture a screenshot containing the graphs in the dashboard.`)
}
