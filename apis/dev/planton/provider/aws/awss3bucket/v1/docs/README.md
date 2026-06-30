# AWS S3 Bucket Deployment: From Manual Clicks to Production Infrastructure

## Introduction

Amazon S3 is deceptively simple. A bucket is just a container for objects, right? Point and click in the AWS console, and you're done. Yet this simplicity masks a crucial reality: **how you deploy and configure S3 buckets determines whether your data remains secure, cost-effective, and resilient—or becomes tomorrow's breach headline.**

The journey from "let me quickly create a bucket" to "we have a production-grade, compliant storage infrastructure" involves navigating deployment methods that range from dangerous anti-patterns to battle-tested automation. Understanding this spectrum isn't academic—every year, misconfigured S3 buckets expose billions of records because teams treated deployment as a one-time console click rather than managed infrastructure.

This document explains the landscape of S3 deployment approaches, why modern infrastructure-as-code (IaC) has become essential, and how Planton abstracts the complexity while preserving production-grade security and flexibility.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern – Manual Console Configuration

**What it is:** Creating buckets through the AWS Management Console, pointing and clicking through settings, and uploading objects through the web interface.

**Why it exists:** It's the fastest way to get started. AWS makes it easy—a few clicks and you have a bucket. For learning S3 concepts or one-off experiments, it's convenient.

**The trap:** Manual console work is inherently unrepeatable and error-prone. Did you remember to enable Block Public Access? Did you set the correct encryption? What about the lifecycle rules you meant to add? When you create your staging bucket tomorrow, will you configure it identically to production?

**Real-world consequences:** The majority of S3 data breaches trace back to manual misconfigurations—a forgotten Block Public Access setting, an overly permissive bucket policy, or default encryption never enabled. These aren't theoretical risks; they've exposed medical records, financial data, and customer information at scale.

**Verdict:** Acceptable only for throwaway experiments in personal AWS accounts. Never for anything that might contain real data or need to be reproduced.

### Level 1: Scripting with CLI and SDKs

**What it is:** Using AWS CLI commands (`aws s3api create-bucket`) or SDKs (Boto3, AWS SDK for JavaScript) to programmatically create buckets and configure settings.

**Why it's better:** You can capture your bucket creation in a script, making it repeatable. Run the script again and you'll get the same bucket (assuming the name isn't taken). This integrates into CI/CD pipelines and can be version-controlled.

**What it solves:** Repeatability. Two developers running the same script will create buckets with identical settings. You can review the script in code review, catching security mistakes before they reach production.

**What it doesn't solve:** Idempotency and state management. Run your creation script twice and it fails because the bucket exists. Want to update bucket configuration? You need a different script. Want to know what buckets were created by scripts vs. console? There's no tracking. Managing complex configurations (like replication rules with IAM roles, lifecycle policies, and encryption keys) becomes unwieldy as JSON/YAML strings in shell scripts.

**Verdict:** A step up from console, useful for simple automation or one-time migrations. But for production infrastructure that evolves over time, you need declarative infrastructure-as-code.

### Level 2: Configuration Management Tools

**What it is:** Using Ansible, Chef, or Puppet modules to declare S3 bucket state. For example, Ansible's `amazon.aws.s3_bucket` module lets you define desired bucket properties in playbooks.

**Why it emerged:** Organizations using config management for servers extended these tools to cloud resources. It provided idempotency (run the playbook repeatedly, it ensures the bucket exists with correct settings) and fitted existing operational workflows.

**What it solves:** Better idempotency than raw scripts. Ansible won't fail if a bucket already exists; it verifies and updates settings. You get the benefits of declarative infrastructure ("I want this bucket to exist with these properties") rather than imperative ("run these commands").

**Limitations:** Configuration management tools were designed for servers, not cloud APIs. S3 feature coverage can lag—new AWS capabilities might not be in modules for months. Deletion is often problematic (bucket must be empty, or you use dangerous "force" flags). These tools shine when integrating bucket provisioning with server setup but are less ideal as the primary infrastructure layer.

**Current usage:** Declining in favor of dedicated IaC tools, but still used in environments heavily invested in Ansible/Chef/Puppet for other reasons.

**Verdict:** Workable if you're already using these tools and need basic S3 integration. For greenfield projects or advanced S3 features, dedicated IaC frameworks are superior.

## Level 3: Production-Grade Infrastructure-as-Code

This is where S3 deployment matures into reliable, auditable, production infrastructure. Four major tools dominate this space, each with trade-offs worth understanding.

### The IaC Landscape

All four tools—Terraform, Pulumi, CloudFormation, and AWS CDK—can create production-grade S3 buckets with comprehensive configuration: versioning, encryption, lifecycle rules, replication, logging, and fine-grained access control. The differences lie in philosophy, ecosystem, and operational characteristics.

#### Terraform (and OpenTofu)

**Language:** HCL (HashiCorp Configuration Language), a declarative DSL designed for infrastructure.

**Philosophy:** Multi-cloud portability. One tool to manage AWS, GCP, Azure, and hundreds of other providers.

**Strengths:**
- **Battle-tested at scale:** Terraform has managed S3 buckets in production for years across thousands of organizations. The AWS provider is mature, well-documented, and rapidly updated.
- **Rich module ecosystem:** Community modules like `terraform-aws-s3-bucket` encapsulate best practices (automatic Block Public Access for private buckets, smart defaults for encryption, etc.).
- **Plan/apply workflow:** The `terraform plan` step shows exactly what will change before you apply it, catching mistakes and preventing surprises.
- **State management flexibility:** Store state in S3 with DynamoDB locking for teams, or use Terraform Cloud for managed state and collaboration features.
- **Multi-cloud consistency:** If you manage resources across AWS, GCP, and Azure, Terraform provides one workflow, one language, and one state mechanism.

**Trade-offs:**
- **State file responsibility:** You must manage state storage, locking, and backups. Losing state can require manual reconciliation.
- **HCL learning curve:** While simpler than general programming languages, HCL has its own syntax and patterns to learn.
- **Drift detection:** Open-source Terraform detects drift by running `terraform plan` (compares desired vs. actual state). It's effective but manual unless you set up automation. Terraform Cloud offers automatic drift detection as a paid feature.

**When to choose Terraform:**
- You have or anticipate multi-cloud infrastructure
- You want a huge ecosystem of community modules and examples
- Your team prefers declarative DSLs over programming languages
- You're comfortable managing state files (or using Terraform Cloud)

**OpenTofu note:** OpenTofu is Terraform's open-source fork (after HashiCorp's license change). It's functionally identical to Terraform for S3 use cases, with the same AWS provider and compatible modules. Choose OpenTofu if you prefer a truly open-source governance model.

#### Pulumi

**Language:** TypeScript, Python, Go, C#, Java—real programming languages, not a DSL.

**Philosophy:** Infrastructure is software. Use your IDE, type checking, testing frameworks, and language constructs (loops, conditionals, functions) to define infrastructure.

**Strengths:**
- **Developer-friendly:** If your team writes application code in TypeScript or Python, they can write infrastructure in the same language with familiar tools and patterns.
- **Type safety and IDE support:** Full autocomplete, compile-time type checking, and refactoring support from modern IDEs.
- **Multi-cloud like Terraform:** Supports AWS, GCP, Azure, Kubernetes, and more through a unified programming model.
- **Can consume Terraform providers:** Pulumi's provider ecosystem includes bridged Terraform providers, keeping it up-to-date with new AWS features.
- **Managed state by default:** Pulumi offers a free managed state backend (Pulumi Service), avoiding the state file management burden. Self-hosted backends are also supported.

**Trade-offs:**
- **Smaller community than Terraform:** Fewer third-party modules and examples, though the gap is closing.
- **Requires Pulumi CLI and account:** Even for self-managed state, you need the Pulumi CLI. The default workflow involves the Pulumi Service, which some organizations prefer to avoid.
- **More abstraction:** Programming language flexibility is powerful but can lead to complex abstractions. Terraform's HCL constraints can actually be helpful in keeping infrastructure definitions simple and consistent.

**When to choose Pulumi:**
- Your team has strong software development skills and wants to leverage them for infrastructure
- You value IDE support, type safety, and testability
- You prefer managed state without additional setup
- You want to use programming patterns (loops, conditionals, shared libraries) extensively in infrastructure code

#### AWS CloudFormation

**Language:** JSON or YAML templates, purely declarative.

**Philosophy:** AWS-native, first-party infrastructure-as-code. Tightly integrated with AWS services and console.

**Strengths:**
- **AWS-managed everything:** No state files to manage—AWS stores stack state and handles all deployment orchestration.
- **First-party support:** New AWS features often appear in CloudFormation before third-party tools update.
- **Built-in drift detection:** CloudFormation can detect and report configuration drift between your template and actual resources without additional tools.
- **Rollback on failure:** If a stack deployment fails, CloudFormation automatically rolls back changes, protecting against partial deployments.
- **No additional tools:** Works entirely within AWS—no external CLI to install or manage.

**Trade-offs:**
- **AWS-only:** Cannot manage resources outside AWS. If you have GCP or Azure infrastructure, you need separate tools.
- **Verbose templates:** JSON/YAML templates for complex buckets (with replication, lifecycle rules, policies) can become large and repetitive.
- **Limited abstractions:** No loops, no conditionals beyond basic intrinsic functions. Reusing configuration across stacks requires template nesting or copy-paste.
- **Slower iteration:** Large stacks can take minutes to update, even for small changes, due to CloudFormation's orchestration overhead.

**When to choose CloudFormation:**
- You're exclusively in AWS with no multi-cloud plans
- You want AWS-managed state and deployment with no external dependencies
- You value built-in drift detection and rollback guarantees
- You prefer AWS-native tooling and integration

#### AWS CDK (Cloud Development Kit)

**Language:** TypeScript, Python, Java, C#, Go—programming languages that synthesize to CloudFormation.

**Philosophy:** CloudFormation's power with a developer-friendly programming interface. Write infrastructure in code, deploy via CloudFormation.

**Strengths:**
- **Programming language benefits:** Type safety, IDE support, code reuse, testing frameworks—all the advantages of Pulumi but deploying to CloudFormation.
- **High-level constructs:** CDK provides "L2" and "L3" constructs that abstract common patterns. An L3 construct for an S3 bucket might automatically enable encryption, Block Public Access, and logging based on best practices, reducing boilerplate.
- **CloudFormation advantages:** Inherits CloudFormation's managed state, drift detection, and rollback capabilities.
- **AWS-native ecosystem:** First-party support from AWS, strong integration with AWS services.

**Trade-offs:**
- **AWS-only:** CDK synthesizes to CloudFormation, so it's inherently AWS-specific. No multi-cloud support.
- **Inherits CloudFormation limitations:** Slow updates for large stacks, limitations on in-place updates (some changes require resource replacement), and CloudFormation's occasional quirks (like non-empty bucket deletion failures).
- **More complex toolchain:** Requires Node.js and CDK CLI. The synthesis step (code → CloudFormation template → deployment) adds a layer of indirection.
- **Learning curve:** You need to understand both CDK's abstractions and CloudFormation's behavior. Debugging issues sometimes requires inspecting the synthesized CloudFormation template.

**When to choose CDK:**
- You're deeply invested in AWS with no multi-cloud needs
- Your team prefers writing infrastructure in programming languages
- You want to leverage high-level constructs that encapsulate AWS best practices
- You value CloudFormation's managed deployment but want better developer experience

### Kubernetes-Based Operators (Advanced Use Cases)

For organizations running Kubernetes-centric infrastructure, managing S3 buckets as custom resources in Kubernetes is possible via **Crossplane** or **AWS Controllers for Kubernetes (ACK)**.

**How it works:** You define an S3 bucket in a Kubernetes YAML manifest. The operator watches these custom resources and calls AWS APIs to create/update buckets, keeping them synchronized.

**Benefits:**
- **Unified control plane:** If you already manage application deployments, databases, and networking via Kubernetes, adding S3 buckets to the same API/GitOps workflow creates consistency.
- **Reconciliation:** Kubernetes operators continuously reconcile desired state, automatically correcting drift (if someone manually modifies the bucket, the operator will revert it).
- **Multi-cloud abstraction (Crossplane):** Crossplane can manage AWS, GCP, and Azure resources through a unified Kubernetes API.

**Trade-offs:**
- **Operational complexity:** You must run and maintain the operator control plane. This adds moving parts (operator pods, CRDs, RBAC) compared to running Terraform or CloudFormation.
- **Feature lag:** Operators may not support all S3 configurations immediately. Advanced features like Object Lock or fine-grained replication rules might take time to be exposed.
- **Overhead for non-Kubernetes workloads:** If your application doesn't run in Kubernetes, managing S3 via a Kubernetes operator adds unnecessary complexity.

**When to consider operators:**
- You have a strong Kubernetes-native operational model
- You already use GitOps (ArgoCD, Flux) for infrastructure
- You want consistent reconciliation across all infrastructure
- You're willing to invest in operator tooling

**Planton perspective:** Most teams are better served by dedicated IaC tools (Terraform, Pulumi, CloudFormation, CDK). Kubernetes operators for cloud resources are niche and best reserved for environments already heavily Kubernetes-centric.

## Production Essentials: What Every S3 Bucket Needs

Regardless of deployment method, production S3 buckets require specific configurations. Understanding these essentials ensures your buckets are secure, cost-effective, and resilient.

### Security: The Non-Negotiables

**Block Public Access:** The single most important S3 security setting. AWS now enables Block Public Access by default for new buckets, but older buckets may not have it. This four-part setting (block public ACLs, ignore public ACLs, block public bucket policies, restrict public buckets) acts as a safety net, preventing accidental public exposure even if someone adds a permissive policy or ACL.

**Rule:** Enable all four Block Public Access settings on every bucket unless you have an explicit, documented reason for public content. Even then, consider using CloudFront with Origin Access Control instead of making the bucket itself public.

**Encryption at rest:** AWS automatically encrypts all new S3 objects with SSE-S3 (AES-256) as of 2023. This means "doing nothing" gets you encryption. However, many organizations require **SSE-KMS** with customer-managed keys for:
- Audit trails (CloudTrail logs every KMS key usage)
- Access control (you can revoke key access to effectively revoke data access)
- Compliance requirements (regulations may mandate customer-controlled encryption keys)

**Rule:** Accept the default SSE-S3 for non-sensitive data. Use SSE-KMS with customer-managed keys for any sensitive or regulated data, ensuring your IAM principals have KMS decrypt permissions.

**IAM and bucket policies:** Follow the principle of least privilege. Grant only the minimum required permissions:
- Application IAM roles should have `s3:GetObject` and `s3:PutObject` only on specific prefixes (e.g., `arn:aws:s3:::my-bucket/app-data/*`)
- Never use wildcard principals (`"Principal": "*"`) in bucket policies unless the bucket is genuinely public
- Use conditions (like `aws:SourceVpce` for VPC endpoint restrictions) to further limit access

**Object Ownership:** AWS defaults new buckets to "Bucket owner enforced," which disables ACLs entirely. This is correct—ACLs are a legacy access control mechanism that complicates security. Use IAM and bucket policies for all access control.

**Encryption in transit:** All AWS SDKs use HTTPS by default. Enforce this in bucket policies with a condition requiring `aws:SecureTransport: true` if you need policy-level guarantees.

### Versioning: Insurance Against Mistakes

**Why enable it:** Versioning protects against accidental deletions and overwrites. When a file is deleted or updated, S3 keeps the previous version rather than permanently discarding it. This has saved countless production systems from "oops, I deleted the wrong file" disasters.

**The cost trade-off:** Every version is a separate object that incurs storage costs. If you update a 10 GB file daily and keep versions forever, you'll accumulate 3.6 TB per year. This is where lifecycle policies become critical (see below).

**Best practice:** Enable versioning on all production buckets containing critical or irreplaceable data. Use lifecycle rules to expire old versions after a retention period (30 days, 90 days, etc.) to control costs. For ephemeral data (like temp files or easily-regenerated content), versioning may be unnecessary.

**Note:** Once enabled, versioning can only be *suspended*, not disabled. Previous versions remain. Enable it from the start for buckets that might need it.

### Lifecycle Policies: Automatic Cost Optimization

Lifecycle policies automate storage class transitions and expiration, preventing runaway costs and manual housekeeping.

**Common patterns:**

**Transition to cheaper storage:**
- Move to **Standard-IA** (Infrequent Access) after 30 days for data accessed less than once per month
- Move to **Glacier Instant Retrieval** after 90 days for archive data that still needs millisecond access
- Move to **Glacier Flexible Retrieval** after 180 days for true archives (retrieval takes minutes/hours)
- Move to **Glacier Deep Archive** after 1 year for long-term retention (retrieval takes 12+ hours)

**Expiration:**
- Delete log files after retention period (90 days, 1 year, etc.)
- Delete old versions after 30-90 days (if versioning is enabled)
- Abort incomplete multipart uploads after 7 days (clean up failed uploads that would otherwise incur storage charges)

**Intelligent-Tiering alternative:** If access patterns are unpredictable, use the **Intelligent-Tiering** storage class. S3 automatically moves objects between tiers based on access patterns. There's a small per-object monitoring fee (\$0.0025 per 1,000 objects), but for large objects with variable access, the automatic optimization often pays for itself.

**Example:** A logging bucket might have:
- Logs transition to Standard-IA after 30 days
- Logs transition to Glacier Flexible Retrieval after 90 days
- Logs deleted after 1 year (or moved to Deep Archive for long-term compliance)
- Old versions deleted after 30 days
- Incomplete multipart uploads aborted after 7 days

This lifecycle policy ensures cost-effective storage without manual intervention.

### Replication: Disaster Recovery and Compliance

**Cross-Region Replication (CRR):** Automatically copies objects to a bucket in a different AWS region. Essential for:
- **Disaster recovery:** If a region fails, data remains available in another region
- **Compliance:** Regulations may require geographically separated copies
- **Latency optimization:** Serve global users from buckets closer to them

**Same-Region Replication (SRR):** Copies objects to another bucket in the same region. Useful for:
- **Cross-account backups:** Replicate production data to a separate AWS account, protecting against credential compromise in the primary account
- **Aggregating logs:** Collect logs from multiple buckets into a single central bucket
- **Compliance without cross-region costs:** Some regulations require multiple copies but don't mandate geographic separation

**Requirements:** Replication requires versioning on both source and destination buckets, plus an IAM role that S3 uses to perform the replication.

**Gotcha:** Replication is not retroactive by default—only new objects after enabling replication are copied. AWS offers Batch Replication to copy existing objects on-demand.

**Cost consideration:** CRR incurs cross-region data transfer charges (often the largest replication cost) plus storage costs in the destination region. Use filters to replicate only necessary prefixes, and consider replicating to a cheaper storage class (like Standard-IA) if replicas are rarely accessed.

### Logging and Monitoring: Visibility Matters

**Server Access Logging:** S3 can log every request (GET, PUT, DELETE, LIST, etc.) to a separate bucket. Each log entry includes requester IP, timestamp, request type, object key, HTTP status, and bytes transferred. Essential for:
- Security audits and forensics
- Identifying unusual access patterns
- Debugging client issues

**Trade-off:** Access logs can be voluminous for busy buckets and are delivered with some delay (minutes to hours).

**Alternative: CloudTrail Data Events:** CloudTrail can log S3 API calls at the object level with near real-time delivery to CloudWatch Logs or S3. This integrates with your central CloudTrail for unified auditing but costs more for high-traffic buckets.

**Best practice:** Enable at least one form of logging (server access logs to a central logging bucket) for production buckets. Use CloudTrail for critical buckets where you need real-time alerts on access.

**S3 Storage Lens:** Provides dashboards and metrics about storage usage, access patterns, and cost efficiency. It can identify:
- Buckets with excessive old versions
- Buckets without lifecycle rules
- Incomplete multipart uploads consuming storage

Use Storage Lens to continuously optimize your S3 infrastructure.

**CloudWatch Metrics and Alarms:** S3 provides basic metrics (storage size, object count) for free. Enable request metrics (charges apply) for busy buckets to track GET/PUT request rates. Set alarms for anomalies (e.g., sudden spike in PUT requests might indicate a misbehaving client or potential abuse).

## Common Anti-Patterns to Avoid

**Public buckets without justification:** The vast majority of S3 buckets should be private. If you need to serve public content (like a static website), use CloudFront with Origin Access Control so the bucket itself remains private and content is served over HTTPS with caching.

**Wildcard bucket policies:** `"Principal": "*"` or overly broad actions (`s3:*`) in bucket policies create security risks. Use fine-grained permissions with conditions.

**Versioning without lifecycle management:** Enabling versioning and forgetting to expire old versions leads to spiraling storage costs. Always pair versioning with lifecycle rules.

**Ignoring encryption:** While AWS now encrypts by default, earlier it didn't. Audit older buckets to ensure encryption is enabled. For sensitive data, explicitly require SSE-KMS.

**No replication for critical data:** If a bucket holds irreplaceable data and there's no backup or replication, you're one accidental deletion or AWS issue away from disaster. S3 is highly durable within a region (11 nines), but that doesn't protect against user error or account compromise.

**Too many buckets:** AWS has a default limit of 1,000 buckets per account (can be increased via support request, but there are operational limits). Many use cases are better served by organizing data with prefixes in fewer buckets rather than proliferating buckets.

**Millions of tiny files without aggregation:** Storing millions of sub-1KB files incurs disproportionate metadata and request costs. Consider combining them into archives or using a different storage strategy. Storage classes like Standard-IA and Intelligent-Tiering charge minimum 128 KB per object, making tiny files expensive.

**Ignoring incomplete multipart uploads:** Failed multipart uploads leave orphaned chunks that consume storage. Always enable the lifecycle rule to abort incomplete uploads after 7 days.

## The Planton Approach

Planton's S3 bucket API abstracts deployment complexity while providing production-grade defaults and flexibility.

### Design Philosophy

**Secure by default:** Every bucket created via Planton has Block Public Access enabled, encryption configured (SSE-S3 minimum, SSE-KMS available), and follows least-privilege IAM patterns unless explicitly overridden.

**80/20 configuration:** The API exposes the 20% of S3 settings that 80% of use cases need:
- Bucket name, region, and basic metadata
- Public/private flag (`isPublic: false` by default)
- Versioning toggle
- Encryption settings (default SSE-S3, optional SSE-KMS with key ID)
- Tags for governance and cost allocation

**Optional advanced features:** For power users, the API supports:
- Lifecycle rules (transitions and expiration)
- Replication (CRR/SRR configuration)
- Server access logging
- CORS configuration
- Bucket policies

This layered approach means simple use cases (a private bucket for application data) require minimal configuration, while complex scenarios (compliance bucket with replication, object lock, and lifecycle policies) are fully supported.

### Multi-Cloud Consistency

Planton uses Terraform (OpenTofu) as the underlying IaC engine for AWS resources. This choice provides:
- **Multi-cloud portability:** The same workflow and tooling that manage AWS S3 also manage GCP Cloud Storage and Azure Blob Storage, ensuring operational consistency
- **Battle-tested reliability:** Terraform's AWS provider is mature, widely used, and rapidly updated
- **Rich ecosystem:** Access to community modules and best practices

### From API to Infrastructure

When you define an S3 bucket via Planton's protobuf API:

1. **Validation:** The API validates configuration (bucket name format, region validity, compatible settings like encryption + replication)
2. **Defaults:** Secure defaults are applied (Block Public Access, encryption, etc.)
3. **Terraform generation:** The API spec translates to Terraform HCL
4. **Plan and apply:** Terraform generates a plan, showing exactly what will be created or changed
5. **State tracking:** Terraform state captures the bucket's configuration, enabling drift detection and safe updates
6. **Stack outputs:** After creation, relevant information (bucket ARN, region, encryption details) is exposed via stack outputs for integration with applications

This abstraction means developers declare *what* they need (a private S3 bucket with versioning) without needing to understand Terraform syntax, manage state files, or navigate AWS API nuances.

## Conclusion

The evolution of S3 deployment mirrors the broader shift from manual infrastructure management to infrastructure-as-code. What began as "just click create bucket" became a minefield of security misconfigurations, cost overruns, and operational fragility.

Modern infrastructure-as-code—whether Terraform, Pulumi, CloudFormation, or CDK—treats buckets as managed, versioned, auditable infrastructure. This isn't over-engineering; it's recognizing that storage underpins critical systems and deserves the same rigor as application code.

Planton builds on this foundation, providing a high-level API that encapsulates years of S3 deployment experience: secure defaults, production-grade features, and the flexibility to handle complex scenarios without sacrificing simplicity for common cases.

The result: developers declare storage requirements in a consistent, portable API, while Planton handles the translation to battle-tested IaC that deploys secure, cost-optimized, resilient S3 infrastructure.

## Further Reading

For deeper dives into specific topics:

- **S3 Security Best Practices Guide** (future: detailed guide to bucket policies, IAM roles, encryption strategies, and VPC endpoints)
- **S3 Cost Optimization Deep Dive** (future: comprehensive analysis of storage classes, lifecycle strategies, and cost monitoring)
- **S3 Replication Patterns** (future: CRR vs. SRR decision trees, cross-account replication, and Replication Time Control)
- **S3 Lifecycle Policy Cookbook** (future: common lifecycle patterns for logs, backups, data lakes, and application data)

---

**Questions or feedback?** Contribute to [Planton on GitHub](https://github.com/plantonhq/planton) or reach out via the community channels.

