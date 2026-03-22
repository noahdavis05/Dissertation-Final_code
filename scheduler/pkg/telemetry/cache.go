package telemetry

import (
	"fmt"
	"scheduler/pkg/types"
	"sync"
	"time"
)

// this is the telemetry data cache
// it has a read write mutex to allow multiple
// reads but only one write at a time
type TelemetryCache struct {
	sync.RWMutex
	data map[string]types.NodeTelemetryMetrics
}

// create a global instance of the cache
var globalCache = &TelemetryCache{
	data: make(map[string]types.NodeTelemetryMetrics),
}

func UpdateCache(newData map[string]types.NodeTelemetryMetrics) {
	globalCache.Lock() // full lock to write
	defer globalCache.Unlock()
	globalCache.data = newData
}

func GetNodeMetrics(nodeName string) (types.NodeTelemetryMetrics, bool) {
	globalCache.RLock() // read lock
	defer globalCache.RUnlock()
	val, ok := globalCache.data[nodeName]
	return val, ok
}

func GetFullCache() map[string]types.NodeTelemetryMetrics {
	globalCache.RLock() // read lock
	defer globalCache.RUnlock()
	return globalCache.data
}

func PodScheduled(nodeName string, pod types.PodRequest) {
	globalCache.Lock() // full lock to write
	defer globalCache.Unlock()

	metrics := globalCache.data[nodeName]
	metrics.LastScheduled = time.Now()
	metrics.PodsScheduled = append(metrics.PodsScheduled, types.ScheduledPod{
		Requests:  pod,
		Timestamp: time.Now(),
	})
	globalCache.data[nodeName] = metrics
	fmt.Printf("Updated Node: %v scheduled time. Node looks like : %v\n", nodeName, globalCache.data[nodeName])
}
