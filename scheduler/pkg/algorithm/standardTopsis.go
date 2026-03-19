package algorithm

import (
	"math"
	"scheduler/pkg/types"
)

func TopsisSelectNode(fuzzyDM types.FuzzyDecisionMatrix) string {
	weightNodes(&fuzzyDM)
	weightIdeals(&fuzzyDM)

	nodeScores := topsisScoreNodes(fuzzyDM)
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

// standard topsis will just look at the CPU and RAM
func topsisScoreNodes(fuzzyDM types.FuzzyDecisionMatrix) map[string]float64 {

	nodeScores := map[string]float64{}

	for node, criterion := range fuzzyDM.Data {
		// iterate over all criteria in each node
		negativeDists := float64(0)
		positiveDists := float64(0)
		for criteria, value := range criterion {
			if criteria == "CPU" || criteria == "RAM" {
				// we only care about the 'b' values as this is standard topsis
				fuzzyNum := value
				positiveIdeal := fuzzyDM.PositiveIdeals[criteria]
				negativeIdeal := fuzzyDM.NegativeIdeals[criteria]
				positiveDists += (fuzzyNum.B - positiveIdeal.B) * (fuzzyNum.B - positiveIdeal.B)
				negativeDists += (fuzzyNum.B - negativeIdeal.B) * (fuzzyNum.B - negativeIdeal.B)
			}
		}
		// score node
		nodeScore := negativeDists / (negativeDists + positiveDists)
		nodeScores[node] = nodeScore
	}
	return nodeScores
}
