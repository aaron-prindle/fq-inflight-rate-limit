package main

import (
	"log"
	"net/http"

	"fmt"

	inflight "github.com/aaron-prindle/fq-inflight-rate-limit"
)

func main() {

	queues := inflight.InitQueuesPriority()
	// queues := inflight.InitQueues(10)
	fqFilter := inflight.NewFQFilter(queues)

	fqFilter.Run()

	http.HandleFunc("/", fqFilter.Serve)
	fmt.Printf("Serving at 127.0.0.1:8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
