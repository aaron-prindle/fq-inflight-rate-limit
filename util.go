package inflight

import (
	"math"

	fq "github.com/aaron-prindle/fq-apiserver"
)

func min(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

func ACS(pl fq.PriorityBand, queues []*fq.Queue) int {
	assuredconcurrencyshares := 0
	for _, queue := range queues {
		if queue.Priority == pl {
			assuredconcurrencyshares += queue.SharedQuota
		}
	}
	return assuredconcurrencyshares
}

// TODO(aaron-prindle) verify this is correct, i would think sum[prioity levels k]ACV(k) == SCL
func ACV(pl fq.PriorityBand, queues []*fq.Queue) int {
	// ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
	denom := 0
	for _, prioritylvl := range fq.Priorities {
		denom += ACS(prioritylvl, queues)
	}
	return int(math.Ceil(float64(fq.SCL * ACS(pl, queues) / denom)))
}
