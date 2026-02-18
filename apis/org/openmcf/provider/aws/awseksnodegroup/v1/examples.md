# AWS EKS Node Group Examples

This document provides practical examples for deploying AWS EKS managed node groups using OpenMCF. 
After creating one of these YAML manifests, deploy it using the OpenMCF CLI:

```shell
# Using Pulumi
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>

# Using Terraform/OpenTofu
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic EKS Node Group

A minimal configuration for a production-ready node group with on-demand instances.

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

**Key Points:**
- Uses `t3.medium` instances for cost-effective performance
- Auto-scaling between 1-3 nodes with 2 desired nodes
- On-demand instances for predictable costs
- 100GB root disk (recommended default)
- Deployed across two private subnets for high availability

---

## Spot Instance Node Group

Optimized for cost savings using AWS Spot instances (up to 90% cost reduction).

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

**Key Points:**
- Spot instances provide up to 90% cost savings
- Larger scaling range for flexible workload handling
- Kubernetes labels for workload targeting (use taints/tolerations)
- Suitable for fault-tolerant, stateless workloads
- Not recommended for critical production services

---

## Node Group with SSH Access

Development-focused configuration with SSH access enabled for debugging.

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

**Key Points:**
- SSH key configured for direct node access
- Larger disk size for development workloads
- Debug-specific labels for node targeting
- On-demand instances for stable access
- Suitable for development and troubleshooting
- **Security Note:** Avoid SSH access in production; use AWS Systems Manager Session Manager instead

---

## Production-Grade Node Group

Enterprise-ready configuration with multiple availability zones and high capacity.

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

**Key Points:**
- Large instance type (m5.xlarge: 4 vCPU, 16 GiB RAM) for performance
- Three subnets across availability zones for high availability
- High scaling range (3-20) for traffic spikes
- Large disk size for data-intensive workloads
- Production-specific labels for workload targeting
- On-demand instances for reliability

---

## Development/Test Node Group

Cost-optimized configuration for non-production environments.

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

**Key Points:**
- Small instance type (t3.small: 2 vCPU, 2 GiB RAM) for cost efficiency
- SSH access for debugging and development
- Smaller disk size for development workloads
- Minimal scaling for cost control
- Development-specific labels
- Can be scaled down to zero nodes during off-hours

---

## Mixed Workload Node Group

Balanced configuration for flexible, non-critical workloads.

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

**Key Points:**
- Spot instances for cost optimization
- Moderate instance type (t3.large: 2 vCPU, 8 GiB RAM)
- Flexible scaling for varying workloads
- Cost optimization labels
- Suitable for non-critical, horizontally scalable workloads

---

## Resource References Example

Demonstrates OpenMCF's foreign key system for automatic resource linking.

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

**Key Points:**
- Cluster name automatically referenced from `AwsEksCluster` resource
- Node role automatically referenced from `AwsIamRole` resource
- Subnets automatically referenced from `AwsVpc` resource
- Enables dependency management and resource linking
- Reduces manual ARN/ID management
- Ensures correct deployment order

---

## Multi-Environment Pattern

Production and development node groups in a single manifest.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: prod-node-group
spec:
  region: us-west-2
  clusterName:
    value: "my-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksNodeRole"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "m5.xlarge"
  scaling:
    minSize: 3
    maxSize: 10
    desiredSize: 5
  capacityType: "on_demand"
  diskSizeGb: 200
  labels:
    environment: "production"
---
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: dev-node-group
spec:
  region: us-west-2
  clusterName:
    value: "my-eks-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksNodeRole"
  subnetIds:
    - value: "subnet-private-1a"
    - value: "subnet-private-1b"
  instanceType: "t3.small"
  scaling:
    minSize: 1
    maxSize: 3
    desiredSize: 2
  capacityType: "spot"
  diskSizeGb: 50
  labels:
    environment: "development"
```

**Key Points:**
- Multiple node groups in a single cluster
- Different instance types and scaling for different workloads
- Use Kubernetes taints/tolerations to control pod placement
- Production uses on-demand, development uses Spot for cost savings

---

## Verification Commands

After deploying your node group, verify it using these commands:

### AWS CLI Verification

```shell
# List all node groups in a cluster
aws eks list-nodegroups --cluster-name <your-cluster-name>

# Describe a specific node group
aws eks describe-nodegroup \
  --cluster-name <your-cluster-name> \
  --nodegroup-name <your-nodegroup-name>

# Check Auto Scaling Group details
aws autoscaling describe-auto-scaling-groups \
  --query "AutoScalingGroups[?contains(Tags[?Key=='eks:nodegroup-name'].Value, '<your-nodegroup-name>')]"
```

### Kubernetes Verification

```shell
# List all nodes
kubectl get nodes

# List nodes in a specific node group
kubectl get nodes --label-selector=eks.amazonaws.com/nodegroup=<your-nodegroup-name>

# View node details and labels
kubectl describe node <node-name>

# Check node group labels
kubectl get nodes --show-labels | grep <your-nodegroup-name>
```

### Cluster Autoscaler Integration

```shell
# Verify cluster autoscaler can discover the node group
kubectl logs -n kube-system deployment/cluster-autoscaler | grep <your-nodegroup-name>

# Check autoscaler status
kubectl get cm cluster-autoscaler-status -n kube-system -o yaml
```

---

## Best Practices

1. **Multi-AZ Deployment**: Always use at least 2 subnets across different availability zones
2. **Scaling Strategy**: Set `minSize >= 1` for production node groups to ensure availability
3. **Spot Instance Pattern**: Use Spot for batch workloads, on-demand for critical services
4. **Instance Sizing**: Start with t3.medium for dev, m5.xlarge for production
5. **Disk Size**: Use 100GB+ for production (default 20GB is often insufficient)
6. **Security**: Avoid SSH access in production; use AWS Systems Manager Session Manager
7. **Labels**: Use consistent labeling strategy for workload targeting and cost allocation
8. **IAM Roles**: Use minimal node IAM roles; prefer IRSA for pod-level permissions
9. **Updates**: Keep node AMIs updated; use managed node group update features
10. **Monitoring**: Enable CloudWatch Container Insights for comprehensive monitoring

---

## Common Issues and Solutions

### Nodes Not Joining Cluster

**Symptoms**: Nodes are created but don't appear in `kubectl get nodes`

**Solutions**:
- Verify node IAM role has required policies (EKSWorkerNodePolicy, EKS_CNI_Policy, EC2ContainerRegistryReadOnly)
- Check subnet tags: `kubernetes.io/cluster/<cluster-name>: shared`
- Verify security group allows communication with cluster control plane
- Check VPC DNS settings (enableDnsHostnames, enableDnsSupport)

### Insufficient Capacity

**Symptoms**: Node group fails to scale up

**Solutions**:
- Try different instance types or availability zones
- Request EC2 capacity increase for your instance type
- Use multiple instance types (consider using managed instance type diversification)

### Spot Instance Interruptions

**Symptoms**: Spot nodes frequently terminated

**Solutions**:
- Deploy AWS Node Termination Handler (https://github.com/aws/aws-node-termination-handler)
- Use Pod Disruption Budgets to ensure graceful pod eviction
- Consider mixing on-demand and Spot node groups
- Use multiple instance types in Spot node groups

---

## Additional Resources

- [AWS EKS Managed Node Groups Documentation](https://docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html)
- [EKS Best Practices Guide](https://aws.github.io/aws-eks-best-practices/)
- [Cluster Autoscaler Documentation](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler)
- [OpenMCF Documentation](https://openmcf.org/docs)

