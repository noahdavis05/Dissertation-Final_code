package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"scheduler/pkg/dashboard"
	"scheduler/pkg/scheduler"
	"scheduler/pkg/telemetry"
)

func main() {
	mux := http.NewServeMux()
	// serve the react frontend
	mux.Handle("/", http.FileServer(dashboard.GetFileSystem()))

	// run the server in go routine
	// this just serves the react app
	go func() {
		fmt.Println("Dashboard available at http://localhost:8080")
		http.ListenAndServe(":8080", mux)
	}()
	fmt.Println("Starting Scheduler")

	// start the websocket server
	// this will serve data to the react app
	go dashboard.StartServer()

	// load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "/home/noah/.kube/config")
	if err != nil {
		fmt.Println("Error loading config")
	}

	// running in cluster
	/*
		config, err := rest.InClusterConfig()
		if err != nil {
			fmt.Println("Error creating config")
		}
	*/

	// create a kubernetes clientset from the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating clientset")
	}

	// create a Kubernetes informer factory
	// this creates informers
	factory := informers.NewSharedInformerFactory(
		client,
		time.Minute,
	)

	// make a pod informer with factory
	podInformer := factory.Core().V1().Pods().Informer()
	nodeInformer := factory.Core().V1().Nodes()

	// make a node Lister
	nodeLister := nodeInformer.Lister()

	// add event handlers to the informer
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// pod is currently a generic object so convert it into
			// actual Pod type
			pod := obj.(*corev1.Pod)

			if pod.Spec.NodeName == "" && (pod.Spec.SchedulerName == "fuzzy-topsis-scheduler" || pod.Spec.SchedulerName == "topsis-scheduler" || pod.Spec.SchedulerName == "custom-fuzzy-topsis-scheduler") {
				fmt.Println("Unscheduled pod detected")
				scheduler.SchedulePod(client, pod, nodeLister, pod.Spec.SchedulerName)
			}
		},
	})

	// create stop channel
	stopCh := make(chan struct{})

	// create a channel to handle shutdown signals
	// this is used to gracefully stop all goroutines
	sigCh := make(chan os.Signal, 1)

	signal.Notify(
		sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	// function to recieve shutdown signal and stop the stopChannel
	// this will stop subroutines using stopChannel and avoid
	// goroutine leaks
	go func() {
		sig := <-sigCh
		fmt.Println("Shutdown signal received:", sig)
		close(stopCh)
	}()

	// start informer
	factory.Start(stopCh)
	cache.WaitForCacheSync(stopCh, podInformer.HasSynced, nodeInformer.Informer().HasSynced)
	fmt.Println("Informer has started")
	// start the telemetry refresher in the background
	go telemetry.AutoRefreshTelemetryCache(stopCh, 10*time.Second, nodeLister)
	<-stopCh // keeps program running indefinitely

}
