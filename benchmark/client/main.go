package main

import (
	"flag"
	"fmt"
	"github.com/hprose/hprose-golang/rpc"
	"github.com/rcrowley/go-metrics"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	server      string
	requests    int
	duration    time.Duration
	concurrency int

	mduration = metrics.GetOrRegisterTimer("duration", nil)
	errs      int64
)

type Stub struct {
	Hello func(string) (string, error)
}

func parseArgs() {
	//flag.StringVar(&server, "server", "", "server to run benchmark")
	flag.IntVar(&requests, "r", 1000000, "Number of requests to perform")
	flag.IntVar(&concurrency, "c", 1, "Number of multiple requests to make at a time")
	flag.Parse()
}

func main() {
	server = os.Args[len(os.Args)-1]
	parseArgs()

	wg := &sync.WaitGroup{}
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			client := rpc.NewClient(server)
			var s *Stub
			client.UseService(&s)
			for i := 0; i < requests/concurrency; i++ {
				start := time.Now()
				_, err := s.Hello("world")
				mduration.UpdateSince(start)
				if err != nil {
					atomic.AddInt64(&errs, 1)
				}
			}
		}()
	}

	wg.Wait()

	fmt.Printf("All done. %s\n", server)
	fmt.Printf("Total requests: %d, ErrCount: %d\n", requests, errs)
	fmt.Printf(".75: %.2f\n", mduration.Percentile(0.75)/float64(time.Millisecond))
	fmt.Printf(".90: %.2f\n", mduration.Percentile(0.90)/float64(time.Millisecond))
	fmt.Printf(".99: %.2f\n", mduration.Percentile(0.99)/float64(time.Millisecond))
}
