package algorithm

import (
	"math"
	"scheduler/pkg/types"
)

func FuzzyTopsisSelectNode(fuzzyDM types.FuzzyDecisionMatrix) string {
	weightNodes(&fuzzyDM)
	weightIdeals(&fuzzyDM)

	nodeScores := fuzzyTopsisScoreNodes(fuzzyDM)
	nodeName := ""
	maxScore := -math.Inf(1)
	for node, score := range nodeScores {
		if score > maxScore {
			maxScore = score
			nodeName = node
		}
	}
	return nodeName
}

func fuzzyTopsisScoreNodes(fuzzyDM types.FuzzyDecisionMatrix) map[string]float64 {
	nodeScores := map[string]float64{}

	for node, criterion := range fuzzyDM.Data {
		negativeDists := float64(0)
		positiveDists := float64(0)
		for criteria, value := range criterion {
			if criteria == "CPU" || criteria == "RAM" {
				fuzzyNum := value
				positiveIdeal := fuzzyDM.PositiveIdeals[criteria]
				negativeIdeal := fuzzyDM.NegativeIdeals[criteria]
				positiveDists += calculateDistance(fuzzyNum, positiveIdeal)
				negativeDists += calculateDistance(fuzzyNum, negativeIdeal)
			}
		}
		nodeScore := negativeDists / (negativeDists + positiveDists)
		nodeScores[node] = nodeScore
	}
	return nodeScores
}
