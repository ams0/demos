# Networking with KAITO

kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/experimental-install.yaml

istioctl install --set profile=ambient --skip-confirmation




apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: model-gateway
spec:
  gatewayClassName: istio
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: Same
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: gateway-http-route
spec:
  parentRefs:
  - name: model-gateway
  rules:
  - matches:
    - path:
        type: Exact
        value: /v1/completions
    backendRefs:
    - name: workspace-phi-4-mini-instruct
      port: 80