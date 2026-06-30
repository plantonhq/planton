# Overview

The **AwsEksNodeGroup** API resource provides a standardized and straightforward way to deploy managed worker node groups
for Amazon EKS clusters on AWS. By focusing on essential configurations like instance type, scaling parameters, 
capacity type (on-demand vs. Spot), and networking, it makes running production-ready EKS worker nodes far more accessible
within the Planton multi-cloud deployment framework.

## Purpose

Deploying EKS node groups typically involves handling multiple moving parts—IAM roles, networking, autoscaling,
instance configuration, and lifecycle management. The **AwsEksNodeGroup** resource aims to streamline that process by:

- **Simplifying EKS Node Deployments**: Offer an easy-to-use interface for spinning up managed worker nodes for your EKS clusters.
- **Aligning with Best Practices**: Provide recommended defaults (e.g., disk size, capacity type) to ensure users have a
  production-ready baseline without repetitive configuration.
- **Promoting Consistency**: Enforce standardized naming and validations, reducing misconfigurations across
  multiple clusters and environments.
- **Cost Optimization**: Support both on-demand and Spot instances for flexible cost management.

## Key Features

### Managed Node Group Focus

- **Minimal, Opinionated Spec**: Focuses on the 80-20 use case—managed worker nodes with essential configuration—while 
  exposing fields for scaling, instance types, and node labels.
- **AWS-Managed Lifecycle**: Leverage AWS EKS managed node groups for automated node provisioning, updates, and termination.

### Flexible Compute Options

- **On-Demand or Spot**: Choose between reliable on-demand instances or cost-effective Spot instances.
- **Instance Type Control**: Specify any EC2 instance type (t3.medium, m5.xlarge, etc.) to match your workload requirements.
- **Auto-Scaling**: Define minimum, maximum, and desired node counts for automatic scaling based on workload.

### Automatic Networking Setup

- **Multi-AZ Deployment**: Distribute nodes across multiple subnets for high availability.
- **VPC Integration**: Seamlessly integrate with your existing VPC and subnet configuration.
- **Private Subnet Support**: Deploy nodes in private subnets for enhanced security.

### Node Customization

- **Kubernetes Labels**: Apply custom labels to nodes for workload targeting and pod scheduling.
- **SSH Access**: Optionally enable SSH access to nodes for debugging and troubleshooting.
- **Custom Disk Size**: Configure EBS root volume size to meet your storage requirements.

### Seamless Integration

- **Planton CLI**: Deploy the same resource across multiple stacks using either Pulumi or Terraform under the hood.
- **Multi-Cloud Ready**: Combine AwsEksNodeGroup on AWS with other providers in the same manifest, adopting Planton's
  uniform resource model.
- **Resource References**: Reference other Planton resources (clusters, IAM roles, VPCs) using foreign key relationships.

## Benefits

- **Reduced Complexity**: A single definition for your EKS node group—instance type, scaling, subnets, and more—means
  fewer files and less overhead.
- **Scalable & Available**: Scale horizontally by adjusting node counts to meet traffic demands without repeatedly editing
  multiple configuration files.
- **Infrastructure Consistency**: Enforce naming conventions, validations, and recommended defaults for node configurations
  so your deployments remain predictable and repeatable.
- **Cost Optimization**: Easily switch between on-demand and Spot instances to balance cost and reliability.
- **Enhanced Observability**: Integrate seamlessly with EKS cluster features like CloudWatch metrics and logs—no extra
  manual setup needed.

## Example Usage

Below is a minimal YAML snippet demonstrating how to configure and deploy an EKS node group using Planton:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEksNodeGroup
metadata:
  name: production-workers
spec:
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
  diskSizeGb: 100
  labels:
    environment: "production"
    workload: "general"
```

### Deploying with Planton

Once your YAML manifest is ready, you can deploy using Planton's CLI. Planton will validate the manifest
against the Protobuf schema and orchestrate everything in Pulumi or Terraform.

- **Using Pulumi**:
  ```bash
  planton pulumi up --manifest awseksnodegroup.yaml --stack org/project/my-stack
  ```
- **Using Terraform**:
  ```bash
  planton tofu apply --manifest awseksnodegroup.yaml --stack org/project/my-stack
  ```

Planton will provision the EKS node group, configure the Auto Scaling Group, attach the IAM role, deploy nodes
across the specified subnets, and ensure your desired number of nodes are running and ready.

---

Happy deploying! If you have questions or run into issues, feel free to open an issue on our GitHub repository or
reach out through our community channels for support.
