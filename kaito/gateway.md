cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: chat
  namespace: default
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
spec:
  gatewayClassName: eg
  listeners:
  - name: https
    hostname: "*.cloudnative.computer"
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: chat-cloudnative-computer-tls  # Created by cert-manager
    allowedRoutes:
      namespaces:
        from: All
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: chat
  namespace: default
spec:
  parentRefs:
    - name: chat
      namespace: default
      sectionName: https
  hostnames:
    - chat.cloudnative.computer
  rules:
    - backendRefs:
        - name: inference-frontend
          port: 8000
          namespace: default

apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: model
  namespace: default
spec:
  parentRefs:
    - name: chat
      namespace: default
      sectionName: https
  hostnames:
    - model.cloudnative.computer
  rules:
    - backendRefs:
        - name: workspace-phi-4-mini-instruct	
          port: 80
          namespace: default
EOF