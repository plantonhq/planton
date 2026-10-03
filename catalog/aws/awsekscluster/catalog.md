# AWS EKS Cluster

Deploys a managed Kubernetes control plane on Amazon EKS — the API server, etcd, and the cluster-level posture everything else composes onto: endpoint exposure, access-entry authentication, secrets encryption, control-plane logging, upgrade policy, and optionally EKS Auto Mode. The cluster is deliberately only the control plane; compute attaches as separate AwsEksNodeGroup resources (or Auto Mode manages it for you).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EKS Cluster** -- a managed Kubernetes control plane attached to your subnets (at least two Availability Zones) and cluster IAM role, with the endpoint exposure, authentication mode, encryption, logging, upgrade, and zonal-shift posture you declare
- **Secrets Encryption Configuration** -- configured only when `kmsKeyArn` is provided; enables envelope encryption of Kubernetes secrets using your customer-managed KMS key (a one-way door: it cannot be disabled on a live cluster)
- **Control Plane Log Streams** -- created for each log type in `enabledClusterLogTypes`; streams to CloudWatch Logs
- **Auto Mode Capabilities** -- when `autoMode.enabled` is set, AWS manages compute, block storage, and load balancing for the cluster's workloads without any node group
- **Provisioned Control-Plane Capacity** -- when `controlPlaneScalingTier` names a provisioned tier, the API server runs on pre-provisioned capacity instead of EKS's reactive scaling
- **Hybrid-Node Remote Networks** -- when `remoteNetworks` declares on-premises node/pod CIDR ranges, machines in those ranges can register as workers over VPN or Direct Connect
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones. Private subnets are recommended for production. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **An IAM role** trusting `eks.amazonaws.com` with the `AmazonEKSClusterPolicy` managed policy attached — attach it on the role itself; this component never modifies a role it merely references. Provide the ARN directly or reference an AwsIamRole Infra Component.
- **A KMS key** (optional) for envelope encryption of Kubernetes secrets. Provide the ARN directly or reference an AwsKmsKey Infra Component. Decide this before deploying — it cannot be added to a live cluster.

## Deploy

### Console

Open the deployment store, find **AWS EKS Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard EKS Cluster** or **Private-Endpoint EKS Cluster** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksCluster
metadata:
  name: platform-cluster
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  endpointPrivateAccess: true
  accessConfig:
    authenticationMode: API
  enabledClusterLogTypes: [audit, authenticator]
```

```shell
planton apply -f eks-cluster.yaml
```

This creates a cluster on the current default Kubernetes version with the public endpoint open (AWS default), private in-VPC access enabled, access-entry authentication, and the two highest-signal log streams. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the EKS cluster to subnets and an IAM role deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  clusterRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: eks-cluster-role
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets and IAM role first, then provisions the EKS cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EKS cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Endpoint access** -- `endpointPublicAccess` (AWS default: true) and `endpointPrivateAccess` (AWS default: false) form the exposure posture. Production clusters almost always want private access on — without it, in-VPC clients reach the API server over the internet. Restrict a public endpoint with `publicAccessCidrs`; it is the single cheapest hardening step.

**Authentication mode** -- `accessConfig.authenticationMode` chooses between EKS access entries (`API`, the modern model), the legacy aws-auth ConfigMap (`CONFIG_MAP`), or both. Migration toward `API` is one-way on a live cluster — new clusters should start there and grant access through AwsEksAccessEntry resources.

**Compute model** -- Explicit AwsEksNodeGroup fleets, or `autoMode.enabled: true` to have AWS provision compute, block storage, and load balancing itself. Alternatives, not companions — pick one per cluster.

**Secrets encryption** -- `kmsKeyArn` enables customer-managed envelope encryption of Kubernetes secrets in etcd. One-way door: once enabled it cannot be disabled or re-keyed on a live cluster, and it cannot be added later without replacement.

**Control plane logging** -- `enabledClusterLogTypes` selects from `api`, `audit`, `authenticator`, `controllerManager`, `scheduler`. Audit and authenticator carry the most signal; all five on a busy cluster carries real CloudWatch ingestion cost.

**Upgrade policy** -- `version` pins the Kubernetes minor (1.24+; EKS never downgrades); `upgradeSupportType: STANDARD` removes the extended-support surcharge risk for teams that upgrade on schedule.

**Operational guard rails** -- `deletionProtection` blocks accidental control-plane deletion; `zonalShiftEnabled` lets Amazon Application Recovery Controller drain an impaired Availability Zone; `bootstrapSelfManagedAddons: false` yields a bring-your-own-add-ons cluster for GitOps flows.

**Control-plane scale** -- `controlPlaneScalingTier` pre-provisions API-server capacity (`tier-xl` through `tier-8xl`, billed on top of the cluster itself) for very large or bursty clusters; `standard` keeps the reactive default. Updates in place.

**Hybrid nodes** -- `remoteNetworks.nodeCidrs` (and optionally `podCidrs`) declare the on-premises private ranges allowed to join; cost follows the hybrid nodes that actually register, not the declared ranges. Ranges update in place, so adding a site later is a plain update.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsIamRole** | `clusterRoleArn` | `status.outputs.role_arn` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `autoMode.nodeRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | Kubernetes API server URL | Kubernetes Provider Connections, kubectl configuration |
| `cluster_ca_certificate` | Base64-encoded cluster CA certificate | TLS verification in downstream connections |
| `cluster_security_group_id` | EKS-managed control-plane security group | Node-to-control-plane traffic rules |
| `oidc_issuer_url` | OpenID Connect issuer URL | Point an AwsIamOidcProvider here to enable IRSA with a single reference |
| `cluster_arn` | Cluster Amazon Resource Name | IAM policies, access entries, CloudWatch alarms |
| `name` | EKS cluster name | What `AwsEksNodeGroup.clusterName` references |
| `platform_version` | EKS platform revision (e.g. `eks.12`) | Fleet-wide platform-version audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard cluster** -- Public endpoint restricted by CIDRs, private access on, access-entry authentication, audit + authenticator logging. The production starting point for most teams. Start from the **Standard EKS Cluster** preset.

**Private endpoint cluster** -- `endpointPublicAccess: false` with `endpointPrivateAccess: true`: the API server is reachable only inside the VPC (VPN, Direct Connect, or bastion). Required for compliance-sensitive environments (PCI-DSS, HIPAA). Start from the **Private-Endpoint EKS Cluster** preset.

**Hands-off compute** -- `autoMode.enabled: true` with the `general-purpose` and `system` built-in pools and an Auto Mode node role: AWS operates the compute so there are no node groups to patch or scale. Start from the **EKS Auto Mode Cluster** preset.

**Hybrid fleet** -- `remoteNetworks` with your on-prem node/pod CIDRs plus regular AwsEksNodeGroup resources for the cloud side: one Kubernetes API over cloud and on-premises capacity. Start from the **EKS Hybrid Nodes Cluster** preset.

## Works With

- [**AWS EKS Node Group**](/infra-catalog/aws-eks-node-group) -- managed worker fleets that register with this cluster's `name` output
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for control plane network interface placement
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the cluster service role (and the Auto Mode node role)
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for Kubernetes secrets encryption
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides additional control-plane security groups
