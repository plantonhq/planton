# AWS Security Group Deployment: From Click-and-Pray to Production-Grade IaC

## Introduction

"Just open it to the world for now, we'll lock it down later." Famous last words in cloud security.

AWS Security Groups are the fundamental building blocks of network security in VPC environments—virtual firewalls that control inbound and outbound traffic to your cloud resources. Yet despite their critical role, they're often managed haphazardly: configured manually through the AWS Console, left overly permissive, or scattered across disparate IaC configurations with no clear ownership or consistency.

The challenge isn't in understanding what Security Groups do (they're conceptually straightforward), but in **managing them systematically across environments, teams, and the full application lifecycle**. How do you define security group rules that are neither too restrictive (breaking applications) nor too permissive (exposing attack surfaces)? How do you handle circular dependencies when two security groups need to reference each other? How do you update rules in production without causing traffic drops?

This document explores the full spectrum of Security Group deployment methods—from manual console clicks to sophisticated IaC orchestration—and explains how Planton provides a production-ready, declarative approach that balances security, flexibility, and operational simplicity.

## The Maturity Spectrum: How Teams Deploy Security Groups

### Level 0: The Console Anti-Pattern

**What it is:** Creating and managing Security Groups through the AWS Management Console's point-and-click interface.

**Why people start here:** It's immediate, visual, and requires no tooling or code. For quick experiments or learning AWS, the console provides instant feedback.

**The trap:** Manual configuration is inherently inconsistent and risky. Common mistakes include:
- Opening `0.0.0.0/0` on sensitive ports (SSH, RDP, databases) "temporarily" and forgetting to tighten restrictions
- Not documenting the purpose of each rule, making it impossible to safely remove old rules
- Creating duplicate or conflicting rules across environments
- Losing track of which Security Groups are actually in use
- No audit trail beyond CloudTrail (which requires forensic analysis to reconstruct intent)

While the AWS Console will warn you about opening SSH to `0.0.0.0/0`, it won't prevent it. One unreviewed click can expose your production database to internet scanners.

**Verdict:** Acceptable for sandbox accounts and initial learning. Unacceptable for any environment where consistency, auditability, or team collaboration matters.

---

### Level 1: Scripted Management (AWS CLI, SDKs, Ansible)

**What it is:** Using the AWS CLI (`aws ec2 create-security-group`, `aws ec2 authorize-security-group-ingress`) or SDKs (Boto3 for Python, AWS SDK for Go/Java/Node.js) to create and modify Security Groups programmatically. Ansible's `ec2_security_group` module fits here as well—it wraps AWS API calls in declarative YAML.

**What it solves:** Repeatability. You can script the creation of a standard web-tier Security Group and run it in dev, staging, and production. Version control becomes possible (store your scripts in Git). You can templatize common patterns.

**What it doesn't solve:** State management and drift detection. If someone manually alters a Security Group, your script doesn't know. Re-running an Ansible playbook will reconcile changes *at that moment*, but drift can accumulate between runs. Scripts also struggle with dependencies—if Security Group A's rule references Security Group B, you must carefully order creation and handle errors when B doesn't exist yet.

**The pain points:**
- **Dependency hell:** Creating two Security Groups that reference each other requires creating one without rules, creating the second, then updating the first.
- **No plan/preview:** Scripts either succeed or fail. There's no "what would change" preview like Terraform's `plan`.
- **Idempotency is manual:** You must write logic to check "does this rule already exist?" to avoid errors.

**Verdict:** A step up from console, suitable for small-scale automation or one-off migrations. Not sustainable for complex, multi-tier architectures where understanding the full desired state matters.

---

### Level 2: Declarative IaC (Terraform, Pulumi, CloudFormation)

**What it is:** Defining Security Groups and their rules in declarative configuration files (HCL for Terraform, TypeScript/Python/Go for Pulumi, YAML/JSON for CloudFormation) and using the tool's state management to apply changes.

**What it solves:** This is where teams typically reach production-grade management. Declarative IaC provides:
- **Desired-state reconciliation:** "Here's what I want" vs. "execute these steps"
- **Drift detection:** Tools compare actual cloud state against declared state
- **Plan/preview:** See exactly what will change before applying it
- **Dependency graphs:** Tools automatically order resource creation based on references
- **Reusable modules:** Abstract common patterns (web tier SG, app tier SG) into modules

**How Terraform handles Security Groups:**
- You can define rules **inline** within the `aws_security_group` resource, or as **separate** `aws_security_group_rule` resources
- **Inline** is simpler (everything in one place) but less modular—different teams can't easily contribute rules
- **Separate** rule resources allow fine-grained control but require careful coordination to avoid conflicts
- Terraform's documentation explicitly warns: **don't mix inline and separate rule definitions** on the same Security Group, or you'll get perpetual diffs

A powerful community module (`terraform-aws-security-group` by terraform-aws-modules) emerged to handle nuances:
- Stable rule identifiers (key-based) to prevent unnecessary rule churn when ordering changes
- An option for **create-before-destroy** behavior: creates a new Security Group with updated rules, swaps instance attachments, then deletes the old SG—ensuring zero downtime but changing the SG ID
- Alternatively, **preserve-security-group-id** mode updates rules in place, accepting a brief window where old rules are removed before new ones are added

**How Pulumi handles Security Groups:**
- Similar to Terraform: you can define `SecurityGroup` with inline rules or use separate `SecurityGroupRule` resources
- Pulumi's documentation recommends **one rule per resource** for clarity (avoid packing multiple CIDRs into a single inline rule, which AWS treats as multiple separate rules anyway)
- Since Pulumi uses real programming languages, you can use loops and conditionals to generate rules—handy for creating a series of rules across multiple ports or environments

**How CloudFormation/CDK handles Security Groups:**
- CloudFormation templates define `AWS::EC2::SecurityGroup` resources with inline `SecurityGroupIngress` and `SecurityGroupEgress` properties, or as separate `AWS::EC2::SecurityGroupIngress` resources
- **Critical limitation:** If two Security Groups reference each other inline in the same CloudFormation template, you'll hit a circular dependency error—CloudFormation can't resolve the order. AWS documentation explicitly says: **use separate SecurityGroupIngress resources** for cross-SG references to break the cycle.
- **Update behavior quirk:** When you modify an inline rule, CloudFormation *removes the old rule then adds the new one*, causing a brief traffic gap. For zero downtime, you must orchestrate changes carefully (e.g., add the new rule first, then remove the old one across two stack updates).

The **AWS CDK** (Cloud Development Kit) abstracts away some of these limitations by automatically creating separate rule resources behind the scenes when you use the high-level `securityGroup.addIngressRule()` API.

**Verdict:** This is the **production baseline**. Terraform and Pulumi are widely adopted for managing Security Groups at scale. CloudFormation/CDK are the natural choice for AWS-native teams. All three require understanding their specific quirks (inline vs. separate, circular deps, update ordering), but they provide the state management, versioning, and automation that production demands.

---

### Level 3: Kubernetes-Native and Crossplane

**What it is:** Managing AWS Security Groups as Kubernetes Custom Resources (using Crossplane or the AWS Controllers for Kubernetes—ACK).

**What it solves:** For organizations running Kubernetes-centric infrastructure platforms, this approach unifies management:
- Security Groups are defined in YAML alongside application manifests
- Kubernetes RBAC controls who can create/modify Security Groups
- GitOps workflows (Flux, ArgoCD) apply Security Group changes the same way they deploy apps
- Continuous reconciliation: if a Security Group is manually altered, the Kubernetes controller reverts it to the declared state

**Crossplane** specifically provides `SecurityGroup` CRDs that map closely to AWS's API:
- You can define rules inline in the SecurityGroup spec, or create separate `SecurityGroupIngressRule` resources (similar to Terraform's pattern)
- Crossplane continuously reconciles, meaning drift is automatically corrected
- Cross-account and cross-cluster scenarios are supported through Crossplane's provider credentials model

**What it doesn't solve:** Added operational complexity. You're now dependent on a Kubernetes control plane to manage AWS networking. If the cluster is down or the Crossplane provider has issues, you can't update Security Groups. This trade-off makes sense for large platform teams that already operate Kubernetes as the control plane for everything, but it's overkill for teams managing a handful of VPCs.

**Verdict:** Powerful for Kubernetes-native organizations, but introduces coupling between cluster health and network infrastructure. Best suited for platform teams building internal developer platforms (IDPs) where all infrastructure is Kubernetes-managed.

---

### Level 4: Specialized Controllers (AWS Load Balancer Controller)

**What it is:** Certain Kubernetes controllers *automatically* manage Security Groups as a side effect of managing higher-level resources.

The **AWS Load Balancer Controller** (for EKS) is the prime example:
- When you create a Kubernetes `Ingress` or `Service` of type LoadBalancer, the controller provisions an AWS Application Load Balancer or Network Load Balancer
- It automatically creates and manages Security Groups for the load balancer:
  - **Frontend SG:** Controls which clients can reach the load balancer (e.g., allow 80/443 from `0.0.0.0/0` for a public-facing service)
  - **Backend SG:** A shared Security Group attached to both the load balancer and the node/pod network interfaces, allowing traffic from LB to targets without creating dozens of individual rules on every node's SG (avoiding AWS's rule limits)
- The controller tags these Security Groups (e.g., `elbv2.k8s.aws/cluster`) to indicate ownership
- You can annotate your Ingress to use an existing Security Group (`alb.ingress.kubernetes.io/security-groups`) or let the controller create one

**Why this matters:** The controller's approach to Security Groups demonstrates a production pattern: using a **shared backend SG** to avoid rule proliferation. Without it, if you had 50 load balancers all pointing to the same target instances, you'd need 50 separate inbound rules on the instance SGs (quickly hitting AWS's 60-rule limit). By having all LBs use one shared SG and allowing that SG on instances, you need only one rule.

**The catch:** You must coordinate with the controller. If you manage an EKS cluster's Security Groups with Terraform *and* use the Load Balancer Controller, ensure you don't fight over ownership. Use tags to delineate boundaries.

**Verdict:** Highly specialized automation for Kubernetes workloads on AWS. Not a general-purpose Security Group management tool, but an elegant pattern for dynamic environments.

---

## IaC Tool Comparison: Production Considerations

When evaluating IaC tools for Security Group management at scale, these dimensions matter:

| **Tool**              | **State Management** | **Drift Detection** | **Rule Management Pattern** | **Circular Dependency Handling** | **Production Readiness** |
|-----------------------|----------------------|---------------------|------------------------------|-----------------------------------|--------------------------|
| **Terraform**         | State file           | On plan             | Inline or separate           | Manual (use separate rules, add depends_on) | ✅ Widely used, mature modules |
| **Pulumi**            | State backend        | On preview          | Inline or separate           | Automatic (dependency graph)     | ✅ Production-ready, real code flexibility |
| **CloudFormation**    | Stack state          | On drift detection  | Inline or separate (required for cross-refs) | Requires separate resources or multiple stacks | ✅ AWS-native, deep integration |
| **AWS CDK**           | CF stack state       | On drift detection  | High-level API (auto-separated) | Handled automatically          | ✅ Best CloudFormation abstraction |
| **Crossplane**        | K8s etcd             | Continuous          | Inline or separate CRDs      | Handled by controller            | ⚠️ Adds K8s dependency   |
| **Ansible**           | No state file        | On playbook run     | Inline YAML                  | Manual ordering                  | ⚠️ No continuous drift correction |

**Key Insights:**

1. **Terraform** and **Pulumi** provide the most flexibility and community momentum. Both require understanding inline vs. separate rule patterns.
2. **CloudFormation/CDK** are the natural choice for AWS-only environments, with CDK providing the best developer experience (avoiding CF's verbosity and circular-dep pitfalls).
3. **Crossplane** is overkill unless you're already using Kubernetes as your control plane for all infrastructure.
4. **Ansible** falls short on continuous state management but can be useful for ad-hoc migrations or integrations with existing config management workflows.

---

## Production Best Practices

### 1. Principle of Least Privilege

**Open only the minimum necessary ports and sources.** For example:
- A public web server needs TCP 80/443 from `0.0.0.0/0`, but **not** SSH from the world. Allow SSH only from a bastion host or corporate VPN CIDR.
- A database should accept traffic **only** from the application tier's Security Group, never from the internet.

**Anti-pattern:** Opening port 22 or 3389 (RDP) to `0.0.0.0/0` "temporarily." Attackers scan continuously. Unless you're running an SSH honeypot, there's no reason for this.

### 2. CIDR vs. Security Group References

**When to use CIDR blocks:**
- External access (corporate office IPs, known partner IPs)
- Cross-VPC or cross-account traffic (where SG references don't work by default)

**When to use Security Group IDs:**
- Internal communication between tiers (web → app, app → database)
- Autoscaling environments where IP addresses change dynamically
- Conveying intent ("allow traffic from app-tier SG" is clearer than "allow 10.0.2.0/24")

**Example:** A three-tier architecture:
- **Web SG:** Allow 80/443 from `0.0.0.0/0`
- **App SG:** Allow 8080 from Web SG (not from a CIDR—instances scale, IPs change)
- **DB SG:** Allow 5432 (PostgreSQL) from App SG only

### 3. Self-Referencing Rules for Cluster Communication

By default, instances in the same Security Group **cannot** talk to each other unless you create a rule allowing it.

**Pattern:** For Cassandra, Elasticsearch, or any clustered application where nodes need to communicate, add an ingress rule where the source is the Security Group's own ID.

In AWS terms: `source_security_group_id = <this SG's ID>`  
In Planton: `self_reference = true`

This allows all cluster members to communicate on the required ports (e.g., gossip, replication) without opening those ports to the broader network.

### 4. Handling AWS Limits

**Default limits:**
- 60 inbound rules and 60 outbound rules per Security Group (can be increased to 100 or 500 per direction via support request, but 1000 aggregated across all SGs on an ENI is an absolute ceiling)
- 5 Security Groups per network interface (by default)

**Strategies when approaching limits:**
- **Consolidate CIDR blocks:** Instead of listing individual `/32` IPs, aggregate into larger subnets where possible
- **Use AWS Managed Prefix Lists:** A prefix list can contain 100+ CIDRs and counts as a single rule when referenced in a Security Group
- **Layer Security Groups:** Attach multiple SGs to an instance for different purposes (common restrictions SG + app-specific SG), but keep the total rule count manageable

### 5. Avoid Update Gaps (Zero-Downtime Changes)

**The problem:** When you modify a rule, most IaC tools will remove the old rule *then* add the new rule. In that brief window, traffic is blocked.

**Solutions:**
- **Create-before-destroy at the SG level:** Create a new Security Group with the new rules, attach it to instances, then delete the old SG. The downside: the SG ID changes, which can complicate references.
- **Add-then-remove at the rule level:** If using separate rule resources, first add the new rule, then in a subsequent apply, remove the old one.
- **Test in staging first:** Always validate rule changes in a non-production environment to confirm traffic isn't disrupted.

CloudFormation explicitly suffers from this (removes then adds inline rules). Terraform's behavior depends on how rules are structured. Using a battle-tested module (like `terraform-aws-security-group`) can abstract away these issues.

### 6. Monitoring and Auditing

- **Enable VPC Flow Logs** to see accepted and rejected traffic. Rejected flows indicate blocked traffic (either by Security Groups or NACLs)—useful for troubleshooting or detecting scans.
- **Use AWS Config rules** to flag risky configurations (e.g., "Security Group allows 0.0.0.0/0 on port 22").
- **Integrate with AWS Security Hub** for automated compliance checks.
- **Tag Security Groups** with owner, environment, and purpose to enable accountability and automated cleanup of unused SGs.

---

## The 80/20 Configuration Analysis

Most Security Group use cases fall into a handful of common patterns. Here's what 80% of teams need:

### Common Scenarios

**1. Public Web Server**
- **Inbound:** TCP 80 and 443 from `0.0.0.0/0` (and `::/0` for IPv6)
- **Outbound:** All (default)
- **Optional:** SSH (22) from a bastion/VPN CIDR

**2. Internal Application Server**
- **Inbound:** Application port (e.g., 8080) from Web SG, SSH from bastion SG
- **Outbound:** All or restricted to database and external APIs

**3. Database Server**
- **Inbound:** Database port (e.g., 3306, 5432) from App SG only
- **Outbound:** All (for updates/monitoring) or restricted

**4. Cluster (Self-Referencing)**
- **Inbound:** All from self (allow intra-cluster traffic), NodePort range from Load Balancer SG
- **Outbound:** All

### Minimal API Fields to Support These

Based on these patterns, a minimal Security Group API needs:

**SecurityGroup resource:**
- `vpc_id` (required)
- `description` (required, immutable in AWS)
- `ingress` (list of rules)
- `egress` (list of rules, default to "allow all" if not specified)

**SecurityGroupRule:**
- `protocol` (tcp, udp, icmp, -1 for all)
- `from_port` and `to_port` (integers, define range)
- `ipv4_cidrs` (list of IPv4 blocks)
- `ipv6_cidrs` (list of IPv6 blocks)
- `source_security_group_ids` (list, for ingress)
- `destination_security_group_ids` (list, for egress)
- `self_reference` (boolean, convenience for self-referencing)
- `description` (optional, for documentation)

This covers the vast majority of use cases. Advanced needs (AWS Managed Prefix Lists, ICMP type/code) can be added incrementally.

---

## Planton's Approach

Planton provides a **production-grade, declarative Security Group API** that balances simplicity for common cases with flexibility for advanced scenarios.

### Design Philosophy

1. **Inline rules in the API, smart implementation under the hood**  
   Users define Security Groups and their rules as a cohesive unit in protobuf. Internally, Planton creates the Security Group first, then applies rules, mirroring the AWS API's behavior. This avoids forcing users to choose between "inline" and "separate"—the API presents a clean abstraction.

2. **Support both CIDR and Security Group references**  
   Rules can specify `ipv4_cidrs`, `ipv6_cidrs`, or `source_security_group_ids` (for ingress) and `destination_security_group_ids` (for egress). This handles both external access (CIDRs) and internal tier-to-tier communication (SG references).

3. **Self-referencing made explicit**  
   The `self_reference` boolean makes cluster communication patterns clear. Instead of requiring users to figure out how to reference "the group's own ID," you just set `self_reference = true`.

4. **Sensible defaults**  
   If no egress rules are specified, Planton defaults to allowing all outbound traffic (matching AWS's default behavior and user expectations). The `description` field is required, preventing the common mistake of creating Security Groups with no context.

5. **Automatic handling of circular dependencies**  
   If two Security Groups reference each other, Planton's orchestration layer can create both SGs first, then populate rules in a second step. Users don't need to manually split operations.

6. **Tagging and metadata**  
   All Security Groups created by Planton are automatically tagged with `ManagedBy: Planton`, ensuring clear ownership and enabling automated cleanup or auditing.

7. **Validation and linting**  
   Planton can warn (or optionally block) risky configurations—e.g., opening SSH to `0.0.0.0/0` in production—guiding users toward best practices without being overly restrictive.

### Comparison to Raw IaC

| **Aspect**                  | **Terraform/Pulumi (raw)** | **Planton** |
|-----------------------------|----------------------------|---------------------|
| Inline vs. separate rules   | User must choose, can conflict | Abstracted away (feels inline, implemented separately) |
| Circular dependencies       | Manual workaround required | Handled automatically |
| Self-referencing            | Specify own SG ID manually | `self_reference = true` |
| Default egress              | Must specify explicitly    | Defaults to allow-all, can override |
| Multi-cloud abstraction     | AWS-specific               | Unified API across AWS, GCP, Azure (for SG equivalents) |
| Validation                  | Linting via external tools | Built-in best-practice checks |

### When to Use Planton's Security Group API

- **Multi-cloud environments:** If you're managing infrastructure across AWS, GCP, and Azure, Planton's unified API reduces cognitive load.
- **Platform teams:** Building internal developer platforms where application teams shouldn't need to understand AWS Security Group nuances.
- **Complex architectures:** Multi-tier apps with many Security Groups where automatic dependency resolution and validation reduce toil.

### When Raw IaC Might Be Better

- **Single-cloud, deep AWS integration:** If you're only on AWS and need access to every SG parameter (e.g., prefix lists, ICMP type/code granularity that isn't yet in Planton's API), raw Terraform/Pulumi provides that.
- **Existing mature Terraform modules:** If your team already has battle-tested Terraform modules for Security Groups with custom logic, migration cost may not justify switching.

---

## Conclusion

AWS Security Groups are deceptively simple in concept but operationally complex at scale. The progression from manual console clicks to declarative IaC isn't just about automation—it's about **establishing a single source of truth, enabling team collaboration, preventing configuration drift, and applying security best practices consistently**.

For teams reaching production maturity, Terraform and Pulumi are the industry-standard tools, each with trade-offs around rule management patterns and update semantics. CloudFormation/CDK remain the best choice for AWS-native teams willing to work within CloudFormation's constraints (or let CDK abstract them away).

**Planton sits at the next level of abstraction:** providing a declarative, production-ready Security Group API that handles the sharp edges (circular dependencies, inline vs. separate, default egress) while remaining flexible enough for advanced use cases. Whether you're building a three-tier web app or a multi-account, multi-cloud platform, Planton's approach ensures Security Groups are defined clearly, versioned, and deployed with confidence.

The paradigm shift isn't just technical—it's cultural. When Security Groups are code-reviewed, tested in staging, and deployed via CI/CD, security stops being an afterthought and becomes an integral part of the development workflow. That's the difference between "we'll lock it down later" and "it's secure by default."

