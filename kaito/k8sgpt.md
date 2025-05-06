

helm repo add k8sgpt https://charts.k8sgpt.ai/
helm repo update
helm install release k8sgpt/k8sgpt-operator -n k8sgpt-operator-system --create-namespace


kubectl apply -f - << EOF
apiVersion: core.k8sgpt.ai/v1alpha1
kind: K8sGPT
metadata:
  name: k8sgpt-kaito
  namespace: k8sgpt-operator-system
spec:
  ai:
    enabled: true
    model: phi-4-mini-instruct
    backend: localai
    baseUrl: http://workspace-phi-4-mini-instruct.default.svc.cluster.local:80/v1
  noCache: false
  repository: ghcr.io/k8sgpt-ai/k8sgpt
  version: v0.4.1
EOF


stern -n k8sgpt-operator-system pod/k8sgpt-kaito

kubectl get results -n k8sgpt-operator-system -o json | jq .

or use local cli

k8sgpt auth add -b localai -u https://model.cloudnative.computer/v1 -m phi-4-mini-instruct

k8sgpt analyze -n default