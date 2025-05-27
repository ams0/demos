helm install eg oci://docker.io/envoyproxy/gateway-helm --version v1.3.2 -n envoy-gateway-system --create-namespace

cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
EOF

helm install --values helm-values/cert-manager.yaml  -n cert-manager --create-namespace cert-manager jetstack/cert-manager

export CLIENT_SECRET="this-is-a-secret"
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    email: alessandro.vozza@linux.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod-issuer-account-key
    solvers:
    - selector:
        dnsZones:
        - 'cloudnative.computer'
      dns01:
        cnameStrategy: Follow
        azureDNS:
          clientID: 62ef1af2-ba49-4dc6-b8d2-ea996700be4d
          clientSecretSecretRef:
          # The following is the secret we created in Kubernetes. Issuer will use this to present challenge to Azure DNS.
            name: azuredns-config
            key: client-secret
          subscriptionID: 1c51d1c3-d83d-4d71-ace1-df3496eddac4
          tenantID: 30c10907-19b2-4fb8-9272-a4a539628560
          resourceGroupName: dns
          hostedZoneName: cloudnative.computer
          # Azure Cloud Environment, default to AzurePublicCloud
          environment: AzurePublicCloud
EOF

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
---
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