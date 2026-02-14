---
title: "Spaces Deployment: From Clicks to Code"
description: "Spaces Deployment: From Clicks to Code deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanbucket"
---

# DigitalOcean Spaces Deployment: From Clicks to Code

## Introduction

Object storage has become ubiquitous infrastructure—every application stores images, backups, logs, or user uploads somewhere. DigitalOcean Spaces offers a compelling proposition: **S3-compatible object storage with predictable pricing and a built-in CDN**. Where AWS S3 pioneered the API and ecosystem, Spaces brings that same interface to DigitalOcean's simplified cloud platform.

The S3 API compatibility is Spaces' superpower. Any tool written for AWS S3—from the AWS CLI to Terraform to SDKs in every language—works with Spaces by simply pointing at a different endpoint (`nyc3.digitaloceanspaces.com` instead of `s3.amazonaws.com`). This means the deployment tooling ecosystem is mature from day one.

But S3 compatibility alone doesn't answer the fundamental question: **how should you provision and manage Spaces buckets in production?** The spectrum runs from clicking through the DigitalOcean control panel to declarative Infrastructure-as-Code that treats buckets as versioned, reviewable configuration.

This document examines that spectrum—from manual configuration through CLI tools to full IaC automation—and explains why OpenMCF chose **Pulumi with the DigitalOcean provider** as the foundation for production Spaces deployments.

## The Deployment Maturity Spectrum

### Level 0: Manual Configuration (The Anti-Pattern)

The DigitalOcean control panel makes creating a Space visually straightforward: click "Spaces," choose a region, pick a name, toggle public/private, enable CDN. For a quick experiment or one-off bucket, it works.

**The problems emerge at scale:**

1. **No Audit Trail**: Who created this bucket? When? Why is it public? The control panel doesn't capture the reasoning behind configuration choices.

2. **Configuration Drift**: You enable CORS in the UI, a colleague tweaks lifecycle rules, someone turns on versioning. Within weeks, no one knows the authoritative state.

3. **No Repeatability**: Provisioning staging and production environments means clicking through the same workflow multiple times—and inevitably making subtle mistakes (wrong region, forgot to enable CDN, incorrect CORS rules).

4. **Credential Management**: Creating Spaces access keys in the UI and copy-pasting them into application config files is a security incident waiting to happen.

The control panel is perfect for learning and exploration. It's an anti-pattern for production infrastructure that needs to be auditable, repeatable, and recoverable.

**Verdict**: Fine for learning, unacceptable for production.

---

### Level 1: CLI Tools (Better, Still Manual)

DigitalOcean's S3 compatibility means a rich ecosystem of CLI tools work out of the box:

- **AWS CLI**: The standard `aws s3` and `aws s3api` commands work by setting `--endpoint-url https://nyc3.digitaloceanspaces.com`. You can create buckets, upload files, set CORS rules—all from the command line.

- **s3cmd**: The open-source tool configures easily for Spaces by setting `host_base = nyc3.digitaloceanspaces.com` in `~/.s3cfg`.

- **rclone**: Popular for sync and backup workflows, rclone treats Spaces as just another S3 endpoint.

**What this enables:**

- **Scripting**: You can write bash scripts to provision buckets consistently.
- **Automation**: CI/CD pipelines can upload static assets to Spaces as part of deployment.
- **Migration**: Moving data between S3 and Spaces becomes a one-liner sync command.

**What it doesn't solve:**

- **State Management**: Scripts don't track "desired state" vs "actual state." If someone manually changes a bucket setting, your script won't detect or correct it.
- **Dependency Ordering**: Scripts handle simple workflows but struggle with complex dependencies (create bucket, then CDN endpoint, then update DNS).
- **Version Control**: Scripts can be versioned, but they're imperative (step-by-step instructions) rather than declarative (describing the end state).

CLI tools represent progress—automation beats clicking—but they're still fundamentally manual workflows codified in scripts.

**Verdict**: Useful for data operations and migrations, inadequate for infrastructure lifecycle management.

---

### Level 2: Configuration Management (Ansible, Chef, Puppet)

Configuration management tools like Ansible can provision Spaces using either DigitalOcean-specific modules or generic S3 modules pointed at the Spaces endpoint.

**Ansible Example**:
- The `community.digitalocean.digital_ocean_spaces` module can create/delete buckets via the DigitalOcean API.
- Generic `amazon.aws.s3_bucket` modules work by specifying DigitalOcean as the cloud provider.

**What this provides:**

- **Idempotency**: Ansible won't create duplicate buckets if they already exist.
- **Role-based Organization**: You can structure Spaces provisioning as reusable Ansible roles.
- **Integration**: If you're already using Ansible for server configuration, adding Spaces feels natural.

**Limitations:**

- **State Tracking**: Ansible is better at convergence than Terraform, but still doesn't maintain explicit state files showing what it manages.
- **Resource Relationships**: Complex scenarios (bucket → CDN endpoint → DNS record) require careful ordering in playbooks.
- **Adoption Trend**: The industry has largely standardized on dedicated IaC tools (Terraform, Pulumi) for cloud resource provisioning, relegating configuration management to server/application configuration.

Configuration management tools are production-capable but represent a previous generation's approach to infrastructure automation.

**Verdict**: Production-ready but superseded by modern IaC for cloud resources.

---

### Level 3: Infrastructure as Code (The Production Solution)

This is where we arrive at **declarative, stateful, version-controlled infrastructure**. Tools like Terraform, Pulumi, and OpenTofu treat Spaces buckets as code—you describe what you want (a bucket named X in region Y with versioning enabled), and the tool makes it so.

**The Paradigm Shift**:

- **Declarative**: You specify desired state, not imperative steps.
- **Stateful**: Tools track what they've created and detect drift.
- **Version Controlled**: Infrastructure definitions live in Git alongside application code.
- **Reviewable**: Changes go through pull requests, code review, and CI checks.
- **Repeatable**: The same code deploys identical infrastructure in dev, staging, and production.

The three major IaC tools for DigitalOcean Spaces are Terraform (with the official DigitalOcean provider), Pulumi (with the DigitalOcean provider), and OpenTofu (using the same provider as Terraform).

---

## Infrastructure-as-Code Comparison

All three tools use the **official DigitalOcean provider**, which offers native support for Spaces resources. Let's compare them for production Spaces deployments:

### Terraform (Official DigitalOcean Provider)

**Maturity**: GA, widely adopted. The DigitalOcean provider is official and well-maintained.

**Spaces Support**:
- `digitalocean_spaces_bucket` resource covers bucket creation, ACL, versioning, CORS, and lifecycle rules.
- Separate resources for CORS configuration, lifecycle policies, and CDN endpoints.
- Full support for all Spaces features: versioning (note: once enabled, cannot be disabled), lifecycle expiration, CDN integration.

**Credentials**: Best practice is environment variables (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for Spaces keys when using Spaces as a Terraform backend). Never hard-code credentials in `.tf` files.

**State Management**: 
- Terraform workspaces support multiple environments (dev/staging/prod).
- Spaces itself can serve as a Terraform backend (storing state in a Space by configuring the S3 backend with the Spaces endpoint).
- Remote state enables team collaboration and locks to prevent concurrent modifications.

**Strengths**:
- Largest ecosystem and community for Terraform in general.
- Extensive documentation and examples for DigitalOcean resources.
- HCL syntax is purpose-built for infrastructure configuration.

**Considerations**:
- HCL is a domain-specific language (DSL)—you can't leverage general programming logic easily.
- State file management requires discipline (backend configuration, locking, encryption).

---

### Pulumi (DigitalOcean Provider)

**Maturity**: Stable, production-ready. The DigitalOcean provider (v4.x) is actively maintained.

**Spaces Support**:
- `digitalocean.SpacesBucket` resource with equivalent fields to Terraform (name, region, ACL, CORS, lifecycle, versioning).
- `digitalocean.Cdn` resource for CDN endpoints.
- Full feature parity with Terraform for Spaces.

**Credentials**: Configure via Pulumi config (`digitalocean:token` for API, `digitalocean:spacesAccessId` and `digitalocean:spacesSecretKey` for Spaces). Supports secret encryption in Pulumi state.

**State Management**:
- Pulumi uses "stacks" for environments (dev, staging, prod)—more explicit than Terraform workspaces.
- State stored in Pulumi Cloud by default (free tier available) or configurable S3-compatible backends (including Spaces).
- Stack configuration can store environment-specific values (e.g., different bucket names per environment).

**Strengths**:
- **Use Real Programming Languages**: Write infrastructure in TypeScript, Python, Go, C#, or Java. This enables proper abstractions, functions, loops, and type safety.
- **IDE Support**: Full autocomplete, type checking, and refactoring support in VS Code or JetBrains IDEs.
- **Testability**: Unit test infrastructure code using standard testing frameworks.
- **Component Reusability**: Create reusable components that encapsulate best practices (e.g., a "production Spaces bucket" component that always enables versioning, lifecycle policies, and CDN).

**Considerations**:
- Smaller community than Terraform (though growing rapidly).
- Requires familiarity with a programming language, not just a DSL.

---

### OpenTofu (Terraform-Compatible)

**Maturity**: OpenTofu is the open-source fork of Terraform (post-license change). It's compatible with Terraform providers and configurations.

**Spaces Support**: Identical to Terraform—uses the same DigitalOcean provider.

**Strengths**:
- **True Open Source**: Licensed under the Mozilla Public License 2.0 (Terraform switched to Business Source License).
- **Provider Compatibility**: Terraform providers work seamlessly.
- **Migration Path**: Easy migration from Terraform.

**Considerations**:
- Newer project (forked in 2023), though built on mature Terraform codebase.
- Ecosystem still consolidating around OpenTofu vs Terraform.

---

## Feature Comparison Table

| Feature | Terraform | Pulumi | OpenTofu |
|---------|-----------|--------|----------|
| **Spaces Bucket Resource** | `digitalocean_spaces_bucket` | `digitalocean.SpacesBucket` | `digitalocean_spaces_bucket` (same as TF) |
| **CORS Configuration** | `cors_rule` block | `corsRules` field | `cors_rule` block |
| **Lifecycle Rules** | `lifecycle_rule` block | `lifecycleRules` field | `lifecycle_rule` block |
| **Versioning** | `versioning` block (cannot disable once enabled) | `versioning` args | `versioning` block |
| **CDN Endpoint** | `digitalocean_cdn` resource | `digitalocean.Cdn` resource | `digitalocean_cdn` resource |
| **State Management** | Workspaces, remote backends | Stacks, Pulumi Cloud or S3 | Workspaces, remote backends |
| **Language** | HCL (DSL) | TypeScript, Python, Go, C#, Java | HCL (DSL) |
| **Production Readiness** | ✅ Mature, widely adopted | ✅ Stable, growing adoption | ✅ Terraform-compatible, true OSS |
| **Type Safety** | Limited (HCL validation) | ✅ Full type checking | Limited (HCL validation) |
| **Testing** | Terratest (external) | Native unit testing | Terratest (external) |
| **Reusability** | Modules | Components (programmatic) | Modules |

---

## The OpenMCF Choice: Pulumi

OpenMCF uses **Pulumi with the DigitalOcean provider** as the foundation for DigitalOcean Spaces deployments. Here's why:

### 1. Programmatic Infrastructure

Pulumi lets us write infrastructure in real programming languages (Go, TypeScript, Python). This enables:

- **Abstraction and Composition**: We create reusable components that encapsulate production best practices. Instead of duplicating HCL for every bucket, we define a `ProductionSpacesBucket` component that always includes versioning, lifecycle policies, and CDN configuration.

- **Type Safety**: Pulumi's SDKs provide full type checking. Misconfigured resources are caught at development time, not runtime.

- **Logic and Validation**: Need to derive bucket names from environment and region? Validate that public buckets have appropriate CORS rules? Use standard programming constructs (functions, conditionals, loops) rather than wrestling with HCL's limitations.

### 2. Open Source Alignment

OpenMCF is an open-source project. Pulumi is open source (Apache 2.0), and while Terraform recently shifted to a Business Source License, Pulumi's commitment to open source aligns with our values. OpenTofu is also a strong open-source option, but Pulumi's programmatic approach offers unique advantages.

### 3. Multi-Language Support

Different teams prefer different languages. Pulumi supports TypeScript, Python, Go, C#, and Java—teams can use OpenMCF's infrastructure components in their preferred language.

### 4. Testing and Validation

Infrastructure should be tested like application code. Pulumi enables:

- **Unit Tests**: Test that bucket configurations meet policies (e.g., production buckets must have versioning enabled).
- **Integration Tests**: Spin up ephemeral test environments, run validations, tear down.
- **Policy as Code**: Pulumi CrossGuard policies enforce organizational standards (e.g., all public buckets must have lifecycle rules).

### 5. The 80/20 API Design

The `DigitalOceanBucketSpec` protobuf focuses on the 20% of configuration that 80% of users need:

- **`bucket_name`**: DNS-compatible name (3–63 characters).
- **`region`**: Datacenter location (e.g., `nyc3`, `sfo2`, `ams3`).
- **`access_control`**: Private (default) or public-read.
- **`versioning_enabled`**: Boolean flag (false by default).
- **`tags`**: Metadata for organization and cost tracking.

**What's intentionally omitted**: Complex bucket policies, granular ACLs for individual objects, advanced lifecycle rules with multiple conditions. These are available through the underlying Pulumi Spaces resource but aren't surfaced in the simplified spec. The 80/20 principle keeps the API approachable for common use cases while allowing power users to extend with custom Pulumi code.

### 6. Production Configuration Example

Here's what a production Spaces deployment looks like:

**Dev Environment**:
```yaml
bucketName: dev-assets
region: nyc3
access_control: PRIVATE
versioningEnabled: false
tags: ["environment:dev", "project:web-app"]
```

**Production Environment**:
```yaml
bucketName: prod-media
region: sfo2
access_control: PUBLIC_READ
versioningEnabled: true
tags: ["environment:prod", "project:web-app", "cdn-enabled:true"]
```

Behind the scenes, the Pulumi module:
- Creates the bucket with specified configuration.
- Sets up lifecycle rules (e.g., expire old logs after 30 days).
- Enables the Spaces CDN for public buckets.
- Configures CORS rules if needed.
- Exports bucket endpoints (origin URL, CDN URL) as stack outputs.

---

## Production Best Practices

Regardless of which IaC tool you use, follow these practices for Spaces in production:

### Access Control
- **Default to Private**: Only make buckets public when serving static assets to the internet.
- **Use Signed URLs**: For controlled public access (time-limited links), use S3-compatible signed URLs rather than making entire buckets public.
- **Least-Privilege Keys**: Create Spaces access keys with limited scope (specific buckets only, not full account access).

### CDN Integration
- **Enable CDN for Public Content**: Spaces includes a free CDN—use it. The CDN reduces latency for global users and lowers egress costs (cached content doesn't hit the origin repeatedly).
- **Custom Domains**: Map your own domain (e.g., `cdn.example.com`) to the Spaces CDN endpoint for branding and flexibility.

### Lifecycle Policies
- **Expire Old Data**: Don't let logs, temporary files, or old versions accumulate forever. Set lifecycle rules to automatically delete objects after N days.
- **Noncurrent Version Expiration**: If versioning is enabled, remove old versions to control storage costs.

### Versioning Strategy
- **Enable for Critical Data**: Versioning protects against accidental overwrites and deletes—essential for backups and archives.
- **Understand the Limitation**: Once enabled, versioning cannot be disabled (only suspended)—plan accordingly.

### Monitoring and Logging
- **Enable Access Logs**: Spaces can automatically log all read/write requests to another bucket. This provides an audit trail and helps diagnose issues.
- **Track Storage Growth**: Monitor bucket size and transfer metrics to avoid surprise bills.

### Credential Management
- **Never Hard-Code Keys**: Inject Spaces access keys via environment variables or secret managers.
- **Rotate Regularly**: Periodically rotate access keys and update applications.
- **Audit Permissions**: Review which applications and users have Spaces access keys.

### Disaster Recovery
- **Cross-Region Backups**: Spaces doesn't natively replicate across regions. For critical data, periodically copy to another region or provider using `rclone` or AWS CLI.
- **Versioning as Recovery**: Versioning provides point-in-time recovery for accidental deletes.

---

## S3 Compatibility: What Works, What Doesn't

DigitalOcean Spaces implements the **core S3 API** but not every AWS-specific feature.

### ✅ Supported (Works Seamlessly)
- Bucket creation/deletion
- Object PUT/GET/DELETE (including multipart uploads)
- ACLs (bucket and object level)
- Versioning (enable/suspend)
- CORS configuration
- Lifecycle rules (expiration, noncurrent version deletion)
- S3-compatible tools (AWS CLI, s3cmd, rclone, SDKs)

### ❌ Not Supported
- **Storage Classes**: Spaces has a single storage class (no Glacier, Infrequent Access, Intelligent-Tiering).
- **Cross-Region Replication**: No native replication—must be done manually.
- **KMS Integration**: Encryption is AES-256 at rest (managed by DigitalOcean), but no customer-managed keys.
- **Advanced IAM**: No integration with DigitalOcean's account IAM—access is managed via Spaces-specific keys and ACLs.

### 🔧 Tooling Configuration

Any S3-compatible tool works by setting the endpoint:

**AWS CLI**:
```bash
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 ls
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 cp file.txt s3://my-bucket/
```

**s3cmd** (`~/.s3cfg`):
```ini
host_base = nyc3.digitaloceanspaces.com
host_bucket = %(bucket)s.nyc3.digitaloceanspaces.com
```

**Pulumi/Terraform**: Use `digitalocean.SpacesBucket` or `digitalocean_spaces_bucket` resources—no need to point generic AWS providers at the Spaces endpoint (though you could).

---

## Cost Optimization

DigitalOcean Spaces uses a **predictable flat-rate pricing model**:

- **$5/month**: Includes 250 GiB storage and 1 TiB egress.
- **Overages**: ~$0.02/GB storage, ~$0.01/GB egress (beyond included amounts).
- **No Per-Request Fees**: Unlike AWS S3 (which charges per 1,000 PUTs/GETs), Spaces includes unlimited requests.

**Cost-Saving Strategies**:

1. **Lifecycle Policies**: Automatically expire old files to avoid storage accumulation.
2. **CDN for Popular Content**: Cached content reduces origin egress.
3. **Compression**: Store gzipped assets to reduce storage and transfer size.
4. **Multi-Bucket Strategy**: Separate hot (frequently accessed) and cold (archival) data into different buckets with different lifecycle rules.

**Comparison with AWS S3**:
- **Spaces**: Simple, predictable—great for small-to-medium workloads.
- **AWS S3**: Variable pricing with more storage class options—better for massive-scale or complex tiering needs.

For typical web applications (static assets, user uploads, backups), Spaces' simplicity and included CDN often result in lower total cost than AWS S3 + CloudFront.

---

## Conclusion

The evolution from manual Spaces configuration to declarative Infrastructure-as-Code mirrors the broader maturation of cloud operations. Clicking through the DigitalOcean control panel teaches you what Spaces can do. CLI tools let you script repetitive tasks. But production infrastructure demands **auditability, repeatability, and version control**—qualities that only IaC provides.

OpenMCF's choice of **Pulumi** reflects a belief that infrastructure should be written in real programming languages, tested like application code, and composed from reusable components. The `DigitalOceanBucketSpec` API distills Spaces configuration down to the essential 20% that covers 80% of use cases: name, region, access control, versioning, and tags.

For teams building on DigitalOcean, Spaces offers a compelling combination: S3 compatibility (bring your existing tools), predictable pricing (no surprise bills), and a built-in CDN (fast global delivery). Managed through Pulumi, Spaces buckets become first-class infrastructure—versioned, reviewed, and deployed alongside your application code.

Whether you're hosting static sites, storing user uploads, or building a data lake, treating Spaces as code is the path to reliable, scalable, maintainable infrastructure.

