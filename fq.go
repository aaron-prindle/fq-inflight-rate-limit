package inflight

import (
	"fmt"

	fq "github.com/aaron-prindle/fq-apiserver"
	"k8s.io/apimachinery/pkg/util/clock"
)

// FQScheduler is a fair queuing implementation designed for the kube-apiserver.
// FQ is designed for
// 1) dispatching requests to be served rather than packets to be transmitted
// 2) serving multiple requests at once
// 3) accounting for unknown and varying service time
type FQScheduler struct {
	fq           *fq.FQScheduler
	queuedistchs map[int][]interface{}
}

func NewFQScheduler(queues []fq.FQQueue, clock clock.Clock) *FQScheduler {
	fq := &FQScheduler{
		fq:           fq.NewFQScheduler(queues, clock),
		queuedistchs: make(map[int][]interface{}),
	}
	return fq
}

func (q *FQScheduler) Enqueue(queue fq.FQQueue) <-chan func() {
	distributionCh := make(chan func(), 1)
	pkt := &fq.Packet{
		// TODO(aaron-prindle) HACK, fix
		QueueIdx: 0,
		// QueueIdx: queue.Index,
	}
	// TODO(aaron-prindle) make it so enqueue fails if the queues are full
	q.fq.Enqueue(pkt)
	q.queuedistchs[pkt.QueueIdx] = append(q.queuedistchs[pkt.QueueIdx], distributionCh)

	return distributionCh
}

// FinishPacket is a callback that should be used when a previously dequeud packet
// has completed it's service.  This callback updates imporatnt state in the
//  FQScheduler
func (q *FQScheduler) FinishPacket(p fq.FQPacket) {
	q.fq.FinishPacket(p)
}

// Dequeue dequeues a packet from the fair queuing scheduler
func (q *FQScheduler) Dequeue() (chan func(), fq.FQPacket) {
	pkt, ok := q.fq.Dequeue()
	if !ok {
		return nil, nil
	}
	// dequeue
	id := pkt.GetQueueIdx()
	fmt.Printf("pkt: %v\n", pkt)
	fmt.Printf("len(q.queuedistchs): %d\n", len(q.queuedistchs))
	distributionCh := q.queuedistchs[id][0].(chan func())
	q.queuedistchs[id] = q.queuedistchs[id][1:]
	return distributionCh, pkt
}
