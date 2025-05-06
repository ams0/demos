# Deploy KAITO on AKS with GPU Provisioner

## Pre-requisites
- Azure CLI
- Azure subscription
- GPU SKU available in the region and within quota limits

az feature register --namespace "Microsoft.ContainerService" --name "NodeAutoProvisioningPreview"
az feature register --namespace "Microsoft.ContainerService" --name "AIToolchainOperatorPreview"


export RG=kaito
export AZURE_LOCATION=francecentral
export CLUSTER_NAME=kaito
export K8S_VERSION=1.32.2
export ACR_NAME=cndro2025
export SUBSCRIPTION=1c51d1c3-d83d-4d71-ace1-df3496eddac4

az group create --name $RG --location $AZURE_LOCATION

az acr create --resource-group $RG --name $ACR_NAME --location $AZURE_LOCATION --sku Basic


MODELNAME=phi-4-mini-instruct
TAG="0.1.0"

az acr import --no-wait -g $RG --name $ACR_NAME --source  mcr.microsoft.com/aks/kaito/kaito-$MODELNAME:$TAG --image $MODELNAME:$TAG


az aks create --location $AZURE_LOCATION \
    --resource-group $RG \
    --tier Standard \
    --name ${CLUSTER_NAME} \
    --node-count 2 \
    --node-vm-size Standard_B4ms \
    --enable-oidc-issuer \
    --enable-workload-identity \
    --node-provisioning-mode auto \
    --enable-keda \
    --enable-vpa \
    --network-dataplane cilium \
    --network-plugin azure \
    --network-plugin-mode overlay \
    --kubernetes-version $K8S_VERSION \
    --attach-acr $ACR_NAME \
    --location $AZURE_LOCATION \
    --enable-azure-monitor-metrics


az aks get-credentials --resource-group $RG --name ${CLUSTER_NAME} --overwrite-existing

# Get the Cluster Resource Group
export RG_ID=$(az group show -n $RG -o tsv --query id)

# Get the managed cluster Resource Group
export MC_RESOURCE_GROUP=$(az aks show --resource-group ${RG} --name ${CLUSTER_NAME} --query nodeResourceGroup -o tsv)

# Set a variable for the KAITO IDentity name
export KAITO_IDENTITY_NAME="ai-toolchain-operator-${CLUSTER_NAME}"

az identity create --name $KAITO_IDENTITY_NAME -g $RG

# Get the principal ID for the KAITO managed identity
export PRINCIPAL_ID=$(az identity show --name "${KAITO_IDENTITY_NAME}" --resource-group "${RG}" --query 'principalId' -o tsv)

# Grant contributor on the cluster resource group
az role assignment create --role "Contributor" --assignee "${PRINCIPAL_ID}" --scope "/subscriptions/${SUBSCRIPTION}/resourceGroups/${RG}"
# Get the OIDC Issuer URL
export AKS_OIDC_ISSUER=$(az aks show -n $CLUSTER_NAME -g $RG --subscription $SUBSCRIPTION --query "oidcIssuerProfile.issuerUrl" -o tsv)

export KAITO_WORKSPACE_VERSION=0.4.5
export GPU_PROVISIONER_VERSION=0.3.3

helm install kaito-workspace  --set clusterName=$CLUSTER_NAME --wait \
https://github.com/kaito-project/kaito/raw/gh-pages/charts/kaito/workspace-$KAITO_WORKSPACE_VERSION.tgz --namespace kaito-workspace --create-namespace

curl -sO https://raw.githubusercontent.com/Azure/gpu-provisioner/main/hack/deploy/configure-helm-values.sh
chmod +x ./configure-helm-values.sh && ./configure-helm-values.sh $CLUSTER_NAME $RG $KAITO_IDENTITY_NAME
kubectl apply -f https://raw.githubusercontent.com/Azure/karpenter-provider-azure/refs/heads/main/pkg/apis/crds/karpenter.sh_nodeclaims.yaml

helm install gpu-provisioner --values gpu-provisioner-values.yaml --set settings.azure.clusterName=$CLUSTER_NAME --wait \
https://github.com/Azure/gpu-provisioner/raw/gh-pages/charts/gpu-provisioner-$GPU_PROVISIONER_VERSION.tgz --namespace gpu-provisioner --create-namespace

az identity federated-credential create --name kaito-federatedcredential --identity-name $KAITO_IDENTITY_NAME -g $RG --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"gpu-provisioner:gpu-provisioner" --audience api://AzureADTokenExchange --subscription $SUBSCRIPTION

