# Terraform Module for NATS on Kubernetes

This Terraform module provides infrastructure-as-code for deploying NATS clusters on Kubernetes. It enables simplified deployment and management of NATS messaging infrastructure with minimal configuration.

## Key Features

### 1. **Namespace Management**
Automatically creates and manages a dedicated Kubernetes namespace for the NATS cluster, ensuring isolation from other workloads.

### 2. **Resource Labeling**
Applies consistent labels to all resources for tracking, organization, and integration with monitoring systems.

### 3. **Ingress Support**
Supports external access via LoadBalancer services with DNS hostname configuration through external-dns annotations.

### 4. **NACK Controller & JetStream Streams**
Optionally deploys the NACK (NATS Controllers for Kubernetes) controller for declarative JetStream resource management:
- **Streams**: Define JetStream streams with subjects, storage types, retention policies, and limits
- **Consumers**: Create durable consumers with delivery policies, acknowledgment settings, and filtering
- **Control-Loop Mode**: Enable for KeyValue, ObjectStore, and Account support

### 5. **Flexible Configuration**
Allows customization of:
- Server replicas
- CPU and memory resources
- Disk size for JetStream persistence
- Authentication schemes (Bearer Token, Basic Auth)
- TLS encryption
- External ingress
- NACK controller settings
- JetStream streams and consumers

## Usage

### Prerequisites

* **Terraform**: Version 0.13 or higher
* **Kubernetes Cluster**: Access to a Kubernetes cluster with valid credentials
* **Planton CLI**: For executing deployment commands

### Example Commands

Initialize Terraform:
```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/nats-basic.tfstate"
```

Plan the deployment:
```shell
openmcf tofu plan --manifest hack/manifest.yaml
```

Apply the configuration:
```shell
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

## Outputs

After successful deployment, the module provides:

* **namespace**: Kubernetes namespace where NATS is deployed
* **internal_client_url**: Internal DNS URL for NATS clients (e.g., `nats://nats-basic.nats-basic.svc.cluster.local:4222`)
* **external_hostname**: External hostname if ingress is enabled
* **auth_secret_name**: Name of the secret containing authentication credentials
* **tls_secret_name**: Name of the TLS secret (if TLS is enabled)
* **nack_controller_enabled**: Whether NACK controller is deployed
* **nack_controller_version**: Version of the NACK controller
* **streams_created**: List of JetStream streams created
* **jetstream_domain**: JetStream domain identifier

## Deployment Architecture

When NACK controller is enabled, the module deploys resources in this order to avoid race conditions:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Deployment Order                          │
│                                                                  │
│  1. NATS Helm Chart ──► 2. NACK CRDs ──► 3. NACK Controller     │
│                                                  │               │
│                                                  ▼               │
│                                    4. Stream/Consumer CRs        │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

The module accepts configuration through a YAML manifest matching the `KubernetesNats` protobuf specification. See `hack/manifest.yaml` for a complete example.

### Example with Streams

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: nats-with-streams
spec:
  namespace:
    value: nats-streams
  serverContainer:
    replicas: 3
    diskSize: 10Gi
  nack_controller:
    enabled: true
    enable_control_loop: true
  streams:
    - name: orders
      subjects:
        - "orders.>"
      storage: file
      replicas: 3
      retention: limits
      max_age: 24h
      consumers:
        - durable_name: orders-processor
          ack_policy: explicit
          max_ack_pending: 1000
```

## Benefits

* **Infrastructure-as-Code**: Version-controlled, repeatable deployments
* **Simplified Management**: Automated namespace and resource creation
* **Declarative Streams**: Define JetStream streams and consumers in YAML
* **Flexibility**: Easy customization for various deployment scenarios
* **Consistency**: Standardized labeling and naming conventions
* **Race Condition Handling**: Proper dependency ordering for CRDs and custom resources
