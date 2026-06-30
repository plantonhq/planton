# Deploying AWS EKS Node Groups: From Manual Chaos to Production Automation

For years, conventional wisdom held that running Kubernetes meant accepting operational complexity. The reality today is different: AWS EKS managed node groups have evolved from a basic feature into a sophisticated platform that handles node lifecycle, upgrades, and scaling—but only if you deploy them correctly.

The challenge isn't whether to use managed node groups (you should), but **how to configure and deploy them** in a way that balances simplicity with production requirements. This guide examines the landscape of deployment methods, from manual console clicks to infrastructure-as-code, and explains why Planton defaults to a specific approach.

## The Evolution of EKS Worker Node Management

### What Are EKS Managed Node Groups?

Amazon EKS **managed node groups** automate the provisioning and lifecycle of worker nodes for your Kubernetes cluster. When you create a managed node group, AWS creates and manages an EC2 Auto Scaling Group in your account, using EKS-optimized AMIs by default. AWS handles node **bootstrapping, upgrades, and draining** during Kubernetes version updates.

This represents a fundamental shift from the early days of EKS, when teams had to:
- Manually configure Auto Scaling Groups
- Write custom user-data scripts to join nodes to clusters  
- Implement their own node draining logic for upgrades
- Debug networking and IAM issues without guardrails

Managed node groups aim to be the **default and simplest way** to deploy EC2 compute for EKS. AWS continually improves this service—adding support for custom AMIs, launch templates, and scaling limits up to 30 node groups of 450 nodes each. In essence, managed node groups offload operational overhead to AWS while keeping nodes in your AWS account for full transparency.

### The Compute Spectrum: Managed vs. Self-Managed vs. Fargate

**Self-Managed Nodes** represent the original approach—EC2 instances you configure and manage yourself. You have complete control over EC2 settings, AMI, and lifecycle, but also complete responsibility. You must manually install Kubernetes components (kubelet, kube-proxy, CNI) and ensure instances register with your cluster. This approach makes sense only for specialized requirements that managed node groups genuinely cannot support (which is increasingly rare).

**AWS Fargate** is the opposite extreme: serverless Kubernetes where you run pods without managing any nodes. AWS launches isolated runtime environments for each pod on demand. You never worry about OS maintenance, instance right-sizing, or scaling nodes. The trade-off is less control: no SSH access, no DaemonSets, limited networking options, and no GPU support. Fargate excels for specific use cases—highly spiky workloads or services requiring maximum isolation—but costs more for steady-state workloads.

**Managed node groups** occupy the sweet spot: you keep the **flexibility of EC2** (full node access, choice of instance types, DaemonSet support, GPU workloads) while AWS handles the undifferentiated heavy lifting. For most production clusters, managed node groups are the recommended default, often combined with Fargate for specific workloads.

## The Deployment Methods Maturity Spectrum

### Level 0: The Anti-Pattern (Manual Console Creation)

Creating node groups through the AWS Management Console is possible but fundamentally flawed for production use. The console wizard walks you through specifying cluster name, instance type, and scaling settings—but this approach suffers from fatal weaknesses:

**Lack of repeatability**: Settings aren't version-controlled. Teams forget what options were chosen, leading to configuration drift between environments.

**Human error**: It's easy to accidentally allow SSH access from 0.0.0.0/0, misconfigure subnets, or forget critical tags for cluster-autoscaler.

**No automation**: Setting up dev, staging, and production environments means clicking through the wizard three times, each time risking inconsistency.

**Orphaned resources**: If you delete resources out-of-band, you can leave EC2 instances running that continue accruing costs.

**Verdict**: Use the console only for one-time experiments or initial learning. Never for persistent clusters.

### Level 1: Imperative Scripting (AWS CLI, eksctl, SDKs)

**AWS CLI** provides direct API access via commands like `aws eks create-nodegroup --cluster-name mycluster --nodegroup-name mynodes ...`. This is scriptable and repeatable—you can version control your commands—but you must handle ordering, idempotency, and error handling yourself. The CLI doesn't automatically create IAM roles or manage dependencies between resources.

**eksctl** represents a significant improvement. This purpose-built CLI (created by Weaveworks in collaboration with AWS) automates the entire cluster and node group creation process. A single command like `eksctl create cluster --name mycluster --nodes 3 --node-type t3.medium` creates the cluster, node group, VPC, subnets, and IAM roles following AWS best practices.

eksctl dramatically reduces common pitfalls by applying sensible defaults. It creates node IAM roles with correct managed policies, tags subnets properly for cluster-autoscaler, and handles resource dependencies automatically. The downside is less fine-grained control compared to writing your own infrastructure code. eksctl is opinionated, which is excellent for getting started but can be limiting in complex enterprise environments.

**SDKs and custom scripts** (Python boto3, Go AWS SDK) offer programmatic control but essentially replicate CLI functionality with more boilerplate. Teams rarely write raw scripts for provisioning unless they have very custom orchestration needs, because IaC tools have solved these problems more elegantly.

**Verdict**: eksctl is excellent for development clusters, rapid prototyping, or as a bootstrap before transitioning to full IaC. CLI and SDK approaches are too low-level for most production use cases.

### Level 2: Configuration Management (Ansible)

Ansible can manage EKS node groups via the `community.aws.eks_nodegroup` module. This allows declarative specifications in playbooks, handling idempotency and resource state. Using Ansible makes sense if you already have an Ansible-based workflow and want to integrate EKS provisioning with other system configuration.

However, compared to purpose-built IaC tools like Terraform, Ansible's experience for cloud resource creation is less smooth. Error handling and dependency management can be cumbersome. The module documentation notes sensible defaults (desired size 1, min 1, max 2), but Ansible is not the dominant pattern for EKS deployment.

**Verdict**: Viable if Ansible is already your standard, but not the first choice for greenfield EKS deployments.

### Level 3: Infrastructure as Code—The Production Standard

Infrastructure-as-Code tools are the de facto standard for deploying AWS resources in production. All major IaC frameworks support EKS managed node groups:

#### AWS CloudFormation

CloudFormation provides the `AWS::EKS::Nodegroup` resource type for declarative node group management. You specify properties like cluster name, node role, subnets, scaling config, AMI type, and launch template in JSON/YAML templates. CloudFormation handles creation, waits for active status, and provides rollback on failure.

**Strengths**:
- AWS-native, deeply integrated with AWS services
- Change sets allow preview before applying changes  
- Transactional deployments with automatic rollback
- Used internally by eksctl and the AWS Console

**Limitations**:
- Verbose template syntax with a learning curve
- Won't automatically create IAM roles (you must include `AWS::IAM::Role` resources)
- Requires managing dependencies explicitly (node groups depend on cluster existence)

CloudFormation is a solid choice, especially if you're already invested in the CloudFormation ecosystem.

#### Terraform and OpenTofu

Terraform's `aws_eks_node_group` resource allows declaring node groups in HCL (HashiCorp Configuration Language). A typical configuration references an `aws_eks_cluster` resource for the cluster name and an `aws_iam_role` resource for the node role ARN:

```hcl
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.mycluster.name  
  node_group_name = "production-nodes"
  node_role_arn   = aws_iam_role.worker_nodes.arn
  subnet_ids      = aws_subnet.private[*].id
  
  scaling_config {
    desired_size = 3
    min_size     = 3
    max_size     = 6
  }
  
  instance_types = ["m6i.xlarge"]
  
  tags = {
    Environment = "production"
  }
}
```

Terraform ensures proper ordering (cluster before node group) through references and detects configuration drift. Many teams use the **official Terraform EKS Module** which bundles cluster, node group, VPC, and IAM role creation with best practices.

**OpenTofu** (the open-source Terraform fork) functions identically for these purposes, using the same provider syntax.

**Strengths**:
- Mature ecosystem with extensive community resources
- Strong state management and drift detection
- Plan/apply workflow provides safety  
- Excellent module ecosystem for reusable patterns

**Considerations**:
- Requires careful state management (losing state can cause Terraform to attempt resource recreation)
- Some node group changes require replacement (instance type, AMI changes trigger rolling updates)

Terraform/OpenTofu is a preferred choice for many production deployments due to maturity, community support, and declarative clarity.

#### Pulumi

Pulumi enables infrastructure definition in TypeScript, Python, Go, or other general-purpose languages. The `aws.eks.NodeGroup` resource maps closely to the AWS API. Pulumi also offers a higher-level `@pulumi/eks` package providing an `eks.Cluster` abstraction that can create both cluster and node groups with minimal code.

Example in TypeScript:
```typescript
const nodeGroup = new aws.eks.NodeGroup("production", {
    clusterName: cluster.name,
    nodeRoleArn: nodeRole.arn,
    subnetIds: privateSubnets,
    scalingConfig: {
        minSize: 3,
        maxSize: 6,
        desiredSize: 3,
    },
    instanceTypes: ["m6i.xlarge"],
});
```

**Strengths**:
- Use familiar programming languages instead of DSLs
- Excellent IDE support (autocomplete, type checking)
- Easy integration with application code
- Built-in best practices (e.g., automatic dependency management)

Pulumi is particularly strong for teams who prefer general-purpose languages and want tight integration between infrastructure and application code.

#### AWS CDK

The AWS Cloud Development Kit allows defining infrastructure in TypeScript, Python, or other languages, which synthesizes to CloudFormation templates. CDK provides high-level constructs like `EksCluster` that handle common patterns automatically:

```typescript
const cluster = new eks.Cluster(stack, 'MyCluster', { 
  version: eks.KubernetesVersion.V1_28 
});

cluster.addNodegroupCapacity('OnDemandNodes', {
    instanceTypes: [new ec2.InstanceType('m6i.xlarge')],
    minSize: 3,
    desiredSize: 3,
    maxSize: 6,
    subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
});
```

If you don't specify a node IAM role, CDK creates one with appropriate managed policies automatically.

**Strengths**:
- High-level abstractions reduce boilerplate
- Automatic resource dependency handling
- Leverages CloudFormation's transactional deployments
- Strong for AWS-centric infrastructure

CDK excels for developers comfortable with code who want CloudFormation's safety with better ergonomics.

#### Comparative Analysis

| Tool | State Management | Learning Curve | Abstraction Level | Best For |
|------|-----------------|----------------|-------------------|----------|
| **CloudFormation** | AWS-managed stacks | Moderate (JSON/YAML) | Low (explicit resources) | AWS-native shops, compliance requirements |
| **Terraform/OpenTofu** | State file (remote backend) | Moderate (HCL) | Medium (modules available) | Multi-cloud, mature ecosystem needs |
| **Pulumi** | State backend | Low (familiar languages) | High (component models) | Developers, app-infra integration |
| **CDK** | CloudFormation stacks | Low-Moderate | High (constructs) | AWS-focused, developer productivity |

**Key Insight**: All these tools ultimately call the same AWS EKS API. Your choice depends on team expertise, existing tooling, and preferences around declarative DSLs versus general-purpose languages.

### Level 4: Kubernetes-Native Infrastructure (ACK, Crossplane)

**AWS Controllers for Kubernetes (ACK)** allows managing AWS resources using Kubernetes CRDs. You `kubectl apply` a YAML manifest defining a `Nodegroup` resource, and the ACK controller calls AWS to create it:

```yaml
apiVersion: eks.services.k8s.aws/v1alpha1
kind: Nodegroup
metadata:
  name: production-nodes
spec:
  clusterName: my-cluster
  nodeRole: arn:aws:iam::123456789:role/NodeRole
  subnets:
    - subnet-abc123
    - subnet-def456
  scalingConfig:
    minSize: 3
    maxSize: 6
    desiredSize: 3
  instanceTypes:
    - m6i.xlarge
```

ACK integrates well with GitOps workflows and treats infrastructure as Kubernetes resources. It's particularly powerful when combined with cluster-autoscaler—you can annotate nodegroup resources to control whether ACK or autoscaler manages desired size.

**Crossplane** offers similar Kubernetes-native IaC with cloud-agnostic composition features. You can create higher-level abstractions (Composite Resources) that generate multiple underlying resources (cluster, node groups, IAM roles) from a simplified API.

**Strengths**:
- Kubernetes-native (fits existing kubectl/GitOps workflows)
- Unified control plane for app and infrastructure
- Powerful composition (Crossplane)

**Considerations**:
- Adds complexity (requires running controllers, managing their IAM permissions)
- Not all AWS services are GA in ACK
- Learning curve for CRD-based infrastructure

**Verdict**: Excellent for platform teams building internal developer platforms, but overkill for straightforward EKS deployments.

## Production Best Practices

Regardless of deployment method, production EKS node groups require attention to several key areas:

### Scaling and Autoscaling

The **Cluster Autoscaler** automatically adjusts node count based on pod scheduling needs. AWS now automatically tags managed node group Auto Scaling Groups with `k8s.io/cluster-autoscaler/<ClusterName>` and `k8s.io/cluster-autoscaler/enabled=true` for auto-discovery.

Best practices:
- Use IAM Roles for Service Accounts (IRSA) to grant autoscaler permissions
- Configure min/max sizes intentionally (don't set min=0 for all groups)
- Enable `--balance-similar-node-groups` if running multiple groups with identical labels
- Test pod disruption budgets to ensure graceful scale-down

### Instance Type Selection

Common instance families:
- **General purpose** (m6i, m5): Balanced CPU/memory for most microservices
- **Compute optimized** (c6i, c5): CPU-intensive workloads
- **Memory optimized** (r6i, r5): Memory-heavy workloads (caches, databases)
- **Graviton** (m6g, c7g): ARM-based for 20-30% cost savings
- **GPU** (p3, g4dn, g5): ML and graphics workloads

Production pattern: **m6i.xlarge or m6i.2xlarge** (4-8 vCPU, 16-32 GiB RAM) offers a good balance for most containerized applications. For dev/test, **t3.medium** (2 vCPU, 4 GiB) is economical.

### On-Demand vs. Spot Instances

EKS makes it easy to use **Spot instances** by setting `capacityType: SPOT` on a node group. Spot instances offer 70-90% cost savings but can be reclaimed with 2 minutes notice.

**Production pattern**: 
- Run a baseline **on-demand node group** (3+ nodes) for critical services
- Add a **spot node group** (0-10+ nodes) for scalable, fault-tolerant workloads
- Use multiple instance types in spot groups to improve availability
- Deploy AWS Node Termination Handler as a DaemonSet for graceful draining
- Consider tainting spot nodes (`spot=true:NoSchedule`) to control pod placement

### Networking and Security

- Deploy nodes in **private subnets** across multiple Availability Zones
- Use **NAT Gateways** or VPC endpoints (ECR, S3) for outbound internet access
- Keep node IAM roles minimal (EKS WorkerNodePolicy + ECR read-only)
- Use **IRSA** for pod-level IAM permissions instead of node role
- Disable SSH or restrict to bastion hosts; prefer AWS Systems Manager Session Manager

### Storage and Disk Sizing

Default root volume: **20 GiB**. For production, increase to **50-100 GiB** to accommodate container images and temporary data. Configure this via the `disk_size` parameter or in a launch template.

### Upgrades and Patching

Managed node groups support rolling updates:
- Upgrade control plane first, then node groups
- Configure `maxUnavailable` or `maxUnavailablePercentage` to control update speed (default: 1 node at a time)
- Test upgrades in dev/staging first
- Consider **Bottlerocket AMI** for minimal attack surface and atomic updates

## What Planton Provides

Planton abstracts the complexity of node group deployment into a simplified, opinionated API that follows the 80/20 principle: expose the 20% of configuration that covers 80% of use cases.

### Core Design Philosophy

1. **Sensible defaults**: Minimal configuration produces a working, production-ready node group
2. **Required fundamentals**: Force explicit choices for critical settings (cluster, instance type, scaling)
3. **Advanced escape hatches**: Support power-user scenarios (custom labels, SSH keys) without cluttering the common path
4. **Multi-cloud consistency**: Same API pattern as other Planton cloud resources

### Minimal Required Configuration

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEksNodeGroup
metadata:
  name: production-workers
spec:
  cluster_name: my-cluster
  node_role_arn: arn:aws:iam::123456789:role/EksNodeRole
  subnetIds:
    - subnet-abc123
    - subnet-def456
  instanceType: m6i.xlarge
  scaling:
    min_size: 3
    max_size: 6
    desired_size: 3
```

This creates a production-ready node group with:
- **On-demand instances** (default capacity type)
- **100 GiB root volumes** (default disk size)
- **Multi-AZ deployment** (via multiple subnet IDs)
- **No SSH access** (secure by default)

### Advanced Options

```yaml
spec:
  capacity_type: spot
  disk_size_gb: 200
  ssh_key_name: my-dev-key
  labels:
    environment: production
    team: platform
```

### Under the Hood

Planton translates your specification into:
- Pulumi or Terraform code (depending on backend choice)
- Proper IAM policy validation
- CloudWatch metrics and logging integration
- Automatic tagging for cluster-autoscaler compatibility

This abstraction means you get production-ready infrastructure without managing the underlying IaC complexity directly.

## Common Anti-Patterns to Avoid

1. **Single-AZ deployment**: Always span at least two Availability Zones for resilience
2. **No autoscaling**: Fixed-size groups waste resources or can't handle traffic spikes
3. **Mixing dissimilar instances**: Don't combine m5.xlarge and m5.large in one group (confuses scheduler)
4. **Manual ASG changes**: Never edit the underlying Auto Scaling Group directly; use EKS APIs
5. **All workloads on one group**: Consider separate groups for different workload types (compute-heavy, memory-heavy, spot-tolerant)
6. **Neglecting node patches**: Regularly update to latest EKS-optimized AMIs for security

## Conclusion

The journey from manual console clicks to production-ready node group automation reflects the broader maturation of Kubernetes operations. EKS managed node groups have evolved from a basic feature to a sophisticated platform—but realizing their full potential requires the right deployment approach.

**For quick experiments**: eksctl provides the fastest path to a working cluster.

**For production infrastructure**: Use IaC tools (Terraform, Pulumi, CloudFormation, or CDK) to ensure repeatability, version control, and team collaboration.

**For platform abstraction**: Planton provides a simplified API that applies best practices automatically while allowing customization when needed.

The key insight is that managed node groups handle the operational heavy lifting—bootstrapping, upgrades, draining—but you must still configure them thoughtfully. Instance types, scaling parameters, capacity types (spot vs. on-demand), and networking decisions directly impact cost, performance, and reliability.

By understanding the deployment method spectrum and following production best practices, you can run EKS clusters that are cost-efficient, resilient, and genuinely simple to operate—not just simple to start.

