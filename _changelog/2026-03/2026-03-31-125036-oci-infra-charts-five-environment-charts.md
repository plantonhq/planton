# OCI InfraCharts: Five Environment Charts for Oracle Cloud Infrastructure

**Date**: March 31, 2026
**Type**: New Chart
**Provider**: OCI (Oracle Cloud Infrastructure)
**Chart(s)**: oke-environment, autonomous-db-stack, compute-environment, serverless-stack, data-platform

## Summary

Added five production-ready InfraCharts for Oracle Cloud Infrastructure, covering the most common enterprise deployment patterns: Kubernetes (OKE), managed databases (Autonomous Database), VM-based compute, serverless functions with API Gateway, and analytics/data lake. These charts bring OCI to parity with the existing AWS, GCP, and Azure chart families.

## Problem Statement / Motivation

OCI had 37 fully implemented Planton deployment components (merged via PR #422 in the planton repo) but zero InfraCharts. Without charts, users had to manually compose individual resources -- wiring VCN outputs to subnets, subnets to NSGs, NSGs to clusters, etc. This is exactly the friction InfraCharts are designed to eliminate.

### Pain Points

- No one-click OCI environment provisioning despite having all building blocks
- Manual valueFrom wiring between 5-8 resources per environment is error-prone
- OCI's compartment model adds a `compartmentId` requirement to every single resource
- No cross-provider parity -- AWS and GCP had environment charts but OCI didn't

## Solution / What's New

Five charts under `oci/` in the infra-charts repo, following the exact same conventions as `aws/eks-environment`, `gcp/gke-environment`, and `azure/web-app-environment`.

### Chart Structure

```
oci/
├── oke-environment/          # Kubernetes (OKE) cluster
├── autonomous-db-stack/      # Managed Oracle Database
├── compute-environment/      # VM-based workloads
├── serverless-stack/         # Functions + API Gateway
└── data-platform/            # Analytics + Data Lake
```

## Implementation Details

### Resources Included

| Chart | Resources | Templates | Params | Conditional |
|-------|-----------|-----------|--------|-------------|
| oke-environment | VCN, 2 Subnets, 2 NSGs, Cluster, NodePool, DnsZone | 4 | 20 | DNS |
| autonomous-db-stack | VCN, Subnet, NSG, ADB, KmsVault, KmsKey | 3 | 11 | KMS encryption |
| compute-environment | VCN, 2 Subnets, NSG, Instance, ALB, BlockVol | 4 | 17 | ALB, block volume |
| serverless-stack | VCN, 2 Subnets, Functions, Gateway, Bucket, LogGroup | 4 | 9 | None |
| data-platform | VCN, Subnet, ADW, Bucket, StreamPool, LogGroup | 5 | 12 | None |

### OCI-Specific Patterns

**Compartment propagation**: Every resource receives `compartmentId.value` from the `compartment_ocid` param. Unlike AWS (ambient account) or GCP (optional project creation), OCI requires explicit compartment on every resource.

**Subnet route rules**: Public subnets route through the VCN's internet gateway (`status.outputs.internetGatewayId`). Private subnets route through the NAT gateway (`status.outputs.natGatewayId`). Route rules are inline on each subnet using `valueFrom` to reference VCN gateway outputs.

**NSG security rules**: Each chart creates purpose-specific NSGs (e.g., API endpoint NSG allows TCP 6443 + 12250 from VCN CIDR; worker NSG allows all from VCN; database NSG allows SQL*Net 1522 + HTTPS 443).

### Conditional Resources

```yaml
# KMS encryption (autonomous-db-stack)
{% if values.enable_encryption | bool %}
---
apiVersion: oci.planton.dev/v1
kind: OciKmsVault
...
{% endif %}

# DNS zone (oke-environment)
{% if values.enable_dns | bool %}
---
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
...
{% endif %}

# Load balancer (compute-environment)
{% if values.enable_load_balancer | bool %}
---
apiVersion: oci.planton.dev/v1
kind: OciApplicationLoadBalancer
...
{% endif %}
```

### Resource Relationships (valueFrom Wiring)

All charts wire resources through `valueFrom` references:

```
OciVcn (outputs: vcnId, internetGatewayId, natGatewayId)
  └─ OciSubnet (inputs: vcnId, routeRules[].networkEntityId)
       └─ OciSecurityGroup (inputs: vcnId)
            └─ OciContainerEngineCluster (inputs: endpointConfig.subnetId, nsgIds)
                 └─ OciContainerEngineNodePool (inputs: clusterId, placementConfigs[].subnetId)
```

## Benefits

- **5-minute OCI environments**: From zero to production-ready OKE cluster, ADB stack, or serverless platform
- **Cross-provider parity**: OCI joins AWS, GCP, and Azure as a first-class InfraChart provider
- **OCI best practices baked in**: Private subnets for workloads, NAT for outbound, NSGs per resource role, service gateways for OCI service access
- **Composable toggles**: Optional DNS, KMS, load balancer, and block volume via boolean flags
- **35 files, 1418 lines**: Complete chart family in a single commit

## Impact

Platform users can now provision OCI environments through InfraCharts with the same experience as AWS EKS or GCP GKE environments. The five charts cover the most requested enterprise OCI patterns.

## Usage Example

```bash
# Preview the OKE environment chart
planton chart build oci/oke-environment

# Create a project from the chart
planton project create --from-chart oci/oke-environment \
  --name my-oke-cluster \
  --values ./my-values.yaml
```

Example `values.yaml` override:

```yaml
params:
  - name: compartment_ocid
    value: ocid1.compartment.oc1..aaaaaaaexample
  - name: cluster_name
    value: production-oke
  - name: kubernetes_version
    value: v1.30.1
  - name: node_shape
    value: VM.Standard.E4.Flex
  - name: node_pool_size
    value: "5"
```

## Related Work

- OCI Planton components: 37 resource kinds merged via planton PR #422
- Quality audit: 36/37 EXCELLENT, 1/37 GOOD (2026-03-31)
- Existing chart families: `aws/eks-environment`, `gcp/gke-environment`, `azure/web-app-environment`

---

**Status**: Production Ready
**Timeline**: Created 2026-03-31
