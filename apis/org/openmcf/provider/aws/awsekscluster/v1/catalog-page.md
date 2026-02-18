# AWS EKS Cluster

Deploys an AWS EKS cluster control plane with configurable public/private API endpoint access, optional envelope encryption of Kubernetes secrets via KMS, and optional control plane logging to CloudWatch. The component requires at least two subnets in distinct Availability Zones for high availability.

## What Gets Created

When you deploy an AwsEksCluster resource, OpenMCF provisions:

- **EKS Cluster** — an `aws:eks:Cluster` control plane placed in the specified subnets, using the provided IAM role for AWS API interactions, with configurable public and private endpoint access
- **Control Plane Log Streams** — created only when `enableControlPlaneLogs` is `true`; enables all five log types (API server, audit, authenticator, controller manager, scheduler) to CloudWatch Logs
- **Secrets Encryption Configuration** — configured only when `kmsKeyArn` is provided; enables envelope encryption of Kubernetes secrets using the specified customer-managed KMS key

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least two subnets** in different Availability Zones within the target VPC (private subnets recommended)
- **An IAM role** with the `AmazonEKSClusterPolicy` attached, for the EKS service to manage cluster resources
- **A KMS key ARN** if enabling envelope encryption of Kubernetes secrets

## Quick Start

Create a file `eks-cluster.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEksCluster.my-cluster
spec:
  region: us-west-2
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  clusterRoleArn: arn:aws:iam::123456789012:role/EksClusterServiceRole
```

Deploy:

```shell
openmcf apply -f eks-cluster.yaml
```

This creates an EKS cluster with a public API endpoint across two subnets, using the default Kubernetes version.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the EKS cluster will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `subnetIds` | `string[]` | Subnet IDs in the cluster's VPC where the EKS control plane attaches network interfaces. Use at least two subnets in distinct Availability Zones. | Minimum 2 items required |
| `subnetIds[].value` | `string` | Direct subnet ID value | — |
| `subnetIds[].valueFrom` | `object` | Foreign key reference to an AwsVpc resource | Default kind: `AwsVpc`, field: `status.outputs.private_subnets.[*].id` |
| `clusterRoleArn` | `string` | ARN of an IAM role for the EKS cluster to use when interacting with AWS services. Must have `AmazonEKSClusterPolicy` attached. | Required |
| `clusterRoleArn.value` | `string` | Direct IAM role ARN value | — |
| `clusterRoleArn.valueFrom` | `object` | Foreign key reference to an AwsIamRole resource | Default kind: `AwsIamRole`, field: `status.outputs.role_arn` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | `string` | Latest supported | Kubernetes version for the cluster control plane (e.g., `"1.29"`). If not set, AWS uses the latest supported version. |
| `disablePublicEndpoint` | `bool` | `false` | When `true`, the cluster API endpoint is accessible only within the VPC. When `false`, the endpoint is publicly accessible. |
| `publicAccessCidrs` | `string[]` | `["0.0.0.0/0"]` | IPv4 CIDR blocks allowed to access the cluster's public API endpoint. Each entry must be a valid IPv4 CIDR (e.g., `"203.0.113.0/24"`). Ignored when `disablePublicEndpoint` is `true`. |
| `enableControlPlaneLogs` | `bool` | `false` | Enables all control plane log types (API, audit, authenticator, controller manager, scheduler) to CloudWatch Logs. |
| `kmsKeyArn` | `string` | — | KMS key ARN for envelope encryption of Kubernetes secrets. If not set, the cluster uses the default AWS-managed EKS key. Can reference an AwsKmsKey resource via `valueFrom`. |

## Examples

### Private Cluster with Restricted Access

An EKS cluster with the public endpoint disabled, accessible only from within the VPC:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: private-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEksCluster.private-cluster
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  clusterRoleArn: arn:aws:iam::123456789012:role/EksClusterServiceRole
  version: "1.29"
  disablePublicEndpoint: true
```

### Cluster with Logging and CIDR Restrictions

A cluster with control plane logging enabled and public access restricted to specific CIDR blocks:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: monitored-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsEksCluster.monitored-cluster
spec:
  region: us-west-2
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  clusterRoleArn: arn:aws:iam::123456789012:role/EksClusterServiceRole
  version: "1.29"
  publicAccessCidrs:
    - 203.0.113.0/24
    - 198.51.100.0/24
  enableControlPlaneLogs: true
```

### Full-Featured Production Cluster

Production configuration with KMS encryption, control plane logging, and private endpoint:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEksCluster.prod-cluster
spec:
  region: us-east-1
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  clusterRoleArn: arn:aws:iam::123456789012:role/EksClusterServiceRole
  version: "1.29"
  disablePublicEndpoint: true
  enableControlPlaneLogs: true
  kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/abcd1234-5678-90ab-cdef-example11111
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs and IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: ref-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEksCluster.ref-cluster
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets.[1].id
  clusterRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: eks-cluster-role
      field: status.outputs.role_arn
  version: "1.29"
  enableControlPlaneLogs: true
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: eks-secrets-key
      field: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `endpoint` | `string` | URL of the Kubernetes API server for the EKS cluster |
| `cluster_ca_certificate` | `string` | Base64-encoded certificate authority data for the cluster |
| `cluster_security_group_id` | `string` | ID of the security group created by EKS for the cluster control plane |
| `oidc_issuer_url` | `string` | URL of the OpenID Connect issuer for the cluster, used for IAM Roles for Service Accounts (IRSA) |
| `cluster_arn` | `string` | Amazon Resource Name of the EKS cluster |
| `name` | `string` | Name assigned to the EKS cluster |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the subnets for cluster control plane placement
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides the IAM role for the EKS cluster service
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides the customer-managed key for secrets encryption
