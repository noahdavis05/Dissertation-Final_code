package types

import "time"

type TelemetryMetric struct {
	Low  float64
	Mean float64
	High float64
}

type NodeTelemetryMetrics struct {
	CPU           TelemetryMetric
	RAM           TelemetryMetric
	LastScheduled time.Time // will default to  0001-01-01 00:00:00 +0000 UTC before anything has been scheduled
	PodsScheduled []ScheduledPod
}

type FuzzyNumber struct {
	A float64
	B float64
	C float64
}

type FuzzyDecisionMatrix struct {
	// e.g. Data["node1"]["CPU"] = fuzzy number
	Data map[string]map[string]FuzzyNumber

	// e.g. Weights["CPU"] = fuzzy number
	Weights map[string]FuzzyNumber

	// positive ideal solutions - e.g. what we want our nodes to look like
	// e.g. PositiveIdeals["CPU"] = (75,75,75)
	PositiveIdeals map[string]FuzzyNumber

	// negative ideal solutions - e.g. what we don't want our nodes to look like
	// e.g. NegativeIdeals["CPU"] = (100,100,100)
	NegativeIdeals map[string]FuzzyNumber

	// a list of our columns. E.g. CPU and RAM
	Criteria []string
}

type TOPSISDistances struct {
	PositiveDistance float64
	NegativeDistance float64
}

// store info about the cluster in memory
// includes CPU and RAM limits per node
type ClusterInfo struct {
	// maps of node names to value
	CPULimits map[string]int64
	RAMLimits map[string]int64
}

type PodRequest struct {
	CPU int64
	RAM int64
}

type ScheduledPod struct {
	Requests  PodRequest
	Timestamp time.Time
}
