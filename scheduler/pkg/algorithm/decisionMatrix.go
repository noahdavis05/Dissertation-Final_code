package algorithm

import (
	"fmt"
	"os"
	"scheduler/pkg/config"
	"scheduler/pkg/telemetry"
	"scheduler/pkg/types"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func BuildFuzzyDM(nodes []*corev1.Node) types.FuzzyDecisionMatrix {
	fuzzyDM := types.FuzzyDecisionMatrix{
		Data: make(map[string]map[string]types.FuzzyNumber),
	}
	fuzzyDM.Criteria = []string{
		"CPU",
		"RAM",
		"CPU RANGE",
		"RAM RANGE",
	}

	// These are the weights used as part of TOPSIS
	fuzzyDM.Weights = map[string]types.FuzzyNumber{
		"CPU":       config.CPUWeights,
		"RAM":       config.RAMWeights,
		"CPU RANGE": config.CPURangeWeights,
		"RAM RANGE": config.RAMRangeWeights,
	}

	// set the Ideal Positives and Ideal Negatives
	fuzzyDM.PositiveIdeals = map[string]types.FuzzyNumber{
		"CPU":       config.PosCPUIdeal,
		"RAM":       config.PosRAMIdeal,
		"CPU RANGE": config.PosCPURangeIdeal,
		"RAM RANGE": config.PosRAMRangeIdeal,
	}

	fuzzyDM.NegativeIdeals = map[string]types.FuzzyNumber{
		"CPU":       config.NegCPUIdeal,
		"RAM":       config.NegRAMIdeal,
		"CPU RANGE": config.NegCPURangeIdeal,
		"RAM RANGE": config.NegRAMRangeIdeal,
	}

	for _, node := range nodes {
		nodeMetrics, ok := telemetry.GetNodeMetrics(node.Name)
		if !ok {
			fmt.Println("Error getting node metrics")
			panic("Error getting node metrics")
		}

		// make a new row in FuzzyDM for this node
		fuzzyDM.Data[node.Name] = map[string]types.FuzzyNumber{
			"CPU": {
				A: nodeMetrics.CPU.Low,
				B: nodeMetrics.CPU.Mean,
				C: nodeMetrics.CPU.High,
			},
			"RAM": {
				A: nodeMetrics.RAM.Low,
				B: nodeMetrics.RAM.Mean,
				C: nodeMetrics.RAM.High,
			},
			"CPU RANGE": {
				A: 0,
				B: nodeMetrics.CPU.High - nodeMetrics.CPU.Low,
				C: nodeMetrics.CPU.High - nodeMetrics.CPU.Low,
			},
			"RAM RANGE": {
				A: 0,
				B: nodeMetrics.RAM.High - nodeMetrics.RAM.Low,
				C: nodeMetrics.RAM.High - nodeMetrics.RAM.Low,
			},
		}
	}

	return fuzzyDM
}

func DisplayFuzzyDM(fuzzyDM types.FuzzyDecisionMatrix) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)

	// build title line
	fmt.Fprint(w, "Node\t")
	for _, criterion := range fuzzyDM.Criteria {
		fmt.Fprintf(w, " %s (a, b, c)\t", criterion)
	}
	fmt.Fprintln(w)

	// add seperator
	fmt.Fprint(w, "---\t")
	for range fuzzyDM.Criteria {
		fmt.Fprint(w, "-----------\t")
	}
	fmt.Fprintln(w)

	// print rows
	for nodeName, metrics := range fuzzyDM.Data {
		fmt.Fprintf(w, "%s\t", nodeName)
		for _, criterion := range fuzzyDM.Criteria {
			f := metrics[criterion]
			fmt.Fprintf(w, " (%.2f, %.2f, %.2f)\t", f.A, f.B, f.C)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
	fmt.Println()
}

// builds a basic fuzzyDM from scratch without the data
// the data is set manually in each test
// this just loads in the weights and ideals from config
func buildTestingDM() types.FuzzyDecisionMatrix {
	fuzzyDM := types.FuzzyDecisionMatrix{
		Data: make(map[string]map[string]types.FuzzyNumber),
	}
	fuzzyDM.Criteria = []string{
		"CPU",
		"RAM",
		"CPU RANGE",
		"RAM RANGE",
	}

	// These are the weights used as part of TOPSIS
	fuzzyDM.Weights = map[string]types.FuzzyNumber{
		"CPU":       config.CPUWeights,
		"RAM":       config.RAMWeights,
		"CPU RANGE": config.CPURangeWeights,
		"RAM RANGE": config.RAMRangeWeights,
	}

	// set the Ideal Positives and Ideal Negatives
	fuzzyDM.PositiveIdeals = map[string]types.FuzzyNumber{
		"CPU":       config.PosCPUIdeal,
		"RAM":       config.PosRAMIdeal,
		"CPU RANGE": config.PosCPURangeIdeal,
		"RAM RANGE": config.PosRAMRangeIdeal,
	}

	fuzzyDM.NegativeIdeals = map[string]types.FuzzyNumber{
		"CPU":       config.NegCPUIdeal,
		"RAM":       config.NegRAMIdeal,
		"CPU RANGE": config.NegCPURangeIdeal,
		"RAM RANGE": config.NegRAMRangeIdeal,
	}

	return fuzzyDM
}

// returns true if node should be filtered out
func filterNode(fuzzyDM *types.FuzzyDecisionMatrix, name string, podRequests types.PodRequest, clusterLimits types.ClusterInfo) bool {
	// calculate the CPU and RAM request as a percentage of the nodes total CPU and RAM limit
	percentageCPURequest := (float64(podRequests.CPU) / float64(clusterLimits.CPULimits[name])) * 100
	percentageRAMRequest := (float64(podRequests.RAM) / float64(clusterLimits.RAMLimits[name])) * 100

	if fuzzyDM.Data[name]["CPU"].B > fuzzyDM.NegativeIdeals["CPU"].C-float64(percentageCPURequest) {
		return true
	}

	if fuzzyDM.Data[name]["RAM"].B > fuzzyDM.NegativeIdeals["RAM"].C-float64(percentageRAMRequest) {
		return true
	}
	return false
}

func weightNodes(fuzzyDM *types.FuzzyDecisionMatrix) {
	// the desicion matrix is passed as pointer so doesn't need to be changed
	for k, v := range fuzzyDM.Data {
		for key, value := range v {
			// key is field e.g. CPU
			// value is the FuzzyNumber we need to update
			weights := fuzzyDM.Weights[key]
			weightedFuzzyNum := types.FuzzyNumber{
				A: value.A * weights.A,
				B: value.B * weights.B,
				C: value.C * weights.C,
			}
			fuzzyDM.Data[k][key] = weightedFuzzyNum
		}
	}
}

func weightIdeals(fuzzyDM *types.FuzzyDecisionMatrix) {
	for key, value := range fuzzyDM.PositiveIdeals {
		weights := fuzzyDM.Weights[key]
		weightedFuzzyNum := types.FuzzyNumber{
			A: value.A * weights.A,
			B: value.B * weights.B,
			C: value.C * weights.C,
		}
		fuzzyDM.PositiveIdeals[key] = weightedFuzzyNum
	}

	for key, value := range fuzzyDM.NegativeIdeals {
		weights := fuzzyDM.Weights[key]
		weightedFuzzyNum := types.FuzzyNumber{
			A: value.A * weights.A,
			B: value.B * weights.B,
			C: value.C * weights.C,
		}
		fuzzyDM.NegativeIdeals[key] = weightedFuzzyNum
	}
}

// will apply manual requests ontop of the current telemetry
// for nodes scheduled in the last minute. This will stop nodes
// getting over filled when multiple pods scheduled at once.
func ApplyManualRequests(fuzzyDM *types.FuzzyDecisionMatrix, clusterLimits types.ClusterInfo) {

	allMetrics := telemetry.GetFullCache()

	for node, metrics := range allMetrics {

		_, exists := fuzzyDM.Data[node]
		if !exists {
			// means our node got filtered out previously
			continue
		}
		// iterate over last scheduled
		for _, pod := range metrics.PodsScheduled {
			if time.Since(pod.Timestamp) < time.Second*40 {
				// pod was scheduled within last minute so we must manually add
				// the requests to the current decision matrix values
				cpuPercent := calculateManualRequest(clusterLimits.CPULimits[node], pod.Requests.CPU)
				ramPercent := calculateManualRequest(clusterLimits.RAMLimits[node], pod.Requests.RAM)

				// add this onto the fuzzyDM
				cpuMetric := fuzzyDM.Data[node]["CPU"]
				cpuMetric.A += cpuPercent
				cpuMetric.B += cpuPercent
				cpuMetric.C += cpuPercent
				fuzzyDM.Data[node]["CPU"] = cpuMetric

				// add this onto the fuzzyDM
				ramMetric := fuzzyDM.Data[node]["RAM"]
				ramMetric.A += ramPercent
				ramMetric.B += ramPercent
				ramMetric.C += ramPercent
				fuzzyDM.Data[node]["RAM"] = ramMetric
			}
		}
	}
}

func calculateManualRequest(limit int64, request int64) float64 {
	return (float64(request) / float64(limit)) * 100
}
