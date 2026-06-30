# AWS EKS: The Infrastructure-as-Code Landscape

## Introduction: Choosing Your Deployment Method

Ask "How should I deploy an EKS cluster?" and you'll receive a dozen different answers. AWS console for quick experiments. eksctl for simplicity. Terraform for infrastructure consistency. CDK for developers who think in code. CloudFormation for AWS-native orchestration. Pulumi for modern polyglot infrastructure.

Every tool works. But **not every tool scales** from prototype to production.

The real question isn't "Can I create an EKS cluster?" but rather "Can I **manage**, **upgrade**, **secure**, and **replicate** EKS clusters across environments in a way that's auditable, repeatable, and integrated with my existing workflows?"

This guide explains the deployment landscape, what production-grade EKS requires, and why Planton implements EKS provisioning with Pulumi as the default while maintaining an abstraction layer that keeps you flexible.

---

## The Maturity Spectrum: From Clicks to Code

Deploying an EKS cluster involves two fundamental layers:

1. **Control Plane**: The Kubernetes API server, scheduler, and etcd—fully managed by AWS
2. **Data Plane**: Worker nodes (EC2, Fargate) where your workloads run

Different deployment approaches handle these layers with varying degrees of automation, repeatability, and operational sophistication.

### Level 0: The Manual Approach (AWS Console & CLI)

**AWS Console** provides a step-by-step wizard. Click through, select your VPC, name your cluster, and AWS provisions the control plane.

**AWS CLI** offers scriptability: `aws eks create-cluster --name prod-cluster --role-arn ...` calls the same APIs.

**What's Missing**: 
- No source of truth. What was configured? When? By whom?
- No dependency management. You must manually create VPCs, IAM roles, security groups, node groups—in the correct order—and track their relationships.
- No change management. Want to update something? Better remember what you clicked six months ago.
- No environment parity. Your dev cluster config exists in someone's bash history; your prod cluster config is a mystery.

**Verdict**: Acceptable for learning or throwaway experiments. Unacceptable for production. You're one console click away from an undocumented, unrepeatable configuration.

### Level 1: The Automation Layer (eksctl)

**eksctl** is AWS and Weaveworks' purpose-built CLI for EKS. One command or a YAML config file can create a cluster with sensible defaults:

```yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: prod-cluster
  region: us-west-2
nodeGroups:
  - name: workers
    instanceType: m5.large
    desiredCapacity: 3
```

Run `eksctl create cluster -f cluster.yaml` and it orchestrates:
- VPC creation (if needed)
- EKS cluster control plane
- IAM roles with correct policies
- Managed node groups
- AWS-auth ConfigMap for node registration
- OIDC provider for IAM Roles for Service Accounts (IRSA)

**What It Solves**: Removes the toil of coordinating AWS resources. No more "did I attach the right policy?" or "why won't my nodes join the cluster?"

**What's Missing**:
- Limited state management. eksctl uses CloudFormation stacks under the hood, but it's not a full IaC framework.
- No drift detection. If someone modifies the cluster via console, eksctl won't know or reconcile.
- AWS-only. Your infrastructure likely includes more than EKS (databases, object storage, DNS). eksctl can't manage those.
- Upgrade automation is basic. Complex scenarios (blue-green node replacements, controlled rollouts) require manual orchestration.

**Verdict**: Excellent for bootstrapping and small-scale deployments. Many teams use eksctl to create clusters, then manage Day 2 operations with other tools. Not a complete infrastructure management solution.

### Level 2: The Foundation (Infrastructure as Code)

This is where production-grade deployments live. **Terraform**, **Pulumi**, **AWS CDK**, and **CloudFormation** treat infrastructure as versioned, testable, peer-reviewed code.

Key properties:
- **Declarative**: Describe the desired state; the tool figures out how to get there.
- **Stateful**: Track what was deployed. Detect drift. Plan changes before applying.
- **Composable**: Cluster creation integrates with VPC provisioning, IAM management, monitoring setup—your entire stack in one coherent definition.
- **Auditable**: Infrastructure changes go through Git, code review, CI/CD pipelines.

**How It Works**:
1. Define your EKS cluster in code (HCL for Terraform, TypeScript/Python/Go for Pulumi/CDK)
2. The tool calls AWS APIs to create/update resources
3. State is tracked (in Terraform state files, Pulumi backends, CloudFormation stacks)
4. Changes are versioned, reviewed, and applied via automation

**What This Enables**:
- **Repeatability**: Deploy identical clusters across dev, staging, prod
- **Change Management**: See exactly what will change before it happens (`terraform plan`, `pulumi preview`)
- **Disaster Recovery**: Cluster destroyed? Redeploy from code in minutes.
- **Multi-Cloud Patterns**: Terraform and Pulumi aren't AWS-specific. Manage GKE, AKS, and on-prem Kubernetes with the same toolchain.

**Verdict**: This is the standard for production infrastructure. The question isn't *if* you should use IaC, but *which* IaC tool aligns with your team's workflows.

---

## The IaC Tool Comparison: Terraform, Pulumi, CDK, CloudFormation

All four tools can deploy production-grade EKS clusters. The differences lie in **language**, **ecosystem**, **abstraction level**, and **operational model**.

### Terraform

**Language**: HCL (HashiCorp Configuration Language)  
**State Management**: Remote backends (S3 + DynamoDB, Terraform Cloud)  
**Philosophy**: Declarative, cloud-agnostic, battle-tested

**Strengths**:
- Massive ecosystem. The AWS provider exposes every EKS feature.
- Strong community modules (e.g., `terraform-aws-modules/eks/aws`) encapsulate best practices
- Excellent plan/apply workflow with clear change preview
- Multi-cloud capability (manage AWS + GCP + Azure with one tool)

**EKS Workflow**:
```hcl
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  cluster_name    = "prod-cluster"
  cluster_version = "1.31"
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  eks_managed_node_groups = {
    general = {
      instance_types = ["m5.large"]
      min_size       = 2
      max_size       = 6
      desired_size   = 3
    }
  }
}
```

**Considerations**:
- HCL has limitations. No loops, limited conditionals, no native testing frameworks.
- State management requires discipline (locking, encryption, backup).
- Some advanced scenarios (custom logic, complex conditionals) feel awkward.

**Best For**: Teams with existing Terraform expertise, multi-cloud requirements, or infrastructure codebases spanning thousands of resources.

### Pulumi

**Language**: Real programming languages (TypeScript, Python, Go, C#, Java)  
**State Management**: Pulumi Cloud (SaaS) or self-hosted backends  
**Philosophy**: Infrastructure as actual code, with all the tooling that implies

**Strengths**:
- Use familiar languages with loops, conditionals, functions, classes, testing frameworks
- First-class IDE support (autocomplete, type checking, inline docs)
- Component model for building high-level abstractions
- Native secrets management (encrypted in state)

**EKS Workflow**:
```typescript
import * as eks from "@pulumi/eks";

const cluster = new eks.Cluster("prod-cluster", {
    vpcId: vpc.id,
    subnetIds: vpc.privateSubnetIds,
    version: "1.31",
    instanceType: "m5.large",
    desiredCapacity: 3,
    minSize: 2,
    maxSize: 6,
    enabledClusterLogTypes: ["api", "audit"],
});
```

**Considerations**:
- Smaller community than Terraform (though growing rapidly)
- Team must be comfortable with programming (which most modern infra teams are)
- Pulumi Cloud SaaS is convenient but adds a dependency (self-hosting available)

**Best For**: Teams that prefer real programming languages, want to build custom infrastructure components, or value integrated testing/validation.

### AWS CDK

**Language**: TypeScript, Python, Java, C#  
**State Management**: CloudFormation stacks  
**Philosophy**: AWS-native with high-level constructs

**Strengths**:
- Extremely high-level abstractions (L2 constructs). A few lines create a fully-configured EKS cluster with best practices.
- Tight AWS integration. New AWS features often appear in CDK before Terraform.
- Synthesizes to CloudFormation, so it's "official" AWS IaC.

**EKS Workflow**:
```typescript
import * as eks from 'aws-cdk-lib/aws-eks';

const cluster = new eks.Cluster(this, 'ProdCluster', {
  version: eks.KubernetesVersion.V1_31,
  defaultCapacity: 3,
  defaultCapacityInstance: ec2.InstanceType.of(ec2.InstanceClass.M5, ec2.InstanceSize.LARGE),
});
```

**Considerations**:
- AWS-only. Cannot manage GCP, Azure, or on-prem resources.
- CloudFormation limitations (stack size limits, slower deploys, limited rollback scenarios)
- Abstractions are powerful but opaque (synthesized CloudFormation can be thousands of lines)

**Best For**: AWS-first organizations, teams that want maximum abstraction, or those building on existing CDK investments.

### AWS CloudFormation

**Language**: JSON or YAML templates  
**State Management**: CloudFormation service (AWS-managed)  
**Philosophy**: AWS-native, declarative templates

**Strengths**:
- Official AWS offering. Every EKS feature is supported on day one.
- StackSets enable multi-account, multi-region deployments
- No external dependencies (no state backends to manage)

**EKS Workflow**:
```yaml
Resources:
  EKSCluster:
    Type: AWS::EKS::Cluster
    Properties:
      Name: prod-cluster
      Version: "1.31"
      RoleArn: !GetAtt ClusterRole.Arn
      ResourcesVpcConfig:
        SubnetIds:
          - !Ref PrivateSubnet1
          - !Ref PrivateSubnet2
```

**Considerations**:
- YAML/JSON verbosity. Complex configurations become unwieldy.
- Limited modularity compared to code-based tools.
- Stack updates can be slow and sometimes get stuck.
- AWS-only (like CDK).

**Best For**: Organizations with deep AWS platform teams, regulatory requirements for AWS-native tools, or existing CloudFormation investments.

---

## Why Planton Uses Pulumi (But Abstracts It)

Planton's EKS implementation defaults to **Pulumi** for several reasons:

### 1. Abstraction Through Code

Pulumi's programming model allows us to build **high-level components** that encapsulate best practices. Our `AwsEksCluster` API doesn't expose every Pulumi knob—it exposes the **80% of configuration** that matters:

- Subnet IDs (where to deploy)
- Cluster IAM role (permissions)
- Kubernetes version
- Public/private endpoint toggle
- Control plane logging
- KMS encryption key

The remaining 20% of edge cases (custom OIDC providers, outpost configurations, IPv6 clusters) can be handled via advanced fields or custom Pulumi programs.

### 2. The 80/20 Configuration Philosophy

Based on production cluster research, most EKS deployments share common patterns:

**Essential Configuration (80% of use cases)**:
- Multi-AZ subnets for high availability
- Private API endpoint (or public with CIDR restrictions)
- Control plane logging to CloudWatch
- KMS encryption for secrets
- Managed node groups with on-demand or spot instances

**Rare Configuration (20% of use cases)**:
- Custom service IP CIDR ranges
- IPv6 networking
- AWS Outposts placement
- Custom security groups beyond defaults

Our API surfaces the essential 80%. This keeps simple cases simple while allowing advanced users to drop down to Pulumi's native SDK when needed.

### 3. Manager-Agnostic API Design

Just as our `PostgresKubernetes` API doesn't lock you into Zalando operator, our `AwsEksCluster` API doesn't lock you into Pulumi.

The protobuf API (`spec.proto`) defines **what** you want:

```protobuf
message AwsEksClusterSpec {
  repeated StringValueOrRef subnet_ids = 1;
  StringValueOrRef cluster_role_arn = 2;
  string version = 3;
  bool disable_public_endpoint = 4;
  bool enable_control_plane_logs = 6;
  StringValueOrRef kms_key_arn = 7;
}
```

The IaC implementation (Pulumi, Terraform, CDK) defines **how** to achieve it. This separation means:
- Teams can implement Terraform modules using the same API
- Organizations can swap IaC backends without changing application manifests
- The API evolves independently of tooling trends

### 4. Integrated Development Experience

Pulumi's TypeScript (or Python/Go) offers:
- **Type safety**: Catch errors before deployment (typo in a subnet ID? IDE flags it)
- **Testing**: Unit test your infrastructure (does this config create a private cluster?)
- **Modularity**: Build reusable components (a "SecureEksCluster" class with encryption and logging baked in)

This developer experience compounds as infrastructure complexity grows.

---

## Production-Grade EKS: The Non-Negotiables

Provisioning an EKS cluster is the beginning, not the end. Production readiness requires:

### High Availability

**Multi-AZ Control Plane**: AWS runs the EKS control plane across at least three availability zones automatically. But you must provide subnets in **at least two AZs**—this is why `subnet_ids` requires a minimum of two entries.

**Multi-AZ Worker Nodes**: Distribute node groups across AZs. If `us-west-2a` experiences an outage, your workloads fail over to nodes in `us-west-2b` and `us-west-2c`.

### Network Security

**Private Endpoint**: Set `disable_public_endpoint: true` to make the Kubernetes API accessible only within your VPC. Access via VPN, Direct Connect, or bastion hosts.

**CIDR Restrictions**: If public access is required, use `public_access_cidrs` to limit it to trusted IP ranges (your office, CI/CD runners).

**Private Subnets for Nodes**: Worker nodes should run in private subnets with no direct internet route. Use NAT Gateways for outbound connectivity or VPC endpoints for AWS services (ECR, S3, CloudWatch).

### IAM Security

**IAM Roles for Service Accounts (IRSA)**: Instead of granting broad IAM permissions to all pods on a node, use IRSA to give individual Kubernetes service accounts scoped IAM roles.

Example: Your pod needs S3 access. Create an IAM role with S3 permissions, annotate the Kubernetes ServiceAccount with the role ARN, and the pod assumes that role—and only that role.

This requires enabling an OIDC provider for your cluster (typically done automatically by IaC tools).

**Least Privilege Everywhere**: The cluster's IAM role should have minimal permissions. Node IAM roles should follow the same principle. Regularly audit with tools like AWS IAM Access Analyzer.

### Secrets Encryption

**Envelope Encryption with KMS**: By default, Kubernetes secrets are base64-encoded (not encrypted). Enable envelope encryption by providing a KMS key ARN via `kms_key_arn`.

AWS encrypts each secret with a data encryption key (DEK), then encrypts the DEK with your KMS key. Even if someone gains access to etcd, they cannot decrypt secrets without the KMS key.

**Critical**: This must be enabled at cluster creation. You cannot add it retroactively without recreating the cluster.

### Observability

**Control Plane Logs**: Enable API server, audit, authenticator, controller manager, and scheduler logs via `enable_control_plane_logs: true`.

These logs go to CloudWatch and are essential for:
- Security audits (who accessed what?)
- Debugging (why is the scheduler failing to place pods?)
- Compliance (regulatory requirements for audit trails)

**Application Logs**: Deploy a log aggregator (Fluent Bit, Fluentd) as a DaemonSet to ship pod logs to CloudWatch, Elasticsearch, or your SIEM.

**Metrics**: Use Prometheus (self-hosted or Amazon Managed Prometheus) with Grafana for cluster and application metrics. Enable CloudWatch Container Insights for AWS-native monitoring.

### Node Provisioning Strategy

**Managed Node Groups (Default)**: AWS handles node lifecycle, rolling updates, and cordoning/draining during upgrades. Use these unless you have specific requirements.

**Self-Managed Nodes**: For custom AMIs, specialized OS configurations, or unique bootstrap scripts. You handle upgrades, patching, and lifecycle.

**AWS Fargate**: Serverless pods with no node management. Great for isolation, bursty workloads, or dev environments. Not suitable for DaemonSets or GPU workloads.

**Mixed Strategy**: Run baseline workloads on managed on-demand nodes, burst capacity on spot instances, and isolated/untrusted workloads on Fargate—all in one cluster.

### Cost Optimization

**Spot Instances**: For fault-tolerant workloads, spot instances can reduce compute costs by 70-90%. Use pod disruption budgets and over-provision slightly to handle spot interruptions.

**Cluster Autoscaler / Karpenter**: Automatically scale nodes based on pod demand. Karpenter is AWS's newer, faster alternative with better instance type optimization.

**Right-Sizing**: Use metrics to identify over-provisioned pods. Reducing resource requests allows tighter bin-packing and fewer nodes.

**Shared Ingress**: Use one ALB (via AWS Load Balancer Controller) with path-based routing for multiple services instead of one ALB per service.

---

## The API: Minimal Config, Maximum Safety

Our `AwsEksClusterSpec` follows the 80/20 principle rigorously:

```protobuf
message AwsEksClusterSpec {
  // Required: At least 2 subnets in different AZs
  repeated StringValueOrRef subnet_ids = 1;
  
  // Required: IAM role for the cluster
  StringValueOrRef cluster_role_arn = 2;
  
  // Optional: Defaults to latest supported version
  string version = 3;
  
  // Optional: Defaults to false (public endpoint enabled)
  bool disable_public_endpoint = 4;
  
  // Optional: Restrict public access (defaults to 0.0.0.0/0)
  repeated string public_access_cidrs = 5;
  
  // Optional: Defaults to false (no logs)
  bool enable_control_plane_logs = 6;
  
  // Optional: No encryption by default
  StringValueOrRef kms_key_arn = 7;
}
```

**What's NOT in the API**:
- Node group definitions (managed separately, often dynamically)
- Add-on configurations (Cluster Autoscaler, ALB Controller—deployed post-cluster)
- OIDC provider setup (handled automatically)
- VPC creation (separate resource; clusters reference existing VPCs)

**Why This Matters**:

A production cluster might have dozens of node groups with different instance types, taints, and labels. Those are workload-specific and change frequently. The **control plane configuration** changes rarely—maybe on major upgrades or security policy updates.

By separating concerns, we make the common case simple (create a cluster) while keeping the complex case manageable (customize node groups, add-ons, and workload placement separately).

---

## Minimal Configuration Examples

### Development Cluster

```yaml
apiVersion: dev.planton.provider.aws.awsekscluster.v1
kind: AwsEksCluster
metadata:
  name: dev-eks
spec:
  subnetIds:
    - subnet-abc123
    - subnet-def456
  cluster_role_arn: arn:aws:iam::123456789012:role/EksClusterRole
  version: "1.31"
```

Accepts defaults:
- Public endpoint (for easy developer access)
- No control plane logs (cost savings)
- No KMS encryption

### Staging Cluster

```yaml
apiVersion: dev.planton.provider.aws.awsekscluster.v1
kind: AwsEksCluster
metadata:
  name: stage-eks
spec:
  subnetIds:
    - subnet-abc123
    - subnet-def456
  cluster_role_arn: arn:aws:iam::123456789012:role/EksClusterRole
  version: "1.31"
  disable_public_endpoint: true
  enable_control_plane_logs: true
  kms_key_arn: arn:aws:kms:us-west-2:123456789012:key/abc-123
```

Mirrors production:
- Private endpoint
- Logging enabled (test your log aggregation)
- Encryption enabled (test key rotation, access patterns)

### Production Cluster

```yaml
apiVersion: dev.planton.provider.aws.awsekscluster.v1
kind: AwsEksCluster
metadata:
  name: prod-eks
spec:
  subnetIds:
    - subnet-abc123  # us-west-2a
    - subnet-def456  # us-west-2b
    - subnet-ghi789  # us-west-2c
  cluster_role_arn: arn:aws:iam::123456789012:role/EksClusterRole
  version: "1.31"
  disable_public_endpoint: true
  public_access_cidrs:
    - 203.0.113.0/24  # Corporate VPN exit IPs
  enable_control_plane_logs: true
  kms_key_arn: arn:aws:kms:us-west-2:123456789012:key/abc-123
```

Full production hardening:
- Three AZs for maximum availability
- Private endpoint with fallback public access from known IPs
- All control plane logs enabled
- Customer-managed KMS key for secrets

---

## Lifecycle Management: Upgrades, Backups, Disaster Recovery

### Cluster Upgrades

EKS supports **in-place control plane upgrades**. Change `version: "1.30"` to `version: "1.31"` and apply.

**Best Practices**:
1. Test in staging first
2. Upgrade control plane, then nodes (EKS supports one minor version skew)
3. Review Kubernetes deprecation announcements (some APIs are removed in new versions)
4. Update add-ons (CoreDNS, VPC CNI) to versions compatible with the new Kubernetes release
5. Upgrade node groups via rolling replacement (blue-green pattern: create new node group, drain old)

**Frequency**: Upgrade at least once per year. EKS supports each Kubernetes version for ~14 months.

### Backup & Disaster Recovery

**Cluster State**: EKS's etcd is managed and backed up by AWS. But your **Kubernetes resources** (Deployments, ConfigMaps, Secrets) are your responsibility.

Use **Velero** to back up cluster resources and persistent volumes to S3:
```bash
velero backup create prod-backup --include-namespaces '*'
```

Schedule daily backups and test restores regularly.

**Cross-Region DR**: For mission-critical workloads, maintain a standby cluster in another region. Replicate persistent data (databases, object storage) cross-region and fail over DNS during outages.

### Add-On Management

Deploy critical add-ons via GitOps (ArgoCD, Flux):
- AWS Load Balancer Controller (for ingress)
- External DNS (for automatic DNS updates)
- Cluster Autoscaler or Karpenter (for node scaling)
- Metrics Server (for HPA)

This ensures add-ons are version-controlled and automatically reconciled if someone manually deletes them.

---

## Common Anti-Patterns to Avoid

### 1. Single-AZ Deployments

Providing only one subnet or all subnets in one AZ eliminates high availability. EKS will reject single-AZ control planes for production.

### 2. Public API Endpoints with No CIDR Restrictions

Leaving `0.0.0.0/0` access on a public endpoint exposes your cluster to internet-wide login attempts and potential attacks. Always restrict to known IPs or use private-only.

### 3. No Control Plane Logging

Operating blind. When something breaks (and it will), you'll wish you had API audit logs. The cost is negligible compared to the value during incidents.

### 4. Hardcoding Credentials in Pods

Storing AWS access keys in Kubernetes Secrets defeats the purpose of IAM. Use IRSA. Always.

### 5. Not Upgrading

Falling multiple versions behind creates security vulnerabilities and makes upgrades risky (big jumps often break things). Upgrade incrementally and regularly.

### 6. Over-Privileged IAM Roles

Granting `AdministratorAccess` to cluster roles or node roles violates least privilege. Scope permissions to exactly what's needed.

---

## Conclusion: Production EKS, Abstracted and Repeatable

The EKS deployment landscape has matured. The console and CLI were yesterday's tools. **Infrastructure as Code is today's standard**.

Among IaC options, no tool is universally "best"—but all production-grade deployments share common traits:
- **Repeatable**: Codified, version-controlled, peer-reviewed
- **Secure**: Private endpoints, least-privilege IAM, encrypted secrets, audit logs
- **Observable**: Metrics, logs, and alerts integrated from day one
- **Resilient**: Multi-AZ, tested disaster recovery, automated failover

Planton's `AwsEksCluster` API abstracts these patterns into a simple, declarative interface. The 80% case—a secure, observable, multi-AZ cluster—requires six fields. The 20% edge cases remain accessible via the underlying IaC layer.

By defaulting to Pulumi while keeping the API manager-agnostic, we give you **flexibility without fragmentation**. You get production-grade defaults backed by AWS best practices, but you're never locked in.

Deploy with confidence. Upgrade with clarity. Scale without surprises.

That's modern EKS infrastructure.

---

## Further Reading

- [AWS EKS Best Practices Guide](https://aws.github.io/aws-eks-best-practices/) - Official AWS recommendations
- [Amazon EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/) - Comprehensive EKS documentation
- [Pulumi AWS EKS Examples](https://github.com/pulumi/examples/tree/master/aws-ts-eks) - Reference implementations
- [EKS Terraform Blueprints](https://github.com/aws-ia/terraform-aws-eks-blueprints) - AWS's curated Terraform modules
- [IRSA Deep Dive](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) - Understanding IAM Roles for Service Accounts

