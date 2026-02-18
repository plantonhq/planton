# AWS EKS Cluster Examples

Below are several examples demonstrating how to define an AWS EKS Cluster component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic EKS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: basic-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-1234567890abcdef0"
    - value: "subnet-0987654321fedcba0"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  version: "1.28"
```

This example creates a basic EKS cluster:
• Uses Kubernetes version 1.28.
• Deployed across two subnets for high availability.
• Public API endpoint enabled by default.
• Uses default AWS-managed KMS key for secrets encryption.
• Control plane logging disabled by default.

---

## EKS Cluster with Private Endpoint

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: private-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  version: "1.27"
  disablePublicEndpoint: true
```

This example creates a private EKS cluster:
• API endpoint accessible only within the VPC.
• Enhanced security for compliance requirements.
• Requires VPN or bastion host for external access.
• Suitable for highly secure environments.

---

## EKS Cluster with Restricted Public Access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: restricted-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  version: "1.28"
  disablePublicEndpoint: false
  publicAccessCidrs:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"
```

This example restricts public access:
• Public API endpoint enabled but restricted to private IP ranges.
• Allows access from corporate networks and VPNs.
• Balances security with accessibility.
• Common configuration for enterprise environments.

---

## EKS Cluster with Control Plane Logging

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: logging-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  version: "1.28"
  enableControlPlaneLogs: true
```

This example enables comprehensive logging:
• All control plane log types enabled (API, audit, authenticator, controller manager, scheduler).
• Logs sent to CloudWatch for monitoring and compliance.
• Useful for security auditing and troubleshooting.
• May incur additional CloudWatch costs.

---

## EKS Cluster with KMS Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: encrypted-eks-cluster
spec:
  region: us-east-1
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/EksClusterServiceRole"
  version: "1.28"
  kmsKeyArn:
    value: "arn:aws:kms:us-east-1:123456789012:key/abcd1234-5678-90ef-ghij-klmnopqrstuv"
```

This example uses customer-managed KMS encryption:
• Kubernetes secrets encrypted with customer-managed KMS key.
• Enhanced security and compliance.
• Full control over encryption key lifecycle.
• Required for certain compliance standards.

---

## Production EKS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: production-eks-cluster
spec:
  region: us-east-1
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
    - value: "subnet-private-1c"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/production-eks-cluster-role"
  version: "1.28"
  disablePublicEndpoint: false
  publicAccessCidrs:
    - "10.0.0.0/8"
  enableControlPlaneLogs: true
  kmsKeyArn:
    value: "arn:aws:kms:us-east-1:123456789012:key/production-eks-key"
```

This example is production-ready:
• Three subnets across availability zones for high availability.
• Restricted public access to corporate network.
• Comprehensive control plane logging.
• Customer-managed KMS encryption.
• Latest stable Kubernetes version.

---

## Development EKS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: development-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  clusterRoleArn:
    value: "arn:aws:iam::123456789012:role/dev-eks-cluster-role"
  version: "1.28"
  disablePublicEndpoint: false
```

This example is optimized for development:
• Public API endpoint for easy access.
• Minimal configuration for rapid deployment.
• Uses default AWS-managed KMS key.
• Control plane logging disabled to reduce costs.
• Suitable for development and testing environments.

---

## EKS Cluster with VPC References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksCluster
metadata:
  name: vpc-ref-eks-cluster
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[0].id"
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[1].id"
  clusterRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: "eks-cluster-role"
      fieldPath: "status.outputs.roleArn"
  version: "1.28"
```

This example uses OpenMCF resource references:
• Subnets automatically referenced from VPC resource.
• Cluster role automatically referenced from IAM role resource.
• Enables dependency management and resource linking.
• Reduces manual ARN management.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the EKS cluster is active via the AWS console or by
using the AWS CLI:

```shell
aws eks list-clusters
```

For detailed cluster information:

```shell
aws eks describe-cluster --name <your-cluster-name>
```

This will show cluster details including endpoint, version, status, and configuration. To configure kubectl:

```shell
aws eks update-kubeconfig --name <your-cluster-name> --region <region>
```

Then verify connectivity:

```shell
kubectl get nodes
```
