package scheduler

import (
	"context"
	"fmt"
	"scheduler/pkg/algorithm"
	"scheduler/pkg/cluster"
	"scheduler/pkg/dashboard"
	"scheduler/pkg/telemetry"
	"scheduler/pkg/types"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
)

func SchedulePod(client *kubernetes.Clientset, pod *corev1.Pod, nodeLister v1listers.NodeLister, schedulerName string) {
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		fmt.Printf("failed to list nodes: %v\n", err)
	}

	// create a PodScheduledMessage struct to record logs of this event
	psm := dashboard.PodScheduledMessage{}

	// get the pod requests and the cluster limits
	podRequest := getPodRequests(pod)
	clusterLimits := cluster.CreateClusterInfo(nodes)

	psm.CPURequests = podRequest.CPU
	psm.RAMRequests = podRequest.RAM
	psm.TelemetryCache = dashboard.JsonCopy(telemetry.GetFullCache())

	fmt.Printf("Pod requests %v CPU and %v RAM\n", podRequest.CPU, podRequest.RAM)

	// create fuzzy decision matrix and print in terminal for debugging
	fuzzyDM := algorithm.BuildFuzzyDM(nodes)
	algorithm.DisplayFuzzyDM(fuzzyDM)
	psm.InitialFuzzyDM = dashboard.JsonCopy(fuzzyDM)

	// filter the nodes
	algorithm.FilterNodes(&fuzzyDM, podRequest, clusterLimits)
	psm.FilteredFuzzyDM = dashboard.JsonCopy(fuzzyDM)

	// run the selection using correct algorithm either fuzzy-topsis, topsis, or custom fuzzy-topsis
	// this is used for casting comparisons in my report. These algorithms will use the same decision matrix
	// as it contains all the correct telemetry, but will only use the necassary elements from it.
	selectedNodeName := ""
	switch schedulerName {
	case "fuzzy-topsis-scheduler":
		fmt.Println("Standard fuzzy TOPSIS")
	case "topsis-scheduler":
		fmt.Println("TOPSIS scheduler")
		selectedNodeName = algorithm.TopsisSelectNode(fuzzyDM)
	case "custom-fuzzy-topsis-scheduler": // this is the final algorithm which I came up with
		fmt.Println("Custom fuzzy TOPSIS scheduler")
		selectedNodeName = algorithm.SelectNode(fuzzyDM, &psm)
	}

	psm.NodeName = selectedNodeName

	fmt.Printf("Selected Node : %v\n", selectedNodeName)
	telemetry.PodScheduled(selectedNodeName)
	bindPod(client, pod, selectedNodeName)

	// now send this to the web UI via websocket
	dashboard.PublishScheduleUpdate(psm)
}

// Bind a pod to a Node
func bindPod(client *kubernetes.Clientset, pod *corev1.Pod, nodeName string) {
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			UID:       pod.UID,
		},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: nodeName,
		},
	}

	err := client.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error binding the pod")
	}

}

func getPodRequests(pod *corev1.Pod) types.PodRequest {
	var totalCPU resource.Quantity
	var totalMem resource.Quantity

	// get the requests for all containers within pod
	for _, container := range pod.Spec.Containers {
		totalCPU.Add(*container.Resources.Requests.Cpu())
		totalMem.Add(*container.Resources.Requests.Memory())
	}

	return types.PodRequest{
		CPU: totalCPU.MilliValue(),
		RAM: totalMem.Value(),
	}
}
