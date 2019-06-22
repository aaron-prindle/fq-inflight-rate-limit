package inflight

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"fmt"
)

// InitQueuesPriority is a convenience method for initializing an array of n queues
// for the full list of priorities
func InitQueuesPriority() []*Queue {
	// queues := make([]*Queue, 0, len(Priorities))
	queues := []*Queue{}
	for i, priority := range Priorities {
		queues = append(queues, &Queue{
			Packets:     []*Packet{},
			Priority:    priority,
			SharedQuota: 10,
			Index:       i,
		})
	}
	return queues
}

// test ideas
// 5 flows
// 1 concurrency for each
// make requests take 1 second
// send like 10 requests to each
// verify that one in each level was consumed

func TestInflight(t *testing.T) {
	var count int64
	queues := InitQueuesPriority()
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	fqFilter.Run()

	http.HandleFunc("/", fqFilter.Serve)
	fmt.Printf("Serving at 127.0.0.1:8080\n")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	// server := httptest.NewServer(http.HandlerFunc(handler))
	// defer server.Close()

	time.Sleep(1 * time.Second)
	header := make(http.Header)
	header.Add("PRIORITY", "0")
	req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	// req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header

	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	// fmt.Printf("results: %v\n", w.results)
	i := 0
	for result := range w.results {
		fmt.Printf("result %d:\n%v\n", i, result)
		i++
	}
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

// test ideas
// 3 flows
// 1 concurrency for each (SCL = 3?)
// make requests take 2 second
// send like 10 requests to each
// stop collection after  3 seconds
// verify that one in each level was consumed

func Test2Inflight(t *testing.T) {
	var count0 int64
	var count1 int64
	var count2 int64
	queues := InitQueuesPriority()
	fmt.Printf("len(queues): %d\n", len(queues))
	fqFilter := NewFQFilter(queues)
	fqFilter.Delegate = func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("r.Header[\"PRIORITY\"]: %s\n", r.Header["PRIORITY"])
		time.Sleep(1 * time.Second)
		if r.Header.Get("PRIORITY") == "0" {
			atomic.AddInt64(&count0, int64(1))
		}
		if r.Header.Get("PRIORITY") == "1" {
			atomic.AddInt64(&count1, int64(1))
		}
		if r.Header.Get("PRIORITY") == "2" {
			atomic.AddInt64(&count2, int64(1))
		}
	}
	fqFilter.Run()

	http.HandleFunc("/", fqFilter.Serve)
	fmt.Printf("Serving at 127.0.0.1:8080\n")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	// server := httptest.NewServer(http.HandlerFunc(handler))
	// defer server.Close()

	time.Sleep(1 * time.Second)
	ws := []*Work{}
	for i := 0; i < 3; i++ {
		header := make(http.Header)
		header.Add("PRIORITY", strconv.Itoa(i))
		req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
		// req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header = header

		ws = append(ws, &Work{
			Request: req,
			N:       5,
			C:       5,
		})

	}
	for i, w := range ws {
		w := w
		fmt.Printf("w %d info: %s\n", i, w.Request.Header.Get("PRIORITY"))
		go func() { w.Run() }()
	}

	time.Sleep(2 * time.Second)
	if count0 != 1 {
		t.Errorf("Expected to send 1 request, found %v", count0)
	}
	if count1 != 1 {
		t.Errorf("Expected to send 1 request, found %v", count0)
	}
	if count2 != 1 {
		t.Errorf("Expected to send 1 request, found %v", count0)
	}

}
