package telemetry

import (
	"context"
	"fmt"
	"scheduler/pkg/config"
	"scheduler/pkg/dashboard"
	"scheduler/pkg/types"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1listers "k8s.io/client-go/listers/core/v1"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type telemetryQuery struct {
	Query      string
	MetricType string // "CPU" or "RAM"
	Stat       string // "Low", "Mean", "High"
}

func AutoRefreshTelemetryCache(stopCh <-chan struct{}, interval time.Duration, nodeLister v1listers.NodeLister) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nodes, err := nodeLister.List(labels.Everything())
			if err != nil {
				fmt.Printf("failed to list nodes: %v\n", err)
				continue
			}
			fmt.Println("Refreshing telemetry cache")
			RefreshTelemetryCache(nodes)
		case <-stopCh:
			fmt.Println("Telemetry refresher finishing")
			return
		}
	}
}

func RefreshTelemetryCache(nodes []*v1.Node) {

	// make a prometheus client which we can make API calls with
	prometheusClient, err := api.NewClient(api.Config{
		Address: config.PrometheusURL,
	})
	if err != nil {
		fmt.Printf("failed to create Prometheus client: %v\n", err)
	}

	promApi := promv1.NewAPI(prometheusClient)

	refreshedData := requestNodeTelemetry(promApi)

	// the keys are currently the IP adresses of nodes
	// change these back to node.Name
	finalData := make(map[string]types.NodeTelemetryMetrics)
	for _, node := range nodes {
		ip := getNodeAddress(node)
		if ip != "" {
			if metrics, exists := refreshedData[ip]; exists {
				finalData[node.Name] = metrics
			}
		}
	}

	// make a telemetry cache struct for the logs/debug UI
	telCacheMessage := dashboard.TelemetryCacheMessage{}
	telCacheMessage.UnfilteredData = dashboard.JsonCopy(finalData)

	returnData := applyRecentBias(promApi, nodes, finalData)
	telCacheMessage.FilteredData = dashboard.JsonCopy(returnData)

	dashboard.PublishTelemetryCache(telCacheMessage)

	// print out data for debugging
	//jsonData, _ := json.MarshalIndent(returnData, "", "  ")
	//fmt.Println(string(jsonData))
	UpdateCache(returnData)

	// update the live telemetry graphs in logs/debug UI
	streamTelemetryToDashboard(promApi, nodes)
}

func applyRecentBias(promAPI promv1.API, nodes []*v1.Node, currentData map[string]types.NodeTelemetryMetrics) map[string]types.NodeTelemetryMetrics {
	returnData := map[string]types.NodeTelemetryMetrics{}
	for _, node := range nodes {
		// get the old data for this node if it exists
		oldNodeData, ok := GetNodeMetrics(node.Name)
		if !ok {
			// there is no old data for this node
			// this means that the telemetry cache has just started
			returnData[node.Name] = currentData[node.Name]
		} else {
			// check the time since last scheduled
			timeSince := time.Since(oldNodeData.LastScheduled)
			if timeSince > 5*time.Minute {
				// means we treat this normally
				returnData[node.Name] = currentData[node.Name]
			} else {
				if timeSince > 2*time.Minute {
					returnData[node.Name] = subFiveMinBias(promAPI, oldNodeData, node, oldNodeData.LastScheduled)
				} else {
					// we will get the mean from the last 2 mins and use the old range as old range may not have settled
					returnData[node.Name] = subTwoMinBias(promAPI, oldNodeData, node)
				}
			}
		}
	}
	return returnData
}

// Sub two min bias is for nodes which have been scheduled on in the last two minutes
// we get their average CPU for a baseline, but as a result of the recent change in
// workloads we use its old range. This is because it's actual current range is going to be
// very large. This would get this node penalised in the fuzzy TOPSIS when it shouldn't be
// as this large range doesn't signify instability.
func subTwoMinBias(promAPI promv1.API, oldNodeData types.NodeTelemetryMetrics, node *v1.Node) types.NodeTelemetryMetrics {
	oldCPURange := oldNodeData.CPU.High - oldNodeData.CPU.Low
	oldRAMRange := oldNodeData.RAM.High - oldNodeData.RAM.Low
	meanCPU, meanRAM := nodeTelemetryMeans(promAPI, getNodeAddress(node), int64(time.Since(oldNodeData.LastScheduled).Seconds()))

	if meanCPU == 0.0 {
		fmt.Println("NO CPU VAL")
		meanCPU = oldNodeData.CPU.Mean
	}
	if meanRAM == 0.0 {
		fmt.Println("NO RAM VAL")
		meanRAM = oldNodeData.RAM.Mean
	}

	return types.NodeTelemetryMetrics{
		CPU: types.TelemetryMetric{
			Low:  meanCPU - (oldCPURange / 2),
			Mean: meanCPU,
			High: meanCPU + (oldCPURange / 2),
		},
		RAM: types.TelemetryMetric{
			Low:  meanRAM - (oldRAMRange / 2),
			Mean: meanRAM,
			High: meanRAM + (oldRAMRange / 2),
		},
		LastScheduled: oldNodeData.LastScheduled,
		PodsScheduled: oldNodeData.PodsScheduled,
	}
}

// no need to keep the list of scheduled pods in this
// as we only need the ones scheduled in less than the last minute
// if we are in this function there can't have been a pod scheduled
// in the last min.
func subFiveMinBias(promAPI promv1.API, oldNodeData types.NodeTelemetryMetrics, node *v1.Node, lastScheduled time.Time) types.NodeTelemetryMetrics {
	seconds := int64(time.Since(oldNodeData.LastScheduled).Seconds())
	nodeIP := getNodeAddress(node)
	fmt.Println("In sub five mins bias")
	newTelemetry := nodeTelemetryAll(promAPI, nodeIP, seconds)
	newTelemetry.LastScheduled = lastScheduled
	return newTelemetry
}

// function gets just the mean telemetry for RAM and CPU over the last N seconds
// Seconds is rounded down to the minute e.g. 78 seconds = 1 min
func nodeTelemetryMeans(prompApi promv1.API, nodeIP string, seconds int64) (float64, float64) {
	// get minutes rounded down
	// this means we use as little telemetry as possible from before the node was last scheduled
	minutes := max(1, (seconds+59)/60)
	queryRAM := fmt.Sprintf(`avg_over_time((100 * (1 - (node_memory_MemAvailable_bytes{instance=~"%s:.*"} / node_memory_MemTotal_bytes{instance=~"%s:.*"})))[%dm:15s])`, nodeIP, nodeIP, minutes)
	queryCPU := fmt.Sprintf(`avg_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{instance=~"%s:.*", mode="idle"}[1m])) * 100))[%dm:15s])`, nodeIP, minutes)

	// make both of these queries
	resultRAM := makePromRequest(prompApi, queryRAM)
	resultCPU := makePromRequest(prompApi, queryCPU)

	extract := func(v model.Value) float64 {
		if v == nil {
			return 0.0
		}
		vec, ok := v.(model.Vector)
		if !ok || len(vec) == 0 {
			return 0.0
		}
		return float64(vec[0].Value)
	}
	return extract(resultCPU), extract(resultRAM)
}

// Gets all node telemetry for a node over the amount of seconds
// seconds is handled the same as the above function
// this function gets high and low over this period as well as the means
func nodeTelemetryAll(prompApi promv1.API, nodeIP string, seconds int64) types.NodeTelemetryMetrics {
	minutes := 1
	selector := fmt.Sprintf(`instance=~"%s:.*"`, nodeIP)
	queries := []telemetryQuery{
		{fmt.Sprintf(`avg_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{%s, mode="idle"}[1m])) * 100))[%dm:15s])`, selector, minutes), "CPU", "Mean"},
		{fmt.Sprintf(`min_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{%s, mode="idle"}[1m])) * 100))[%dm:15s])`, selector, minutes), "CPU", "Low"},
		{fmt.Sprintf(`max_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{%s, mode="idle"}[1m])) * 100))[%dm:15s])`, selector, minutes), "CPU", "High"},
		{fmt.Sprintf(`avg_over_time((100 * (1 - (node_memory_MemAvailable_bytes{%s} / node_memory_MemTotal_bytes{%s})))[%dm:15s])`, selector, selector, minutes), "RAM", "Mean"},
		{fmt.Sprintf(`min_over_time((100 * (1 - (node_memory_MemAvailable_bytes{%s} / node_memory_MemTotal_bytes{%s})))[%dm:15s])`, selector, selector, minutes), "RAM", "Low"},
		{fmt.Sprintf(`max_over_time((100 * (1 - (node_memory_MemAvailable_bytes{%s} / node_memory_MemTotal_bytes{%s})))[%dm:15s])`, selector, selector, minutes), "RAM", "High"},
	}

	nodeTelemetry := types.NodeTelemetryMetrics{}

	for _, query := range queries {
		result := makePromRequest(prompApi, query.Query)

		resVector := result.(model.Vector)
		value := float64(resVector[0].Value)

		switch query.MetricType {
		case "CPU":
			switch query.Stat {
			case "Mean":
				nodeTelemetry.CPU.Mean = float64(value)
			case "Low":
				nodeTelemetry.CPU.Low = float64(value)
			case "High":
				nodeTelemetry.CPU.High = float64(value)
			}
		case "RAM":
			switch query.Stat {
			case "Mean":
				nodeTelemetry.RAM.Mean = float64(value)
			case "Low":
				nodeTelemetry.RAM.Low = float64(value)
			case "High":
				nodeTelemetry.RAM.High = float64(value)
			}
		}
	}
	return nodeTelemetry
}

func requestNodeTelemetry(promApi promv1.API) map[string]types.NodeTelemetryMetrics {
	queries := []telemetryQuery{
		{`avg by (instance) (100 - (rate(node_cpu_seconds_total{mode="idle"}[3m]) * 100))`, "CPU", "Mean"},
		{`min_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100))[3m:1m])`, "CPU", "Low"},
		{`max_over_time((100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100))[3m:1m])`, "CPU", "High"},
		{`avg_over_time((100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)))[3m:1m])`, "RAM", "Mean"},
		{`min_over_time((100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)))[3m:1m])`, "RAM", "Low"},
		{`max_over_time((100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)))[3m:1m])`, "RAM", "High"},
	}

	// make a results map
	refreshedData := map[string]types.NodeTelemetryMetrics{}

	// iterate over queries and add them into a result
	for _, query := range queries {
		result := makePromRequest(promApi, query.Query)

		resVector := result.(model.Vector)
		// iterate over the resultVector
		for _, sample := range resVector {
			// we will use the ip address as the key
			instance := string(sample.Metric["instance"])
			ip := strings.Split(instance, ":")[0]

			value := sample.Value

			// check if key exists for this IP, if not make it
			if _, exists := refreshedData[ip]; !exists {
				refreshedData[ip] = types.NodeTelemetryMetrics{}
			}

			metrics := refreshedData[ip]

			// now add the values
			switch query.MetricType {
			case "CPU":
				switch query.Stat {
				case "Mean":
					metrics.CPU.Mean = float64(value)
				case "Low":
					metrics.CPU.Low = float64(value)
				case "High":
					metrics.CPU.High = float64(value)
				}
			case "RAM":
				switch query.Stat {
				case "Mean":
					metrics.RAM.Mean = float64(value)
				case "Low":
					metrics.RAM.Low = float64(value)
				case "High":
					metrics.RAM.High = float64(value)
				}
			}
			refreshedData[ip] = metrics
		}
	}

	return refreshedData
}

func makePromRequest(promAPI promv1.API, query string) model.Value {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := promAPI.Query(ctx, query, time.Now())
	if err != nil {
		fmt.Printf("Prometheus query error: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Println("Prometheus query warnings:", warnings)
	}

	return result
}

func getNodeAddress(node *v1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}

func streamTelemetryToDashboard(promAPI promv1.API, nodes []*v1.Node) {
	queries := []struct {
		Query      string
		MetricType string
	}{
		{`100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[1m])) * 100)`, "CPU"},
		{`100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))`, "RAM"},
	}

	message := dashboard.TelemetryLiveMessage{
		Data: make(map[string]map[string]float64),
	}

	for _, query := range queries {
		result := makePromRequest(promAPI, query.Query)
		resVector := result.(model.Vector)

		for _, sample := range resVector {
			instance := string(sample.Metric["instance"])
			ip := strings.Split(instance, ":")[0]
			nodeName := ""
			for _, node := range nodes {
				for _, addr := range node.Status.Addresses {
					if addr.Type == v1.NodeInternalIP {
						if addr.Address == ip {
							nodeName = node.Name
						}
					}
				}
			}

			value := float64(sample.Value)
			if _, exists := message.Data[nodeName]; !exists {
				message.Data[nodeName] = make(map[string]float64)
			}
			message.Data[nodeName][query.MetricType] = value
		}
	}
	message.Timestamp = time.Now()

	dashboard.PublishTelemetryLive(message)
}
