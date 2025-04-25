# KEDA with Azure Service Bus Scaler, Azure Load Testing, and Grafana

This repository demonstrates how to use [KEDA](https://keda.sh/) with the Azure Service Bus scaler to handle dynamic workloads. It also integrates Azure Load Testing to simulate traffic and Grafana for monitoring.

## Prerequisites

- Azure CLI installed ([Install Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli))
- Kubernetes CLI (`kubectl`) installed ([Install kubectl](https://kubernetes.io/docs/tasks/tools/))
- Azure subscription
- Grafana installed or accessible

---

## Steps to Set Up and Test

### 1. Create an AKS Cluster with KEDA Addon

1. Log in to Azure:

```bash
az login
```

1. Create a resource group:

```bash
az group create --name keda-demo-rg --location eastus
```

1. Create an AKS cluster with the KEDA addon enabled:

```bash
az aks create \
   --resource-group keda-demo-rg \
   --name keda-demo-aks \
   --enable-addons keda \
   --generate-ssh-keys
```

1. Connect to the AKS cluster:

```bash
az aks get-credentials --resource-group keda-demo-rg --name keda-demo-aks
```

---

### 2. Create an Azure Service Bus Namespace and Queue

1. Create a Service Bus namespace:
   
```bash
az servicebus namespace create \
   --resource-group keda-demo-rg \
   --name keda-demo-sb \
   --location eastus \
   --sku Standard
```

1. Create a Service Bus queue:

```bash
az servicebus queue create \
   --resource-group keda-demo-rg \
   --namespace-name keda-demo-sb \
   --name demo-queue
```

1. Retrieve the connection string for the namespace:

```bash
SB_CONNECTION_STRING=$(az servicebus namespace authorization-rule keys list \
   --resource-group keda-demo-rg \
   --namespace-name keda-demo-sb \
   --name RootManageSharedAccessKey \
   --query primaryConnectionString \
   --output tsv)
```

Save this connection string for later use.


### 3. Deploy KEDA ScaledObject

1. Create a Kubernetes secret for the Service Bus connection string:

```bash
kubectl create secret generic servicebus-secret \
   --from-literal=connectionString=${SB_CONNECTION_STRING}
```

1. Deploy the KEDA `ScaledObject` to scale based on the Service Bus queue:

```yaml
# scaledobject.yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
   name: servicebus-scaledobject
spec:
   scaleTargetRef:
      name: <your-deployment-name>
   triggers:
   - type: azure-servicebus
      metadata:
      queueName: demo-queue
      connectionFromEnv: SERVICEBUS_CONNECTION_STRING
```

Apply the configuration:

```bash
kubectl apply -f scaledobject.yaml
```

### 4. Test with Azure Load Testing

Create a Keyvault to store the connection string:

```bash
az keyvault create \
   -n ${CLUSTER_NAME}999 \
   --resource-group $CLUSTER_RG \
   --location $LOCATION \
   --enable-rbac-authorization false
```

Add yourself as a Keyvault admin:

```bash
az keyvault set-policy \
   --name ${CLUSTER_NAME}999 \
   --resource-group $CLUSTER_RG \
   --upn $(az ad signed-in-user show --query userPrincipalName -o tsv) \
   --secret-permissions get list set delete \
   --enable-soft-delete true \
   --soft-delete-retention-days 7
```
Store the connection string in the Keyvault:


```bash
PRIMARY_KEY=$(az servicebus queue authorization-rule keys list -g $CLUSTER_RG \
--namespace-name keda-demo-azcloudative \
--queue-name demo-queue -n queuerule \
--query "primaryKey" -o tsv)

az keyvault secret set \
   --vault-name ${CLUSTER_NAME}999 \
   --name sendAccessKey \
   --value $PRIMARY_KEY -o none
```

1. Create an Azure Load Testing resource:

```bash
az load create -g $CLUSTER_RG \
-l $LOCATION \
-n keda-test \
--identity-type SystemAssigned \
--no-wait
```

Add the test maua

Get the Azure Load Test System Assigned Identity:

```bash
LOAD_TEST_IDENTITY=$(az load show \
   --name keda-test -o json \
   --resource-group $CLUSTER_RG \
--query identity.principalId -o tsv)
```
Assign the Keyvault access policy to the Load Test System Assigned Identity:

```bash
az keyvault set-policy \
   --name ${CLUSTER_NAME}999 \
   --resource-group $CLUSTER_RG \
   --object-id $LOAD_TEST_IDENTITY \
   --secret-permissions get list
```


1. Configure a load test to send messages to the Service Bus queue. Use a tool like Apache JMeter or a custom script to generate traffic.

2. Run the load test and observe the scaling behavior of your AKS deployment.


### 5. Monitor with Grafana

Create an Azure Managed Grafana instance or use an existing one. Follow these steps to visualize the metrics:
1. Create a Grafana instance in Azure:


```bash
# make sure you have the Azure Managed Grafana extension installed with az extension add --name amg
az grafana create \
   --resource-group $CLUSTER_RG\
   --name keda-grafana \
   --location $LOCATION \
   --sku Standard
```


1. Import the provided Grafana dashboard JSON file located in the `grafana` folder.

2. Update the data source in Grafana to point to your Azure Monitor or Service Bus metrics.

3. Visualize metrics such as incoming messages, message count, and scaling behavior.

## Clean Up Resources

To avoid incurring unnecessary costs, delete the resource group when you're done:
```bash
az group delete --name keda-demo-rg --yes --no-wait
```

## Repository Structure

- `grafana/`: Contains Grafana dashboard JSON files for monitoring.
- `scaledobject.yaml`: Example KEDA `ScaledObject` configuration.
- `README.md`: Documentation for setting up and testing the demo.



## References

- [KEDA Documentation](https://keda.sh/docs/)
- [Azure Service Bus Documentation](https://learn.microsoft.com/en-us/azure/service-bus-messaging/)
- [Azure Load Testing Documentation](https://learn.microsoft.com/en-us/azure/load-testing/)
- [Grafana Documentation](https://grafana.com/docs/)