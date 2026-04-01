package dashboard

import (
	"encoding/json"
	"fmt"
	"scheduler/pkg/types"
	"time"
)

// This file is the hub which will take messages from the scheduler and pass them to the web UI

type HubMessage struct {
	// subject tells the web UI which component data is meant for
	Subject string `json:"subject"`

	// payload is the actual data of the json body
	Payload interface{} `json:"payload"`
}

type PodScheduledMessage struct {
	// name of the selected node
	NodeName string `json:"nodeName"`

	// requests of the selected node
	CPURequests int64 `json:"cpuRequest"`
	RAMRequests int64 `json:"ramRequest"`

	// current value of the telemetry cache
	TelemetryCache map[string]types.NodeTelemetryMetrics `json:"telemetryCache"`

	// initial fuzzy Decision Matrix
	InitialFuzzyDM types.FuzzyDecisionMatrix `json:"initialFuzzyDM"`

	// fuzzy DM after having manual requests added
	ManualFuzzyDM types.FuzzyDecisionMatrix `json:"manualFuzzyDM"`

	// Fuzzy DM after filtering
	FilteredFuzzyDM types.FuzzyDecisionMatrix `json:"filteredFuzzyDM"`

	// Fuzzy DM after weighting
	WeightedFuzzyDM types.FuzzyDecisionMatrix `json:"weightedFuzzyDM"`

	// final node scores
	NodeScores map[string]float64 `json:"nodeScores"`
}

func PublishScheduleUpdate(data PodScheduledMessage) {
	wrapper := HubMessage{
		Subject: "scheduling_update",
		Payload: data,
	}

	jsonBytes, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	broadcastToWS(jsonBytes)
}

type TelemetryCacheMessage struct {
	UnfilteredData map[string]types.NodeTelemetryMetrics `json:"unfilteredCache"`
	FilteredData   map[string]types.NodeTelemetryMetrics `json:"filteredCache"`
}

func PublishTelemetryCache(data TelemetryCacheMessage) {
	wrapper := HubMessage{
		Subject: "telemetry_cache",
		Payload: data,
	}

	jsonBytes, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	broadcastToWS(jsonBytes)
}

type TelemetryLiveMessage struct {
	// map is as follows - data["NodeName"]["CPU"] = val
	Data map[string]map[string]float64

	Timestamp time.Time
}

func PublishTelemetryLive(data TelemetryLiveMessage) {
	wrapper := HubMessage{
		Subject: "live_telemetry",
		Payload: data,
	}

	jsonBytes, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	broadcastToWS(jsonBytes)
}

func broadcastToWS(msg []byte) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		select {
		case client.send <- msg:
		default:
			// if channel is full - do nothing
		}
	}
}

// This is a 'hacky' way to copy a full struct with maps or arrays in it deeply
// this is used for the decision matrix and telemetry cache as the values can change
// between when the variable in our monitoring struct is written and when sent to the user
func JsonCopy[T any](value T) T {
	data, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("Error marshalling: %v", err)
		return value
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		fmt.Printf("Error unmarshalling: %v", err)
		return value
	}
	return result
}
