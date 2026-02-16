# GcpVertexAiEndpoint Examples

## Minimal Public Endpoint

The simplest configuration -- a public endpoint accessible via the shared regional DNS.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: my-ml-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Recommendation Endpoint
```

## Public Endpoint with Description

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: prod-recommendations
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Production Recommendation Engine
  description: Serves real-time product recommendations for the e-commerce platform
```

## Dedicated DNS Endpoint

Better performance and traffic isolation via a dedicated DNS name.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: high-perf-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: High Performance Endpoint
  dedicatedEndpointEnabled: true
```

## VPC-Peered Private Endpoint

Accessible only within a peered VPC network. Requires Private Services Access.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: private-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Private ML Endpoint
  network:
    value: projects/123456789/global/networks/my-vpc
```

## VPC-Peered with CMEK Encryption

Private endpoint with customer-managed encryption for sensitive workloads.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: secure-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Secure ML Endpoint
  description: HIPAA-compliant endpoint with CMEK and VPC isolation
  network:
    value: projects/123456789/global/networks/prod-vpc
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/ml-ring/cryptoKeys/endpoint-key
  dedicatedEndpointEnabled: true
```

## Private Service Connect Endpoint

The strongest network isolation without VPC peering.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: psc-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: PSC ML Endpoint
  privateServiceConnectConfig:
    projectAllowlist:
      - consumer-project-a
      - consumer-project-b
```

## Infra-Chart Composition (valueFrom)

Referencing other OpenMCF resources for composable infrastructure.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: composed-endpoint
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: ml-project
      fieldPath: status.outputs.project_id
  location: us-central1
  displayName: Composed ML Endpoint
  network:
    valueFrom:
      kind: GcpVpc
      name: ml-vpc
      fieldPath: status.outputs.network_self_link
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: ml-encryption-key
      fieldPath: status.outputs.key_id
```
