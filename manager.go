package inflight

import (
	"fmt"

	fq "github.com/aaron-prindle/fq-apiserver"
	"k8s.io/apimachinery/pkg/util/clock"
)

type sharedDispatcher struct {
	producers map[fq.PriorityBand]*Dispatcher
}

func queuesForPriority(priority fq.PriorityBand, queues []*fq.Queue) []*fq.Queue {
	// TODO(aaron-prindle) change this to actual impl
	return []*fq.Queue{queues[priority]}
}

func newSharedDispatcher(queues []*fq.Queue) *sharedDispatcher {

	mgr := &sharedDispatcher{
		producers: make(map[fq.PriorityBand]*Dispatcher),
	}

	clock := clock.RealClock{}
	for _, priority := range fq.Priorities {
		mgr.producers[priority] = &Dispatcher{
			queues: queues,
			ACV:    1,
			// ACV:         ACV(priority, queues),
			fqScheduler: NewFQScheduler(queuesForPriority(priority, queues), clock),
		}
	}
	// TODO(aaron-prindle) FIX - this eventually needs to be dynamic...
	for _, priority := range fq.Priorities {
		mgr.producers[priority].ACV = 1
		// mgr.producers[priority].ACV += ACV(priority, mgr.producers[priority].queues)
	}

	return mgr
}

func (m *sharedDispatcher) Run() {
	for i, producer := range m.producers {
		producer := producer
		fmt.Printf("producer[%d] starting...\n", i)
		go producer.Run()
	}
}
