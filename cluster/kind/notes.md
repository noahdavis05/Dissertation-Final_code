# Kind Cluster

- This cluster is for local development using Kind. (Kubernetes in Docker) to run a multi node Kubernetes cluster on one machine.
- There is a `kind-cluster-config.yaml` file in this directory. This can be used to spin up the Kind cluster with the command `kind create cluster --config kind-cluster-config.yaml`.
- Following this the cluster will require prometheus and grafana for the scheduler. To do this we must run these commands:
  - Ensure helm is installed: `helm repo add prometheus-community https://prometheus-community.github.io/helm-charts` and `helm repo update`
  - Create monitoring namespace: `kubectl create namespace monitoring`
  - Install prometheus via helm: `helm install kube-prom-stack prometheus-community/kube-prometheus-stack --namespace monitoring`
  - Get grafana password using the returned command
  - Port forward grafana to allow access: `kubectl port-forward -n monitoring svc/kube-prom-stack-grafana 9090:9090`
  - `kubectl port-forward -n monitoring svc/kube-prom-stack-kube-prome-prometheus 9090:9090` to port forward for running scheduler locally.


