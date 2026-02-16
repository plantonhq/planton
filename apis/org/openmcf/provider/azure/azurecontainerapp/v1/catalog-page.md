# Azure Container App

Deploys an Azure Container App within a Container App Environment, providing a fully managed serverless container platform with configurable containers, init containers, autoscaling (HTTP, TCP, Azure Queue, custom KEDA rules), ingress with traffic splitting, Dapr sidecar integration, secrets management, container registry authentication, managed identity, volumes, and health probes. This is the most feature-rich Azure component in OpenMCF.

## What Gets Created

When you deploy an AzureContainerApp resource, OpenMCF provisions:

- **Container App** -- a `containerapp.App` resource within the specified Container App Environment, configured with containers, scaling rules, ingress, secrets, and operational settings
- **Containers** -- one or more application containers with configurable image, CPU, memory, environment variables, volume mounts, and health probes (liveness, readiness, startup)
- **Init Containers** -- optional initialization containers that run to completion before application containers start
- **Ingress** -- optional HTTP/TCP ingress with traffic weights, IP security restrictions, CORS policy, and mTLS
- **Scale Rules** -- optional autoscaling via HTTP concurrent requests, TCP connections, Azure Queue length, or custom KEDA triggers
- **Dapr Sidecar** -- optional Dapr integration for service invocation, pub/sub, state management, and bindings
- **Managed Identity** -- optional SystemAssigned, UserAssigned, or both for credential-free Azure service authentication

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Container App Environment** where the app will run (can reference an AzureContainerAppEnvironment resource)
- **Container images** accessible from the specified registries -- public registries, ACR with credentials, or ACR with managed identity
- **An Azure Resource Group** is not required directly -- the Container App inherits its resource group and region from the Container App Environment

## Quick Start

Create a file `containerapp.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureContainerApp.my-app
spec:
  resourceGroup: my-rg
  name: my-app
  containerAppEnvironmentId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.App/managedEnvironments/my-env
  containers:
    - name: app
      image: mcr.microsoft.com/k8se/quickstart:latest
      cpu: 0.25
      memory: 0.5Gi
  ingress:
    external: true
    targetPort: 80
```

Deploy:

```shell
openmcf apply -f containerapp.yaml
```

This creates a Container App with a single container, external HTTP ingress on port 80, and Single revision mode.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Container App name. Must be lowercase, hyphens allowed, 2-32 characters. | Required, pattern `^[a-z][a-z0-9-]{0,30}[a-z0-9]$` |
| `containerAppEnvironmentId` | `StringValueOrRef` | Container App Environment ID. Can reference an AzureContainerAppEnvironment resource via `valueFrom`. | Required |
| `containers` | `list` | Application containers. At least one required. Each has `name`, `image`, `cpu`, `memory`. | Minimum 1 item |
| `containers[].name` | `string` | Container name (required) | Required |
| `containers[].image` | `string` | Container image (required) | Required |
| `containers[].cpu` | `double` | CPU cores (required, e.g., `0.25`, `0.5`, `1.0`, `2.0`) | Required |
| `containers[].memory` | `string` | Memory (required, e.g., `0.5Gi`, `1Gi`, `4Gi`) | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `revisionMode` | `string` | `"Single"` | Revision mode: `Single` (latest only) or `Multiple` (traffic splitting). |
| `workloadProfileName` | `string` | -- | Workload profile name for dedicated compute. Must match a profile in the environment. |
| `maxInactiveRevisions` | `int` | -- | Maximum inactive revisions to retain. |
| `minReplicas` | `int` | -- | Minimum replicas (0 enables scale-to-zero). |
| `maxReplicas` | `int` | -- | Maximum replicas for autoscaling. |
| `revisionSuffix` | `string` | -- | Custom suffix for the revision name. |
| `containers[].env` | `list` | `[]` | Environment variables with `name` and either `value` or `secretRef`. |
| `containers[].probes` | `list` | `[]` | Health probes (liveness, readiness, startup) with HTTP or TCP checks. |
| `containers[].volumeMounts` | `list` | `[]` | Volume mounts with `name` and `mountPath`. |
| `initContainers` | `list` | `[]` | Init containers that run before app containers. |
| `volumes` | `list` | `[]` | Volumes: `EmptyDir` or `AzureFile` (with storage account and share). |
| `httpScaleRules` | `list` | `[]` | HTTP autoscale rules with `name` and `concurrentRequests`. |
| `tcpScaleRules` | `list` | `[]` | TCP autoscale rules with `name` and `concurrentRequests`. |
| `azureQueueScaleRules` | `list` | `[]` | Azure Queue autoscale rules with `name`, `queueName`, `queueLength`. |
| `customScaleRules` | `list` | `[]` | Custom KEDA autoscale rules with `name`, `customRuleType`, `metadata`. |
| `secrets` | `list` | `[]` | Secrets: plain `value` or Key Vault reference (`keyVaultUrl` + `identity`). |
| `registries` | `list` | `[]` | Container registries with `server` and authentication (username/password or identity). |
| `ingress` | `object` | -- | Ingress configuration. See ingress fields below. |
| `dapr` | `object` | -- | Dapr sidecar with `appId`, `appPort`, `appProtocol`. |
| `identity.type` | `string` | -- | Managed identity: `SystemAssigned`, `UserAssigned`, or `SystemAssigned,UserAssigned`. |

**Ingress fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `external` | `bool` | `false` | Allow traffic from outside the Container App Environment. |
| `targetPort` | `int` | -- | Port the container listens on (required when ingress is set). |
| `transport` | `string` | `"auto"` | Transport protocol: `auto`, `http`, `http2`, `tcp`. |
| `exposedPort` | `int` | -- | External port for TCP transport. |
| `allowInsecureConnections` | `bool` | `false` | Allow HTTP traffic (disable HTTPS redirect). |
| `trafficWeights` | `list` | `[]` | Traffic weight rules for revision-based routing. |
| `ipSecurityRestrictions` | `list` | `[]` | IP-based access restrictions. |
| `corsPolicy` | `object` | -- | CORS configuration with `allowedOrigins`, `allowedMethods`, `allowedHeaders`. |

## Examples

### Web Service with HTTP Autoscaling

A web service with external ingress, HTTP-based autoscaling, and environment variables:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: web-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerApp.web-service
spec:
  resourceGroup: prod-rg
  name: web-service
  containerAppEnvironmentId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.App/managedEnvironments/prod-env
  minReplicas: 1
  maxReplicas: 10
  containers:
    - name: app
      image: myregistry.azurecr.io/myorg/web-service:v1.0.0
      cpu: 0.5
      memory: 1Gi
      env:
        - name: NODE_ENV
          value: production
        - name: DB_CONNECTION
          secretRef: db-conn
      probes:
        - type: liveness
          httpGet:
            path: /healthz
            port: 3000
          periodSeconds: 10
        - type: readiness
          httpGet:
            path: /ready
            port: 3000
  secrets:
    - name: db-conn
      value: "postgresql://..."
  registries:
    - server: myregistry.azurecr.io
      username: myregistry
      passwordSecretRef: acr-password
  ingress:
    external: true
    targetPort: 3000
  httpScaleRules:
    - name: http-scaling
      concurrentRequests: "50"
```

### Background Worker with Queue Scaling

A background worker that scales based on Azure Queue depth:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerApp.worker
spec:
  resourceGroup: prod-rg
  name: queue-worker
  containerAppEnvironmentId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.App/managedEnvironments/prod-env
  minReplicas: 0
  maxReplicas: 20
  containers:
    - name: worker
      image: myregistry.azurecr.io/myorg/worker:v2.0.0
      cpu: 1.0
      memory: 2Gi
      env:
        - name: QUEUE_CONNECTION
          secretRef: queue-conn
  secrets:
    - name: queue-conn
      value: "Endpoint=sb://..."
  azureQueueScaleRules:
    - name: queue-scaling
      queueName: work-items
      queueLength: 10
      authentications:
        - secretRef: queue-conn
          triggerParameter: connection
```

### Enterprise API with Dapr and Identity

A production API with Dapr sidecar, managed identity, and IP restrictions:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: enterprise-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerApp.enterprise-api
spec:
  resourceGroup: prod-rg
  name: enterprise-api
  containerAppEnvironmentId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.App/managedEnvironments/prod-env
  minReplicas: 2
  maxReplicas: 15
  identity:
    type: SystemAssigned
  containers:
    - name: api
      image: myregistry.azurecr.io/myorg/enterprise-api:v3.0.0
      cpu: 1.0
      memory: 2Gi
      probes:
        - type: liveness
          httpGet:
            path: /healthz
            port: 8080
        - type: startup
          httpGet:
            path: /ready
            port: 8080
          failureThreshold: 30
          periodSeconds: 2
  dapr:
    appId: enterprise-api
    appPort: 8080
    appProtocol: http
  ingress:
    external: true
    targetPort: 8080
    transport: http2
    corsPolicy:
      allowedOrigins:
        - "https://app.example.com"
      allowedMethods:
        - GET
        - POST
        - PUT
      allowCredentials: true
    ipSecurityRestrictions:
      - name: allow-office
        action: Allow
        ipAddressRange: 203.0.113.0/24
```

### Using Foreign Key References

Reference OpenMCF-managed resources:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerApp
metadata:
  name: ref-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerApp.ref-app
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-app
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: my-env
      field: status.outputs.environment_id
  containers:
    - name: app
      image: mcr.microsoft.com/k8se/quickstart:latest
      cpu: 0.25
      memory: 0.5Gi
  ingress:
    external: true
    targetPort: 80
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `container_app_id` | `string` | Azure Resource Manager ID of the Container App |
| `latest_revision_name` | `string` | Name of the latest revision (e.g., `{app}--{suffix}`) |
| `latest_revision_fqdn` | `string` | FQDN of the latest revision, bypassing traffic splitting |
| `outbound_ip_addresses` | `string[]` | Outbound IP addresses for firewall allowlisting |
| `ingress_fqdn` | `string` | Ingress FQDN (only set when ingress is configured). In Single mode, same as latest revision FQDN. |

## Related Components

- [AzureContainerAppEnvironment](/docs/catalog/azure/azurecontainerappenvironment) -- provides the hosting environment for the Container App
- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group
- [AzureUserAssignedIdentity](/docs/catalog/azure/azureuserassignedidentity) -- provides managed identity for Key Vault and ACR access
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- stores secrets referenced by the Container App
- [AzureContainerRegistry](/docs/catalog/azure/azurecontainerregistry) -- hosts container images for the app
