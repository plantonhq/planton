# Deploying Cloudflare R2 Buckets: Breaking Free from Egress Fees

## Introduction

For years, the conventional wisdom around cloud object storage was simple: "It's cheap to store, expensive to retrieve." AWS S3, Google Cloud Storage, and Azure Blob all follow the same playbook—modest storage costs (around $0.02–0.023 per GB-month) coupled with hefty egress fees ($0.08–0.12 per GB transferred out). For content delivery, media hosting, or any workload that serves data globally, these egress charges can dwarf storage costs. A video streaming service moving 100 TB out per month faces $9,000–12,000 in bandwidth charges alone from the big-3 clouds—before even touching compute or storage.

**Cloudflare R2** flips this model on its head with a radical promise: **zero egress fees**. Not "reduced" or "discounted"—actually zero. You pay only for storage ($0.015/GB-month for Standard, $0.010 for Infrequent Access) and API operations (~$4.50 per million writes, ~$0.36 per million reads). Data leaving R2—whether to end users, to Cloudflare Workers, or to other clouds—costs nothing. This isn't charity; it's strategy. Cloudflare leverages its global peering network (which already handles trillions of requests daily) to absorb bandwidth costs that competitors pass through to customers.

The impact? Up to **99% cost savings** for egress-heavy workloads. R2 positions itself as a multi-cloud escape hatch: store once, serve anywhere, without vendor lock-in or exit penalties. It's S3-compatible (most tools "just work"), strongly consistent (no eventual-consistency pitfalls), and designed to pair seamlessly with Cloudflare's CDN, Workers, and Pages.

But R2 isn't S3. It omits features like object versioning, bucket policies, and static website hosting. It's simpler, cheaper, and opinionated—optimized for the 80% use case of storing and serving content, not the 20% of complex AWS integrations. For teams deploying multi-cloud architectures or escaping egress fees, R2 is a strategic weapon. This guide explains how to deploy R2 buckets in production, what methods work at scale, and how OpenMCF abstracts the complexity into a clean API.

---

## The Deployment Spectrum: From Manual to Production

Not all approaches to managing R2 buckets are created equal. Here's how deployment methods stack up, from what to avoid to what works at scale:

### Level 0: The Manual Dashboard (Anti-Pattern for Production)

**What it is:** Using Cloudflare's web console to click through bucket creation, enabling public access, and configuring CORS.

**What it solves:** Initial experimentation. The dashboard is intuitive for learning R2's capabilities—creating a bucket takes three clicks: name, region, and create.

**What it doesn't solve:** Repeatability, auditability, version control, multi-environment consistency. If you can't codify it, you can't reproduce it reliably across dev/staging/prod or hand it off to another engineer without a screenshot-riddled wiki page.

**Verdict:** Fine for learning the interface or one-off testing. Never for production or even staging environments that matter. Manual configurations drift, settings get forgotten, and troubleshooting becomes archeology.

---

### Level 1: Wrangler CLI (Scriptable, But State-Blind)

**What it is:** Using Cloudflare's Wrangler CLI to create and manage buckets in shell scripts:

```bash
npx wrangler r2 bucket create my-bucket
npx wrangler r2 bucket list
```

**What it solves:** Automation and scriptability. Wrangler is synchronous, works in CI/CD pipelines, and handles authentication via API tokens. You can version-control your scripts and integrate bucket creation into deploy workflows.

**What it doesn't solve:** State management. Scripts don't know what already exists. Running the same create command twice fails (bucket names must be unique per account). There's no plan/preview step. Cleanup on failure is manual. You're executing imperative commands, not declaring desired state.

**Verdict:** Acceptable for dev environments or integration tests where you create and destroy everything in one pass (like ephemeral PR preview environments). Not suitable for production, where you need idempotency, drift detection, and rollback capabilities.

---

### Level 2: Cloudflare API (Maximum Flexibility, High Maintenance)

**What it is:** Calling Cloudflare's REST API directly from custom tooling or configuration management systems:

```bash
curl -X POST "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/r2/buckets" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{"name":"my-bucket"}'
```

**What it solves:** Complete control. You can integrate bucket management into any HTTP-capable tool. The API exposes all R2 features: CORS rules, lifecycle policies, custom domains, event notifications.

**What it doesn't solve:** Abstraction. You're managing HTTP calls, handling authentication, sequencing operations (create bucket before attaching custom domain), and implementing idempotency yourself. There's no state file to track what exists. You're building your own IaC layer from scratch.

**Verdict:** Useful if you're building a custom provisioning system (like OpenMCF) or integrating R2 into broader orchestration. But for most teams, higher-level tools (Terraform, Pulumi) handle the API calls and state management for you.

---

### Level 3: Infrastructure-as-Code (Production-Ready)

**What it is:** Using Terraform or Pulumi with Cloudflare and AWS providers to declaratively define buckets and their lifecycle.

**Terraform example:**

```hcl
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "cloudflare_r2_bucket" "media" {
  account_id = var.account_id
  name       = "media-bucket"
  location   = "weur"  # Western Europe
}

# For CORS, use AWS provider pointing at R2
provider "aws" {
  region                      = "auto"
  skip_credentials_validation = true
  skip_region_validation      = true
  endpoints {
    s3 = "https://${var.account_id}.r2.cloudflarestorage.com"
  }
}

resource "aws_s3_bucket_cors_configuration" "media_cors" {
  bucket = cloudflare_r2_bucket.media.name

  cors_rule {
    allowed_origins = ["https://example.com"]
    allowed_methods = ["GET", "HEAD"]
    allowed_headers = ["*"]
    max_age_seconds = 3600
  }
}
```

**Pulumi example (Go, as used by OpenMCF):**

```go
import (
    "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

bucket, err := cloudflare.NewR2Bucket(ctx, "media-bucket", &cloudflare.R2BucketArgs{
    AccountId: pulumi.String(accountId),
    Name:      pulumi.String("media-bucket"),
    Location:  pulumi.String("weur"),
})
```

**What it solves:** Everything production requires:
- **Declarative configuration**: State what you want, not how to get there
- **State management**: Track what exists, what changed, and dependencies
- **Idempotency**: Running the same config twice produces the same result
- **Plan/preview**: See changes before applying them
- **Version control**: Treat infrastructure as code, with diffs, reviews, and rollbacks
- **Multi-environment support**: Reuse configs for dev/staging/prod with different parameters
- **Drift detection**: Identify manual changes made outside IaC

**What it doesn't solve:** The underlying limitations of R2 (no object versioning, no bucket policies, no static website hosting API). But it makes managing those constraints predictable and reproducible.

**Verdict:** This is the production standard. Both Terraform and Pulumi are solid choices—more on the comparison below.

---

### Level 4: S3-Compatible Tools (Migration and Hybrid Workflows)

**What it is:** Leveraging R2's S3 API compatibility to use existing tools—AWS CLI, s3cmd, rclone, Cyberduck—by simply pointing them at R2's endpoint:

```bash
# AWS CLI configured with R2 endpoint
aws s3 ls --endpoint-url https://$ACCOUNT_ID.r2.cloudflarestorage.com

# rclone for bulk migration
rclone sync s3:my-s3-bucket r2:my-r2-bucket
```

**What it solves:** Migration from S3 to R2 with minimal code changes. Most S3 SDKs (Boto3, AWS SDK for Go/Java/Node) work out-of-the-box—just change the endpoint URL and credentials.

**What it doesn't solve:** Bucket provisioning (you still need to create the bucket via Terraform/API/etc.). S3 tools handle data operations, not infrastructure operations.

**Verdict:** Essential for data migration and hybrid workflows (e.g., serving from R2 while archiving to Glacier). Cloudflare's Sippy feature even automates incremental migration by lazily pulling objects from S3 on-demand, avoiding massive upfront egress fees.

---

## IaC Tool Comparison: Terraform vs. Pulumi for R2

Both Terraform and Pulumi can provision R2 buckets in production. Here's how they compare:

### Terraform: The Battle-Tested Standard

**Maturity:** Terraform has been the IaC gold standard for years. The Cloudflare provider (v4+) officially supports `cloudflare_r2_bucket` for bucket creation and deletion.

**Configuration Model:** Declarative HCL. You define resources, Terraform builds the dependency graph and handles execution order.

**Feature Coverage:** The Cloudflare provider handles basic bucket creation (name, region/location). For advanced features—CORS rules, lifecycle policies, custom domains—you use the AWS provider with a custom S3 endpoint pointed at R2. This hybrid approach works but feels like a workaround until the Cloudflare provider matures.

**State Management:** Local or remote backends (S3, Terraform Cloud, etc.). State tracks bucket IDs and configuration, making updates and deletions predictable.

**Strengths:**
- Broad ecosystem and community support
- Familiarity across ops teams
- Straightforward for standard use cases (create bucket, set region)
- Clear plan/apply workflow

**Limitations:**
- HCL is less expressive than a full programming language (no complex loops, limited conditionals)
- Hybrid provider approach for CORS/lifecycle is awkward
- No native support (yet) for some R2 features like public URL toggle or custom domain attachment

**Verdict:** The default choice for teams already using Terraform or prioritizing stability and ecosystem maturity.

---

### Pulumi: The Programmer's IaC

**Maturity:** Newer than Terraform, but production-ready. The Pulumi Cloudflare provider includes R2 bucket support equivalent to Terraform's.

**Configuration Model:** Real programming languages (TypeScript, Python, Go). You write infrastructure logic as code, with loops, conditionals, and unit tests.

**Feature Coverage:** Similar to Terraform—basic bucket creation via Cloudflare provider, advanced features via AWS provider pointed at R2. The difference is you can wrap these in high-level abstractions using programming constructs.

**State Management:** Pulumi Cloud or self-managed backends (S3, Azure Blob). Similar state tracking to Terraform.

**Strengths:**
- Full programming language expressiveness (easier for dynamic configs, conditional logic)
- Better for complex provisioning (e.g., "create 10 buckets in different regions if prod")
- Native testing frameworks (unit test your infrastructure code)
- Can inline API calls in code (e.g., enable public URL via Cloudflare API after bucket creation)

**Limitations:**
- Smaller community than Terraform
- Requires a runtime (Node.js, Python, Go)
- Bridged provider means any quirks in Terraform's Cloudflare provider carry over

**Verdict:** Great if your team prefers coding infrastructure in familiar languages or needs complex orchestration logic. Slightly more overhead than Terraform for simple use cases.

---

### Which Should You Choose?

- **Default to Terraform** if you want the most mature, widely-adopted solution with straightforward HCL configs.
- **Choose Pulumi** if you prefer writing infrastructure in TypeScript/Python/Go and need advanced logic (dynamic resource generation, complex conditionals, tight integration with application code).
- **Both work equally well** for standard bucket provisioning. The choice is more about team preference and existing tooling than capability.

**OpenMCF uses Pulumi (Go)** because:
1. **Language consistency**: Our broader multi-cloud orchestration is Go-based
2. **Abstraction flexibility**: Pulumi's programming model makes it easier to wrap R2 behind a clean protobuf API
3. **Future-proofing**: As R2 features evolve, programmatic IaC adapts more easily than declarative HCL

That said, the user-facing API is identical regardless of whether we use Terraform or Pulumi under the hood—that's the power of abstraction.

---

## Production Essentials: Configuration, Security, and Cost

Deploying R2 in production requires understanding essential configurations beyond just creating a bucket:

### Public Access: Dev URLs vs. Custom Domains

By default, R2 buckets are **private**—neither listing nor object retrieval is allowed without credentials. For public-facing content (images, videos, static sites), you have two options:

1. **R2-managed public URL** (`https://<bucket>.r2.dev/<object>`): Enable via dashboard or API. **Not recommended for production**—it's rate-limited and lacks Cloudflare CDN features (caching, WAF, bot protection). Use only for dev/test.

2. **Custom domain** (`https://media.example.com/<object>`): Attach a custom domain to the bucket via Cloudflare's custom domain API. The domain gets full CDN treatment—caching, security, page rules, etc. **This is the production approach.**

**Best Practice:** Never enable public access for buckets with sensitive data. Use Cloudflare Access (Zero Trust authentication) if you need to restrict access to public buckets.

---

### Security and Access Control

R2 uses **Access Keys** (key ID + secret) for authentication, akin to AWS IAM keys. These are scoped to R2 operations and generated per account. Unlike S3, R2 **does not support bucket policies or ACLs**—access is all-or-nothing per bucket via keys.

**For Cloudflare Workers**, R2 bindings provide secure access without exposing keys (the binding is authorized under your account).

**Best Practices:**
- Use API tokens with minimal scope (`R2 Storage: Edit` only, scoped to specific account)
- Rotate Access Keys regularly
- Never hard-code keys in code—use environment variables or secret managers
- For public buckets, consider Cloudflare Access policies to restrict who can read

---

### Durability and Data Location

R2 offers **11×9's durability** (99.999999999% annual durability)—the same level as AWS S3, GCS, and Azure Blob. Cloudflare achieves this via erasure coding and replication across multiple locations within a region.

**Location hints** let you choose a broad region (Western Europe, Eastern North America, Asia-Pacific, etc.). Cloudflare stores copies within that region for latency optimization.

**Jurisdictional restrictions** let you enforce data residency:
- `EU` jurisdiction: Guarantees data stays within the EU (GDPR compliance)
- `FedRAMP` jurisdiction: For US government compliance (enterprise only)

**Strong consistency**: After an object write or delete, all global clients immediately see the updated state. No eventual-consistency surprises.

---

### Lifecycle Management and Cost Optimization

R2 supports **lifecycle rules** (like S3) to automate data transitions and deletions:

- **Expiration**: Automatically delete objects after N days (e.g., delete logs after 90 days)
- **Transition to Infrequent Access (IA)**: Move objects to cheaper IA storage after inactivity (e.g., transition to IA after 30 days)

IA storage costs $0.010/GB-month (vs. $0.015 for Standard) but has a $0.01/GB retrieval fee and 30-day minimum storage duration. Use IA for archival data that's rarely accessed.

**Cost Optimization Tips:**
1. **Right-size storage class**: Use Standard for frequently accessed data, IA for cold data
2. **Delete ephemeral data**: Use lifecycle rules to auto-delete temp files, logs, or derived data
3. **Watch for orphaned buckets**: Deleted apps might leave buckets behind—audit regularly
4. **Leverage zero egress**: Design architectures to maximize data served from R2 (e.g., use R2 as CDN origin instead of S3 + CloudFront)

**Free Tier:** Cloudflare offers 10 GB storage, 1 million Class A ops, and 10 million Class B ops per month for free—great for dev/test or small production workloads.

---

### Logging, Analytics, and Event Notifications

- **Analytics**: Cloudflare dashboard shows storage usage, operation counts, and egress (which should be zero). Use for cost monitoring.
- **Logpush**: Send access logs to Cloudflare analytics or external SIEM (similar to S3 server access logs).
- **Event Notifications**: Push object create/delete events to Cloudflare Queues, triggering Workers or microservices (like S3 → SNS/SQS). Useful for serverless data pipelines (e.g., image upload → Worker resize → store back to R2).

---

### What R2 Doesn't Support (Know the Gaps)

R2 is **not** a drop-in S3 replacement for all use cases. Missing features:

- **Object Versioning**: R2 does not support versioning. Every upload overwrites the previous object (like S3 with versioning disabled).
- **Bucket Policies / ACLs**: No IAM-style policies or ACLs. Access is managed via keys and Cloudflare Access.
- **Static Website Hosting**: No built-in index/error document configuration. Use Cloudflare Pages or Workers instead.
- **S3 Select**: No SQL queries on object contents. (R2 SQL is separate, in beta.)
- **Multi-Attach**: Buckets can't be "attached" to compute instances—R2 is pure object storage, not block storage.

These omissions simplify R2 (less surface area, lower cost) but mean you can't blindly migrate complex S3 setups. Plan accordingly.

---

## 80/20 Configuration Analysis: What Most Users Need

When integrating R2 into OpenMCF's multi-cloud IaC framework, we design the schema to cover **80% of use cases with minimal, intuitive config**, while allowing advanced features as opt-in.

### Essential Fields (The 80%)

1. **Bucket Name**: Unique identifier (3-63 chars, lowercase alphanumeric + hyphens). **Required.**

2. **Account ID**: Cloudflare account to create the bucket in. **Required.** (OpenMCF may infer this from provider config, but it must be specified somewhere.)

3. **Region/Location**: Location hint (`wnam`, `enam`, `weur`, etc.) for latency optimization. **Recommended.** Default to `auto` (Cloudflare chooses) if unspecified.

4. **Public Access**: Boolean to indicate if the bucket should be publicly accessible. **Common.** If `true`, decide between:
   - `public_dev_url`: Enable `r2.dev` URL (dev/test only)
   - `custom_domain`: Attach a custom domain (production)

### Common Options (The Next 15%)

5. **CORS Settings**: Allow cross-origin requests from browsers. Common for public asset buckets. A simple `allowed_origins` list and `allowed_methods` covers 95% of cases.

6. **Lifecycle Policy**: Auto-expiration or transition to IA after N days. Most users want a single global rule (e.g., "delete after 90 days" or "move to IA after 30 days"). Advanced multi-rule lifecycles are rare.

7. **Custom Domain**: For production public buckets. Takes a domain name (e.g., `media.example.com`) and optionally auto-creates DNS record.

### Rare/Advanced Options (The 5%)

8. **Jurisdiction**: EU or FedRAMP compliance. Most users don't need this unless legally required.

9. **Event Notifications**: Push object events to Cloudflare Queues. Useful for serverless pipelines but niche.

10. **Bucket Lock / WORM**: Write-once-read-many retention policies (compliance use cases). Very rare.

11. **Infrequent Access Default**: Most buckets start with Standard storage; use lifecycle rules to transition. Defaulting to IA is unusual.

### What We Omit from the Spec

- **Versioning**: R2 doesn't support it, so the spec does not model it.
- **Bucket Policies / ACLs**: Not applicable to R2.
- **Static Website Hosting**: Not a bucket-level config in R2 (use Cloudflare Pages/Workers instead).

---

## OpenMCF's Approach: Abstraction with Pragmatism

OpenMCF abstracts R2 provisioning behind a clean, protobuf-defined API (`CloudflareR2Bucket`). This provides a consistent interface across clouds while respecting R2's unique characteristics.

### The `CloudflareR2BucketSpec` (Simplified)

Based on the 80/20 analysis, the ideal spec includes:

```protobuf
message CloudflareR2BucketSpec {
  // Bucket name (3-63 chars, lowercase alphanumeric + hyphens)
  string bucket_name = 1;

  // Cloudflare account ID (32-char hex)
  string account_id = 2;

  // Region/location hint (wnam, enam, weur, etc.; auto = let Cloudflare choose)
  CloudflareR2Location location = 3;

  // Enable public access via r2.dev URL (dev/test only)
  bool public_access = 4;

  // Optional: Custom domain for production public access
  string custom_domain = 5;

  // Optional: CORS settings
  CorsConfig cors = 6;

  // Optional: Lifecycle rules
  LifecycleConfig lifecycle = 7;

  // Optional: Jurisdiction (EU, FedRAMP)
  string jurisdiction = 8;
}
```

**Note:** R2 does not support object versioning, so the spec deliberately does not model it.

### Default Choices

- **Location**: Default to `auto` (Cloudflare chooses optimal region) unless user specifies.
- **Public Access**: Default to `false` (private bucket).
- **CORS**: Omit by default (no CORS = no cross-origin browser access).
- **Lifecycle**: Omit by default (data persists indefinitely unless explicitly managed).

### Under the Hood: Pulumi (Go)

OpenMCF uses **Pulumi (Go)** to provision R2 buckets. Why?

1. **Language consistency**: Our broader multi-cloud stack is Go-based.
2. **Equivalent coverage**: Pulumi's Cloudflare provider supports all R2 operations we need (bucket creation, CORS via AWS provider, etc.).
3. **Future-proofing**: Pulumi's programming model adapts more easily to evolving R2 features (event notifications, advanced lifecycle, etc.).

The protobuf API remains the same regardless of backend—users never see the Pulumi/Terraform choice.

---

## Configuration Examples: Dev, Staging, Production

### Development: Private Bucket for Testing

**Use Case:** Backend dev storing test uploads.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: dev-test-uploads
spec:
  bucketName: dev-test-uploads
  account_id: 1234567890abcdef1234567890abcdef
  location: wnam
  public_access: false
```

**Rationale:**
- Private by default (no public access needed for dev)
- Western North America region (close to dev team)
- No CORS, no lifecycle (ephemeral test data)

---

### Staging: Public Assets with CORS

**Use Case:** Staging web app serving images and fonts from R2.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: staging-assets
spec:
  bucketName: staging-assets
  account_id: 1234567890abcdef1234567890abcdef
  location: weur
  public_access: true
  public_dev_url: true  # Use r2.dev for staging (custom domain not needed yet)
  cors:
    allowed_origins:
      - "https://staging.example.com"
    allowed_methods:
      - GET
      - HEAD
    max_age_seconds: 3600
```

**Rationale:**
- Public `r2.dev` URL for quick staging access
- CORS restricted to staging domain
- Western Europe region (staging infra in Europe)

---

### Production: Custom Domain, Lifecycle, and IA

**Use Case:** Production media hosting for a video streaming platform.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: prod-media
spec:
  bucketName: prod-media
  account_id: 1234567890abcdef1234567890abcdef
  location: enam
  public_access: true
  custom_domain: media.example.com
  cors:
    allowed_origins:
      - "https://www.example.com"
      - "https://app.example.com"
    allowed_methods:
      - GET
      - HEAD
    max_age_seconds: 86400
  lifecycle:
    transition_to_ia_days: 30  # Move to Infrequent Access after 30 days
    expiration_days: 365        # Delete after 1 year
```

**Rationale:**
- Custom domain (`media.example.com`) for production CDN delivery
- CORS restricted to production domains
- Lifecycle rules: transition to IA (save storage cost) after 30 days, delete after 1 year (assuming user-generated content with limited shelf life)
- Eastern North America region (primary user base in US)

---

### Production: EU Compliance Bucket

**Use Case:** European SaaS storing user data with GDPR requirements.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: prod-user-data-eu
spec:
  bucketName: prod-user-data-eu
  account_id: 1234567890abcdef1234567890abcdef
  location: weur
  jurisdiction: EU  # Enforce data stays within EU
  public_access: false
  lifecycle:
    expiration_days: 2555  # 7 years retention (GDPR compliance)
```

**Rationale:**
- EU jurisdiction guarantees data residency (GDPR compliance)
- Private bucket (user data, not public assets)
- Western Europe region + EU jurisdiction
- Lifecycle retention matches legal requirements (7 years, then auto-delete)

---

## Migration Strategies: Moving from S3 to R2

R2's S3 compatibility makes migration straightforward. Here are proven strategies:

### 1. Manual/Scripted Copy (Small to Medium Data)

Use `aws s3 sync` or `rclone`:

```bash
# AWS CLI (configure endpoint for R2)
aws s3 sync s3://my-s3-bucket s3://my-r2-bucket \
  --endpoint-url https://$ACCOUNT_ID.r2.cloudflarestorage.com

# rclone (more efficient, resume support)
rclone sync s3:my-s3-bucket r2:my-r2-bucket --progress
```

**Pros:** Simple, works for <1 TB datasets.  
**Cons:** AWS egress fees apply. For large datasets, this can be expensive.

---

### 2. Cloudflare Sippy (Incremental/Lazy Migration)

**What it is:** Enable Sippy on an R2 bucket, pointing it at your source S3 bucket. When an object is requested from R2 that doesn't exist yet, R2 fetches it from S3, serves it to the user, and stores it in R2.

**Pros:** Zero downtime. No massive upfront egress cost. Only accessed data migrates.  
**Cons:** First access per object is slower (cache-miss penalty). Requires S3 credentials in R2.

**Use Case:** Gradual cutover from S3 to R2 for content delivery (e.g., media hosting, asset CDN).

---

### 3. Cloudflare Super Slurper (Bulk Migration)

**What it is:** Cloudflare's server-side bulk transfer service. You provide source credentials, and Cloudflare copies the entire bucket to R2.

**Pros:** Faster than client-side sync. Potentially lower egress cost (depending on Cloudflare's peering with source cloud).  
**Cons:** Still incurs source egress fees. Requires coordination with Cloudflare (may need enterprise support).

**Use Case:** Large-scale migration (multi-TB datasets) with a defined cutover window.

---

### 4. Hybrid Approach (Production Best Practice)

1. **Create R2 bucket** with IaC (Terraform/Pulumi)
2. **Enable Sippy** for lazy migration of existing data
3. **Redirect new uploads** to R2 (change application config)
4. **Run Super Slurper** or `rclone` in background to migrate remaining data
5. **Decommission S3 bucket** once all data is in R2 and traffic is stable

**Pros:** Minimal downtime, controlled migration, cost-effective.  
**Cons:** Requires careful application config changes and monitoring.

---

## Key Takeaways

1. **R2 eliminates egress fees**, making it ideal for content delivery, media hosting, backups, and multi-cloud architectures. This is a strategic cost advantage over S3, GCS, and Azure.

2. **R2 is S3-compatible** for data operations (PUT/GET/DELETE, multi-part uploads, presigned URLs), but lacks advanced S3 features (versioning, bucket policies, static website hosting). Design accordingly.

3. **Manual management is an anti-pattern.** Use IaC (Terraform or Pulumi) for production. The Cloudflare provider is mature and stable.

4. **The 80/20 config is bucket name, account ID, location, and public access settings.** Advanced features (CORS, lifecycle, custom domains) are opt-in for specific use cases.

5. **For public buckets, use custom domains in production.** The `r2.dev` URL is rate-limited and lacks CDN features—fine for dev/test, not for production.

6. **R2's strong consistency** eliminates eventual-consistency pitfalls. After a write or delete, all clients see the updated state immediately.

7. **Migration from S3 is straightforward** thanks to S3 API compatibility. Use Sippy for lazy migration, `rclone` for bulk copy, or hybrid approaches for production cutover.

8. **OpenMCF abstracts R2** into a clean protobuf API, making multi-cloud deployments consistent while respecting R2's unique characteristics (zero egress, simplified security model, no versioning).

---

## Conclusion: The Multi-Cloud Storage Play

Cloudflare R2 represents a paradigm shift in cloud object storage: **pay for what you store and how you use it, not for how much you serve**. For workloads that deliver data globally—media streaming, software distribution, public datasets, backup repositories—R2's zero-egress model can reduce costs by 90% or more compared to the big-3 clouds.

R2 isn't a universal S3 replacement. It's simpler, cheaper, and more opinionated—optimized for the common case of storing and serving data, not for deep AWS integrations. But for multi-cloud architectures, that simplicity is a strength. You can store once in R2 and serve to AWS, GCP, Azure, or on-prem without exit fees or vendor lock-in.

Deploy R2 with Terraform or Pulumi. Design lifecycle policies to manage costs. Use custom domains for production public buckets. Leverage Sippy for migrations. And let OpenMCF abstract the complexity behind a clean, protobuf-defined API that works the same way whether you're deploying to Cloudflare, AWS, or GCP.

The cloud storage wars are over. Cloudflare won on egress. The question is: are you ready to take advantage?

---

## Further Reading

- **Cloudflare R2 Documentation:** [Cloudflare R2 Docs](https://developers.cloudflare.com/r2/)
- **Terraform Cloudflare Provider:** [GitHub - cloudflare/terraform-provider-cloudflare](https://github.com/cloudflare/terraform-provider-cloudflare)
- **Pulumi Cloudflare Provider:** [Pulumi Cloudflare Provider](https://www.pulumi.com/registry/packages/cloudflare/)
- **R2 Pricing:** [Cloudflare R2 Pricing](https://developers.cloudflare.com/r2/pricing/)
- **R2 API Compatibility:** [S3 API Compatibility](https://developers.cloudflare.com/r2/api/s3/api/)
- **Migration with Sippy:** [Sippy Incremental Migration](https://blog.cloudflare.com/sippy-incremental-migration-s3-r2/)
- **Cloudflare Workers + R2:** [Using R2 with Workers](https://developers.cloudflare.com/r2/api/workers/)

---

**Bottom Line:** Cloudflare R2 gives you S3-compatible object storage with zero egress fees, strong consistency, and production-grade durability at a fraction of the cost of AWS/GCP/Azure for content delivery and multi-cloud workflows. Manage it with Terraform or Pulumi, design for the 80% use case, and let OpenMCF abstract the complexity into a clean, protobuf API. The result: predictable costs, portable architecture, and freedom from vendor lock-in.

