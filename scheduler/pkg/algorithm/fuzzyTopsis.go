package algorithm

import (
	"encoding/json"
	"fmt"
	"math"
	"scheduler/pkg/dashboard"
	"scheduler/pkg/types"
)

// TODO - In future need to update this to filter based on the pod's request
// function tkaes the pointer to fuzzy decision matrix and filters out
// all nodes which are no good. E.g. Nodes which are over the Negative
// ideal limits.
func FilterNodes(fuzzyDM *types.FuzzyDecisionMatrix, podRequests types.PodRequest, clusterLimits types.ClusterInfo) {
	// list of nodes we will remove
	nodeNames := []string{}
	for nodeName, _ := range fuzzyDM.Data {
		if filterNode(fuzzyDM, nodeName, podRequests, clusterLimits) {
			nodeNames = append(nodeNames, nodeName)
		}
	}
	for _, name := range nodeNames {
		delete(fuzzyDM.Data, name)
	}
}

func selectNode(fuzzyDM types.FuzzyDecisionMatrix, scheduleData *dashboard.PodScheduledMessage, debug bool) string {
	// all our values in fuzzyDM are percentages e.g. between 0 and 100
	// therefore already normalised/on same scale
	weightNodes(&fuzzyDM)
	weightIdeals(&fuzzyDM)
	if debug {
		DisplayFuzzyDM(fuzzyDM)
	}

	scheduleData.WeightedFuzzyDM = fuzzyDM

	nodeScores := scoreNodes(fuzzyDM)
	scheduleData.NodeScores = dashboard.JsonCopy(nodeScores)

	if debug {
		fmt.Println("Node scores:")
		data, _ := json.MarshalIndent(nodeScores, "", "  ")
		fmt.Println(string(data))
	}

	// now get the key of the node with the highest value
	nodeName := ""
	maxScore := -math.Inf(1)
	for node, score := range nodeScores {
		if score > maxScore {
			maxScore = score
			nodeName = node
		}
	}
	//
	return nodeName
}

// wrapper function which allows me to set debug mode or not
// useful for running tests
func SelectNode(fuzzyDM types.FuzzyDecisionMatrix, scheduleData *dashboard.PodScheduledMessage) string {
	return selectNode(fuzzyDM, scheduleData, false)
}

func scoreNodes(fuzzyDM types.FuzzyDecisionMatrix) map[string]float64 {
	// iterate over the fuzzyDM and score each node
	nodeScores := map[string]float64{}

	// iterate over the nodes and score their positive and negative distances
	for node, criterion := range fuzzyDM.Data {
		// iterate over all criteria in each node
		negativeDists := float64(0)
		positiveDists := float64(0)
		for criteria, value := range criterion {
			fuzzyNum := value
			positiveIdeal := fuzzyDM.PositiveIdeals[criteria]
			negativeIdeal := fuzzyDM.NegativeIdeals[criteria]
			positiveDists += calculateDistance(fuzzyNum, positiveIdeal)
			negativeDists += calculateDistance(fuzzyNum, negativeIdeal)
		}
		// now with positive and negative dists
		nodeScore := negativeDists / (negativeDists + positiveDists)
		nodeScores[node] = nodeScore
	}
	return nodeScores
}

func calculateDistance(fuzzyNum types.FuzzyNumber, fuzzyIdeal types.FuzzyNumber) float64 {
	dist1 := (fuzzyNum.A - fuzzyIdeal.A) * (fuzzyNum.A - fuzzyIdeal.A)
	dist2 := (fuzzyNum.B - fuzzyIdeal.B) * (fuzzyNum.B - fuzzyIdeal.B)
	dist3 := (fuzzyNum.C - fuzzyIdeal.C) * (fuzzyNum.C - fuzzyIdeal.C)

	totalSquaredDistances := (dist1 + dist2 + dist3) / 3

	return math.Sqrt(totalSquaredDistances)
}
