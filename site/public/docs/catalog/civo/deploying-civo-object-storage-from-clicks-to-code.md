---
title: "Deploying Civo Object Storage: From Clicks to Code"
description: "Deploying Civo Object Storage: From Clicks to Code deployment documentation"
icon: "package"
order: 100
componentName: "civobucket"
---

# Deploying Civo Object Storage: From Clicks to Code

## The S3-Compatible Cloud Storage Landscape

Object storage has become the bedrock of cloud infrastructure—storing everything from application assets and backups to machine learning datasets and static websites. Amazon S3 pioneered the space and created what became a de facto API standard. Today, S3-compatible storage is everywhere: from hyperscale clouds (Google Cloud Storage, Azure Blob) to regional providers, self-hosted solutions (MinIO, Ceph), and developer-friendly platforms like Civo.

Civo Object Storage exemplifies the "S3-compatible with fewer knobs to twist" philosophy. It provides the familiar S3 API you already know—meaning your existing tools, libraries, and workflows work unchanged—while eliminating the complexity that makes AWS pricing calculators a cottage industry. Under the hood, all data is encrypted at rest with AES-256 and served over SSL. The service is designed for teams who value straightforward pricing, regional data sovereignty, and zero data transfer fees when used within Civo's platform.

For developers building on Civo, the value proposition is clear: **collocated storage with predictable costs**. A Kubernetes cluster in Frankfurt accessing a Frankfurt object store sees low latency, high throughput, and no surprise egress charges. You pay a flat capacity rate (roughly $0.01/GB-month, billed in 500 GB increments) with no per-API-call fees and no bandwidth charges for in-platform access. This predictability stands in stark contrast to AWS S3, where storage, transfer, and request costs stack up and vary by usage patterns.

This document explores how to provision and manage Civo Object Storage buckets—from manual dashboard workflows to production-grade Infrastructure as Code (IaC) approaches—and explains why OpenMCF defaults to a Pulumi-based implementation with a simplified Protobuf API.

## The Deployment Methods Spectrum

Like most cloud resources, Civo buckets can be created through various methods. Each has its place, but they vary dramatically in suitability for production environments.

### Level 0: The Dashboard (Learning and One-Offs)

The Civo web dashboard provides a straightforward UI: select a region, choose a bucket name, allocate capacity in 500 GB increments, and attach (or create) an access credential. It's perfect for exploration, demos, or creating a single bucket for a pet project.

**The Reality Check**: Dashboard provisioning doesn't scale. You can't version-control clicks, diff changes, or replay them across environments. Worse, it's easy to introduce inconsistencies—one bucket created in London with versioning enabled, another in Frankfurt without, each with different lifecycle policies. When you have multiple environments (dev, staging, prod) or need to recreate infrastructure after an incident, manual provisioning becomes a liability.

**Common Pitfalls**: Region mismatches (creating a credential in one region, a bucket in another), lost secret keys, or hitting quota limits without realizing it. The dashboard won't stop you from making configuration mistakes that only surface later.

**Verdict**: Use the dashboard to learn how Civo Object Storage works. Then move to automation.

### Level 1: The Civo CLI (Scripts and Semi-Automation)

The `civo` CLI wraps Civo's REST API, allowing bucket creation from the terminal:

```bash
civo objectstore create mybucket --size 500 --owner-name my-credential --region FRA1
```

This is a step up from clicking. You can script bucket creation, integrate it into deployment scripts, or add it to CI/CD pipelines. The CLI outputs JSON for parsing, making it scriptable.

**The Reality Check**: While scriptable, CLI-based automation still requires you to manage state manually. Did the bucket already exist? Is it the right size? Has someone changed the credential? Shell scripts can become complex quickly when handling idempotency, error recovery, and dependency ordering. You'll reinvent state management poorly.

**Where It Shines**: Quick prototyping, one-off migrations, or emergency bucket provisioning during an incident. The CLI is also invaluable for fetching credentials or debugging (e.g., retrieving the secret key for an existing credential).

**Verdict**: Good for one-off tasks and scripts, but insufficient for production infrastructure management.

### Level 2: Infrastructure as Code (Production Standard)

This is where infrastructure management matures. Tools like Terraform and Pulumi allow you to declare your desired state—"I want a bucket named `acme-prod-backups` in London with 1 TB capacity"—and the tool handles creation, updates, and drift detection.

Both Terraform and Pulumi support Civo through official providers. The Terraform provider (`civo/civo`) is in the 1.x release series and production-ready. Pulumi's Civo provider is a bridge over the same underlying implementation, offering the same stability with a code-first approach.

#### What IaC Gives You

**State Management**: The tool tracks what exists. Run `terraform plan` or `pulumi preview` and you'll see exactly what will change before applying it. If someone manually changes a bucket's size in the dashboard, the next run will detect drift and can correct it.

**Repeatability**: The same configuration creates identical infrastructure across environments. Your dev, staging, and prod buckets differ only in parameterized values (names, regions, sizes), not in structure.

**Version Control**: Your infrastructure is code. It's reviewed in pull requests, versioned in Git, and has a clear history. You can trace who changed what and when.

**Dependency Orchestration**: IaC tools handle ordering. If a bucket depends on a credential, or if multiple resources reference each other, the tool creates them in the correct sequence.

#### The Coverage Reality

Here's where you need to understand the nuance: **Civo's Terraform and Pulumi providers excel at bucket lifecycle management but don't expose S3-level configurations directly**.

The providers give you:
- Bucket creation and deletion
- Capacity allocation (`max_size_gb`)
- Region selection
- Credential attachment (auto-create or reference existing)

They **don't** give you first-class resource attributes for:
- Enabling versioning
- Setting lifecycle rules (e.g., expire objects after 90 days)
- Configuring bucket policies or CORS
- Setting up cross-region replication

**Why?** These settings aren't managed through Civo's control plane API—they're set via the S3 API itself. This is different from AWS, where CloudFormation (and thus Terraform) can set bucket versioning or lifecycle policies through the control API.

**The Workaround**: After provisioning a bucket via Terraform/Pulumi, you can use the AWS CLI (pointed at the Civo endpoint) or the AWS SDK to configure versioning, policies, and lifecycle rules:

```bash
aws s3api put-bucket-versioning \
  --bucket mybucket \
  --versioning-configuration Status=Enabled \
  --endpoint-url https://your-civo-endpoint.civo.com
```

Or wrap these calls in a Terraform `local-exec` provisioner or Pulumi dynamic provider. It's not elegant, but it works.

**Verdict**: Terraform and Pulumi are production-ready for **bucket provisioning**. Additional S3 configurations require supplementary automation, but the core resource lifecycle is solid.

### Level 3: Abstraction Layers (OpenMCF's Approach)

OpenMCF takes IaC a step further by abstracting the underlying tool complexity. Instead of writing Terraform HCL or Pulumi code directly, you declare your infrastructure using Protobuf-defined APIs. A YAML file specifies what you want (a `CivoBucket` with certain properties), and OpenMCF's Pulumi modules translate that intent into reality.

**Why This Matters**: Most teams don't need to know how to call Civo's API or AWS S3 API endpoints—they just want a bucket that's secure, versioned, and sized correctly. By defining a higher-level API (`CivoBucketSpec`), OpenMCF captures the **80% of configuration that 80% of users need**:

- `bucket_name`: What to call it
- `region`: Where to create it
- `versioning_enabled`: Boolean for data protection
- `tags`: Metadata for organization

Behind the scenes, the Pulumi module:
1. Creates the Civo bucket via the provider
2. Configures S3-level settings (versioning, policies) via SDK calls
3. Outputs the bucket URL and credentials for consumption

You get **declarative simplicity** without sacrificing functionality. The abstraction handles the two-API dance (Civo control plane + S3 API) so you don't have to.

**Verdict**: This is the production pattern for teams managing multi-cloud infrastructure at scale. You gain consistency, reduced cognitive load, and a unified interface across all cloud resources.

## Comparing IaC Tools: Terraform vs. Pulumi

If you're implementing Civo Object Storage in IaC, here's how the two leading tools stack up:

| Aspect | Terraform | Pulumi |
|--------|-----------|--------|
| **Maturity for Civo** | Stable 1.x provider, maintained by Civo | Bridged from Terraform provider, equally stable |
| **Bucket Lifecycle** | ✅ Full support (create, update, delete, size) | ✅ Full support (same underlying API) |
| **S3 Config (versioning, etc.)** | ❌ Not in provider; requires workarounds | ❌ Not in provider; handle via code or provisioners |
| **State Management** | Remote backends (S3, Terraform Cloud, etc.) | Pulumi Service or self-hosted backend |
| **Multi-Environment** | Workspaces or separate state files | Stacks (first-class concept) |
| **Language** | HCL (domain-specific) | TypeScript, Python, Go, C#, Java (general-purpose) |
| **Secret Handling** | Sensitive values in state, external secret managers | Encrypted secrets in Pulumi config/state |
| **Community/Docs** | Larger Terraform community, more examples | Smaller but growing; official Civo docs use Terraform examples |
| **Integration with OpenMCF** | Possible but requires separate toolchain | Native (Planton uses Pulumi under the hood) |

**Recommendation**: Both are production-ready. Choose **Terraform** if you have an existing Terraform workflow and prefer HCL. Choose **Pulumi** if you want type-safe code, better secret management, or are integrating with OpenMCF (which uses Pulumi natively).

## The 80/20 Configuration Principle

Most object storage buckets need:
- A name
- A region
- A size
- Privacy controls
- Optional versioning
- Optional lifecycle rules

They **don't** need:
- Fine-grained IAM policies (Civo uses simpler credential-per-bucket model)
- Object Lock or legal hold (compliance features not supported)
- Multi-region active-active configurations (handled via replication, if needed)
- Storage tiering (Civo has one tier, not Standard/IA/Glacier)

OpenMCF's `CivoBucketSpec` reflects this reality. The API is minimal but expressive:

```protobuf
message CivoBucketSpec {
  string bucket_name = 1;       // DNS-compatible, unique name
  CivoRegion region = 2;        // LON1, NYC1, FRA1, etc.
  bool versioning_enabled = 3;  // Protect against accidental deletion
  repeated string tags = 4;     // Organizational metadata
}
```

Notably absent: `max_size_gb` (capacity). Why? Because in many scenarios, the platform can manage sizing automatically—start with 500 GB, monitor usage, and scale up when needed. Exposing size is possible, but the default abstraction prioritizes developer simplicity over every knob.

### Example Configurations

**Dev Bucket (Minimal)**:
```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: myapp-dev-storage
spec:
  bucketName: myapp-dev-storage
  region: FRA1
  versioningEnabled: false
  tags:
    - env:dev
    - team:backend
```
Private, no versioning, minimal overhead.

**Production Backups (Hardened)**:
```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: acme-prod-backups
spec:
  bucketName: acme-prod-backups
  region: LON1
  versioningEnabled: true
  tags:
    - env:prod
    - criticality:high
    - retention:180-days
```
Versioned, region-specific for compliance, tagged for governance. Behind the scenes, the Pulumi module also applies a lifecycle rule to expire backups after 180 days and remove old versions after 30 days.

**Public Assets Bucket**:
```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: marketing-static-assets
spec:
  bucketName: marketing-static-assets
  region: NYC1
  versioningEnabled: false
  tags:
    - env:prod
    - public:true
```
The `public:true` tag signals the module to apply an S3 bucket policy allowing public reads. Simple intent, handled automatically.

## S3 Compatibility: Use Your Existing Tools

One of Civo Object Storage's greatest strengths is **100% S3 API compatibility for object operations**. This means:

- **AWS CLI**: `aws s3 ls s3://mybucket --endpoint-url https://...` works perfectly
- **AWS SDKs**: Boto3 (Python), AWS SDK (JavaScript, Java, Go, .NET) all work with endpoint override
- **s3cmd, rclone, MinIO client**: Configure the endpoint and credentials, use as normal
- **Terraform S3 Backend**: Store Terraform state in a Civo bucket
- **Helm S3 Plugin**: Host Helm charts in object storage

**Key Configuration**: Use **path-style addressing** (`https://endpoint.civo.com/bucket/key`) rather than virtual-host style (`https://bucket.endpoint.civo.com/key`). Most S3 clients support a `forcePathStyle` or equivalent flag.

**Limitations**: You can't create or delete buckets via the S3 API (that requires Civo's control plane API). Some advanced S3 features (S3 Select, event notifications, Glacier tiering) don't exist on Civo. But for 95% of use cases—PUT, GET, LIST, DELETE, multipart uploads, presigned URLs—it's seamless.

## Production Best Practices

### Security
- **Never hardcode credentials**: Use environment variables, Kubernetes Secrets, or secret managers
- **Create separate credentials per service**: Limit blast radius if a key is compromised
- **Default to private**: Only make buckets public when necessary, and only expose specific prefixes
- **Rotate keys periodically**: Civo allows multiple credentials; create new, migrate, delete old

### Data Protection
- **Enable versioning for critical data**: Protects against accidental deletion or overwrites
- **Implement lifecycle policies**: Expire old versions after 30-90 days to control storage growth
- **Cross-region replication for disaster recovery**: Civo offers free transfer; replicate production buckets to a second region
- **Encrypt sensitive data client-side**: Even though Civo encrypts at rest, you control the keys with client-side encryption

### Cost Optimization
- **Right-size allocations**: Start with 500 GB, monitor usage, scale in 500 GB increments
- **Use lifecycle rules aggressively**: Auto-delete logs after 30 days, old backups after 180 days
- **Leverage free in-platform transfer**: Run processing jobs on Civo instances in the same region
- **Compress data**: Text logs, JSON, CSVs compress 5-10x, saving storage costs

### Monitoring
- Track bucket usage relative to allocated capacity (Civo API provides metrics)
- Set alerts when usage exceeds 80% to avoid hitting limits
- Monitor application errors for failed writes (could indicate full bucket)

## What OpenMCF Supports (and Why)

OpenMCF defaults to **Pulumi** for Civo bucket provisioning because:

1. **Native Integration**: Planton's infrastructure-as-data pattern uses Pulumi modules under the hood
2. **Code-First Flexibility**: Pulumi allows wrapping both the Civo provider (for bucket creation) and AWS SDK calls (for S3 configuration) in a single stack
3. **Secret Management**: Pulumi's encrypted config and outputs safely handle access keys and secrets
4. **Type Safety**: Pulumi's SDKs catch errors at development time, not runtime

The abstraction layer (Protobuf APIs translated to Pulumi) gives you:
- A **simplified, opinionated API** covering the most common 80% of use cases
- **Automatic handling** of multi-API workflows (Civo control plane + S3 API)
- **Consistency** across all cloud resources (GCP, AWS, Kubernetes, Civo) using the same pattern

You still get the full power of Pulumi (and by extension, S3 API) for edge cases, but most teams never need to drop down to that level.

## Conclusion: Simplicity Without Compromise

The evolution from dashboard clicks to Infrastructure as Code mirrors the broader maturation of cloud infrastructure management. For Civo Object Storage, the journey ends at a sweet spot: **S3-compatible storage with straightforward pricing, managed via declarative IaC, abstracted to a developer-friendly API**.

Civo's flat-rate pricing eliminates the surprise bills that plague S3 users. The S3 API compatibility means zero learning curve for existing tools. And OpenMCF's Protobuf-defined specs reduce cognitive load, letting you focus on what you're building rather than how to provision it.

Whether you're storing Kubernetes persistent volume backups, serving static website assets, or archiving application logs, Civo Object Storage—deployed through OpenMCF—gives you production-grade reliability with minimal complexity.

**Next Steps**: 
- For implementation details, see the IaC modules in `../iac/pulumi/`
- For examples, see `../hack/manifest.yaml`
- For Civo-specific pricing and regions, consult [Civo's documentation](https://www.civo.com/docs/object-stores)

