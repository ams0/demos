MACHINE_SIZE=Standard_NC64as_T4_v3
MODELNAME=phi-4-mini-instruct	
TAG="0.1.0"
export ACR_NAME=cndro2025

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-inference-params
data:
  inference_config.yaml: |
    max_probe_steps: 6
    vllm:
      gpu-memory-utilization: 0.98  # Controls GPU memory usage (0.0-1.0)
      tensor-parallel-size: 4        # Number of GPUs for tensor parallelism
      max-model-len: 95000         # Maximum sequence length 
      swap-space: 4                 # CPU swap space in GB
      cpu-offload-gb: 0             # Amount of GPU memory to offload to CPU
---
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  annotations:
    kaito.sh/bypass-resource-checks: "True"
  name: workspace-${MODELNAME}
resource:
  instanceType: "${MACHINE_SIZE}"
  labelSelector:
    matchLabels:
      apps: ${MODELNAME}
inference:
  config: "my-inference-params"
  preset:
    name: "${MODELNAME}"
    presetOptions:
      image: ${ACR_NAME}.azurecr.io/${MODELNAME}:${TAG}
EOF