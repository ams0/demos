# Deploy a small model (phi-4-mini-instruct) on AKS with GPU Provisioner

MACHINE_SIZE=Standard_NC64as_T4_v3
MODELNAME=phi-4-mini-instruct	
TAG="0.1.0"
export ACR_NAME=webinar2025

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
    kaito.sh/bypass-resource-checks: "true"
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


# Deploy a large model (deepseek-r1) on AKS with GPU Provisioner


MACHINE_SIZE=Standard_NC24ads_A100_v4
MODELNAME=deepseek-r1-distill-llama-8b
TAG="0.1.0"
export ACR_NAME=webinar2025

az acr import --no-wait -g $RG --name $ACR_NAME --source  mcr.microsoft.com/aks/kaito/kaito-$MODELNAME:$TAG --image $MODELNAME:$TAG

kubectl apply -f - <<EOF
apiVersion: kaito.sh/v1beta1
kind: Workspace
metadata:
  name: workspace-deepseek-r1-distill-llama-8b
resource:
  instanceType: "Standard_NC24ads_A100_v4"
  labelSelector:
    matchLabels:
      apps: deepseek-r1-distill-llama-8b
inference:
  preset:
    name: "deepseek-r1-distill-llama-8b"
EOF


# Deploy a CPU-friendly model

kubectl apply -f - <<EOF
apiVersion: kaito.sh/v1beta1
kind: Workspace
metadata:
  name: workspace-custom-llm
  annotations:
    kaito.sh/bypass-resource-checks: "true"
resource:
  instanceType: "Standard_D4s_v3" #because of https://github.com/kaito-project/kaito/blob/f6a1f152266589281446150e6fce5f16f7bae07a/test/e2e/preset_test.go#L178
  labelSelector:
    matchLabels:
      apps: cpu-model
inference:
  template:
    spec:
      containers:
        - name: custom-llm-container
          image: ghcr.io/kaito-project/kaito/llm-reference-preset:latest
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /health
              port: 5000
              scheme: HTTP
            initialDelaySeconds: 600
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /health
              port: 5000
              scheme: HTTP
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          command:
            - "accelerate"
          args:
            - "launch"
            - "--num_processes"
            - "1"
            - "--num_machines"
            - "1"
            - "--gpu_ids"
            - "all"
            - "inference_api.py"
            - "--pipeline"
            - "text-generation"
            - "--trust_remote_code"
            - "--allow_remote_files"
            - "--pretrained_model_name_or_path"
            - "distilbert/distilgpt2"
            - "--torch_dtype"
            - "float16" 
          volumeMounts:
            - name: dshm
              mountPath: /dev/shm
      volumes:
      - name: dshm
        emptyDir:
          medium: Memory
EOF


helm install inference-frontend ./charts/DemoUI/inference --set env.workspaceServiceURL="http://workspace-custom-llm:80" --set env.runtime=transformers