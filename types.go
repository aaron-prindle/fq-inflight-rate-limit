package inflight

import (
	fq "github.com/aaron-prindle/fq-apiserver"
)

const SCL = 3

type PriorityBand int

const (
	SystemTopPriorityBand = PriorityBand(iota)
	SystemHighPriorityBand
	SystemMediumPriorityBand
	SystemNormalPorityBand
	SystemLowPriorityBand

	// This is an implicit priority that cannot be set via API
	SystemLowestPriorityBand
)

var Priorities = []PriorityBand{
	SystemTopPriorityBand,
	SystemHighPriorityBand,
	SystemMediumPriorityBand,
	SystemNormalPorityBand,
	SystemLowPriorityBand,
	SystemLowestPriorityBand,
}

// Queue is an array of packets with additional metadata required for
// the FQScheduler
type Queue struct {
	fq.Queue
	Priority    PriorityBand
	SharedQuota int
	Index       int
}
