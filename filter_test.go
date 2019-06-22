package inflight

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	fq "github.com/aaron-prindle/fq-apiserver"
)

// initQueuesPriority is a convenience method for initializing an array of n queues
// for the full list of priorities
func initQueuesPriority() []*Queue {
	queues := make([]*Queue, 0, len(Priorities))
	for i, priority := range Priorities {
		queues = append(queues, &Queue{
			Priority:    priority,
			SharedQuota: 10,
			Index:       i,
		})
		packets := []*fq.Packet{}
		fqpackets := make([]fq.FQPacket, len(packets), len(packets))
		for i := range packets {
			fqpackets[i] = fqpackets[i]
		}

		queues[i].Packets = fqpackets
	}
	return queues
}

func TestInflightSingle(t *testing.T) {
	var count int64
	queues := initQueuesPriority()
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	fqFilter.Run()

	server := httptest.NewServer(http.HandlerFunc(fqFilter.Serve))
	defer server.Close()

	time.Sleep(1 * time.Second)
	header := make(http.Header)
	header.Add("PRIORITY", "0")
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header

	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	// i := 0
	// for result := range w.results {
	// fmt.Printf("result %d:\n%v\n", i, result)
	// i++
	// }
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

// 5 flows
// 1 concurrency for each (SCL = 3?)
// ACV = 1
// make requests take 1 second
// send like 10 requests to each, with 10 concurrent channels
// stop collection after  2 seconds
// verify that only one request in each level was consumed, showing adherence
// to SCL/ACV and fairness

func TestInflightMultiple(t *testing.T) {
	countnum := len(Priorities)
	counts := []int64{}
	for range Priorities {
		counts = append(counts, int64(0))
	}
	queues := initQueuesPriority()
	// fmt.Printf("len(queues): %d\n", len(queues))
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		for i := 0; i < countnum; i++ {
			if r.Header.Get("PRIORITY") == strconv.Itoa(i) {
				atomic.AddInt64(&counts[i], int64(1))
			}
		}
	}
	fqFilter.Run()

	server := httptest.NewServer(http.HandlerFunc(fqFilter.Serve))
	defer server.Close()

	time.Sleep(1 * time.Second)
	ws := []*Work{}
	for i := 0; i < countnum; i++ {
		header := make(http.Header)
		header.Add("PRIORITY", strconv.Itoa(i))
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header = header

		ws = append(ws, &Work{
			Request: req,
			N:       5,
			C:       5,
		})

	}
	for _, w := range ws {
		w := w
		// fmt.Printf("w %d info: %s\n", i, w.Request.Header.Get("PRIORITY"))
		go func() { w.Run() }()
	}

	// in 1 seconds, only 1 request should be seen for each flow w/ Delegation
	// taking .5 seconds
	time.Sleep(1 * time.Second)
	for i, count := range counts {
		if count != 1 {
			t.Errorf("Expected to dispatch 1 request for Priority %d, found %v", i, count)
		}
	}
}

// 5 flows
// 1 concurrency for each (SCL = 3?)
// ACV = 1
// make requests take .001 second
// send  10 requests to each, with 10 concurrent channels
// stop collection after  2 seconds
// verify that 10 requests in each level was consumed

func TestInflightMultiple2(t *testing.T) {
	const countnum = 5
	counts := [countnum]int64{}
	queues := initQueuesPriority()
	// fmt.Printf("len(queues): %d\n", len(queues))
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < countnum; i++ {
			if r.Header.Get("PRIORITY") == strconv.Itoa(i) {
				atomic.AddInt64(&counts[i], int64(1))
			}
		}
	}
	fqFilter.Run()

	server := httptest.NewServer(http.HandlerFunc(fqFilter.Serve))
	defer server.Close()

	time.Sleep(1 * time.Second)
	ws := []*Work{}
	for i := 0; i < countnum; i++ {
		header := make(http.Header)
		header.Add("PRIORITY", strconv.Itoa(i))
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header = header

		ws = append(ws, &Work{
			Request: req,
			N:       5,
			C:       5,
		})

	}
	for _, w := range ws {
		w := w
		// fmt.Printf("w %d info: %s\n", i, w.Request.Header.Get("PRIORITY"))
		go func() { w.Run() }()
	}

	// in 1 seconds, 5 requests should be seen for each flow w/ Delegation
	time.Sleep(1 * time.Second)
	for i, count := range counts {
		if count != 5 {
			t.Errorf("Expected to dispatch 1 request for Priority %d, found %v", i, count)
		}
	}
}
