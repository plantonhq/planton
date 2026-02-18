# AWS EKS Node Group Examples

Below are several examples demonstrating how to define an AWS EKS Node Group component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic EKS Node Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: basic-node-group
spec:
  region: us-west-2
  clusterName:
    value: "my-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksNodeRole"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "t3.medium"
  scaling:
    minSize: 1
    maxSize: 3
    desiredSize: 2
  capacityType: "on_demand"
  diskSizeGb: 100
```

This example creates a basic EKS node group:
• Uses t3.medium instances for cost-effective performance.
• Auto-scaling between 1-3 nodes with 2 desired nodes.
• On-demand instances for predictable costs.
• 100GB root disk for application data.
• Deployed across two private subnets for high availability.

---

## EKS Node Group with Spot Instances

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: spot-node-group
spec:
  region: us-west-2
  clusterName:
    value: "my-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksNodeRole"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "t3.large"
  scaling:
    minSize: 2
    maxSize: 10
    desiredSize: 3
  capacityType: "spot"
  diskSizeGb: 100
  labels:
    node-type: "spot"
    workload: "batch"
```

This example uses Spot instances for cost optimization:
• Spot instances provide up to 90% cost savings.
• Larger instance type for better performance.
• Higher scaling range for workload flexibility.
• Kubernetes labels for workload targeting.
• Suitable for fault-tolerant workloads.

---

## EKS Node Group with SSH Access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: ssh-node-group
spec:
  region: us-west-2
  clusterName:
    value: "my-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksNodeRole"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "m5.large"
  scaling:
    minSize: 1
    maxSize: 5
    desiredSize: 2
  capacityType: "on_demand"
  diskSizeGb: 200
  sshKeyName: "my-ssh-key"
  labels:
    node-type: "debug"
    environment: "development"
```

This example enables SSH access for debugging:
• SSH key configured for direct node access.
• Larger disk size for development workloads.
• Debug-specific labels for node targeting.
• On-demand instances for stable access.
• Suitable for development and troubleshooting.

---

## Production EKS Node Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: production-node-group
spec:
  region: us-west-2
  clusterName:
    value: "production-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/production-eks-node-role"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
    - value: "subnet-private-1c"
  instanceType: "m5.xlarge"
  scaling:
    minSize: 3
    maxSize: 20
    desiredSize: 5
  capacityType: "on_demand"
  diskSizeGb: 500
  labels:
    node-type: "production"
    workload: "critical"
    environment: "production"
```

This example is production-ready:
• Large instance type for performance.
• Three subnets across availability zones.
• High scaling range for traffic spikes.
• Large disk size for data-intensive workloads.
• Production-specific labels for workload targeting.
• On-demand instances for reliability.

---

## Development EKS Node Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: development-node-group
spec:
  region: us-west-2
  clusterName:
    value: "dev-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/dev-eks-node-role"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "t3.small"
  scaling:
    minSize: 1
    maxSize: 5
    desiredSize: 2
  capacityType: "on_demand"
  diskSizeGb: 50
  sshKeyName: "dev-ssh-key"
  labels:
    node-type: "development"
    environment: "dev"
    team: "engineering"
```

This example is optimized for development:
• Small instance type for cost efficiency.
• SSH access for debugging and development.
• Smaller disk size for development workloads.
• Development-specific labels.
• Minimal scaling for cost control.
• Suitable for development and testing.

---

## Mixed Capacity EKS Node Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: mixed-capacity-node-group
spec:
  region: us-west-2
  clusterName:
    value: "mixed-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/mixed-eks-node-role"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "t3.large"
  scaling:
    minSize: 2
    maxSize: 15
    desiredSize: 4
  capacityType: "spot"
  diskSizeGb: 150
  labels:
    node-type: "mixed"
    workload: "flexible"
    cost-optimized: "true"
```

This example balances cost and performance:
• Spot instances for cost optimization.
• Moderate instance type for performance.
• Flexible scaling for varying workloads.
• Cost optimization labels.
• Suitable for non-critical workloads.

---

## EKS Node Group with Resource References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: ref-node-group
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: "my-eks-cluster"
      fieldPath: "metadata.name"
  nodeRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: "eks-node-role"
      fieldPath: "status.outputs.roleArn"
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[0].id"
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[1].id"
  instanceType: "t3.medium"
  scaling:
    minSize: 1
    maxSize: 5
    desiredSize: 2
  capacityType: "on_demand"
  diskSizeGb: 100
```

This example uses OpenMCF resource references:
• Cluster name automatically referenced from EKS cluster.
• Node role automatically referenced from IAM role.
• Subnets automatically referenced from VPC.
• Enables dependency management and resource linking.
• Reduces manual ARN management.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the EKS node group is active via the AWS console or by
using the AWS CLI:

```shell
aws eks describe-nodegroup --cluster-name <your-cluster-name> --nodegroup-name <your-nodegroup-name>
```

For detailed node group information:

```shell
aws eks list-nodegroups --cluster-name <your-cluster-name>
```

To verify the nodes are ready in Kubernetes:

```shell
kubectl get nodes --label-selector=eks.amazonaws.com/nodegroup=<your-nodegroup-name>
```

This will show the worker nodes associated with your node group, including their status and labels.

