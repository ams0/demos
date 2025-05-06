# VLLM Metrics
# VLLM provides a set of metrics that can be scraped by Prometheus. To enable this, you need to set up a ServiceMonitor resource in your Kubernetes cluster.

Enable Managed Prometheus in your AKS cluster. This can be done using the Azure CLI or the Azure portal. Make sure to enable the "Enable managed Prometheus" option when creating or updating your AKS cluster. 

Enable Managed Grafana in your AKS cluster. This can also be done using the Azure CLI or the Azure portal. Make sure to enable the "Enable managed Grafana" option when creating or updating your AKS cluster.

Import the VLLM metrics into your Grafana dashboard. You can do this by creating a new dashboard in Grafana importing the file `grafana-dashboard.json` that contains the VLLM metrics. You can find this file in the VLLM GitHub repository or create it manually.

Collect vLLM metrics using Prometheus. You can do this by creating a ServiceMonitor resource in your Kubernetes cluster that specifies the vLLM service and the metrics endpoint.

you have to label the service:

kubeclt label service vllm app=vllm

Then create the ServiceMonitor resource:
```bash
kubectl apply -f - <<EOF
apiVersion: azmonitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: vllm-metrics
spec:
  labelLimit: 63
  labelNameLengthLimit: 511
  labelValueLengthLimit: 1023
  selector:
    matchLabels:
      app: vllm
  endpoints:
  - port: http
    path: /metrics
EOF
```

or use podmonitor:

```bash
kubectl apply -f - <<EOF
apiVersion: azmonitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: workspace-phi-4-mini-instruct
  labels:
    release: prometheus  # must match your Prometheus instance's label selector
spec:
  namespaceSelector:
    matchNames:
      - default
  selector:
    matchLabels:
      kaito.sh/workspace: workspace-phi-4-mini-instruct
  podMetricsEndpoints:
    - path: /metrics   # change if your metrics endpoint differs
      targetPort: 5000
      interval: 15s
EOF
```