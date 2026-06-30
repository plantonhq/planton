# AWS CloudFront Deployment: From Click-Ops to Production Infrastructure

## Introduction

AWS CloudFront is often misunderstood as "just another CDN." While it certainly accelerates static content delivery, its true power lies in its role as a **programmable, globally-distributed application perimeter**. It serves as the security gateway, performance accelerator, and cost optimizer for modern AWS applications—whether they're serving static React builds from S3, dynamic APIs from Application Load Balancers, or real-time gRPC streams.

The challenge isn't whether to use CloudFront. For AWS-native applications, the decision is often architectural: free data transfer from any AWS origin, deep integration with WAF and Shield, and native support for Lambda@Edge make it the default choice. The real challenge is **how to deploy and manage it without drowning in configuration complexity**.

This document explores the deployment landscape for CloudFront distributions, from manual console operations to production-grade Infrastructure as Code, and explains why Planton provides an opinionated, simplified API that gets you to production faster without sacrificing control.

## The CloudFront Deployment Landscape

### Level 0: The AWS Management Console (The Anti-Pattern)

The AWS Management Console is where most CloudFront journeys begin. It's visual, exploratory, and immediately accessible. For learning and one-off experiments, it serves its purpose.

**For production, it's an anti-pattern.**

The console workflow leads to several predictable failure modes:

- **The S3 Permission Trap**: The most common production incident. You create a distribution pointing to an S3 bucket, test it, and receive "Access Denied" errors. The bucket policy wasn't updated to grant CloudFront's service principal read access. Now you're debugging IAM policies in production.

- **The CORS Blackhole**: Your application works perfectly when accessing S3 directly but breaks intermittently when served through CloudFront. The issue: the cache behavior doesn't forward the `Origin` header, so CORS requests fail unpredictably depending on cache state.

- **The Certificate Region Mistake**: You provision an ACM certificate in `eu-west-1` (where your application lives) and attach it to your distribution. CloudFront rejects it. All CloudFront certificates **must** be in `us-east-1`, regardless of your origin's region. This is non-negotiable, non-obvious, and not validated until deploy-time.

- **The Caching Disaster**: You forward all headers, cookies, and query strings to "be safe." Your cache hit ratio is 2%. Every request goes to the origin. Your S3 data transfer costs explode, and your dynamic origin is crushed under load.

These aren't hypothetical. They represent the four most common CloudFront configuration failures in production environments. The console doesn't prevent them. It enables them.

AWS itself acknowledges this reality. Their new "Console-to-Code" feature uses generative AI to record your manual actions and output equivalent CloudFormation or CDK code. This is AWS explicitly building a bridge **away** from manual configuration toward Infrastructure as Code.

**Verdict**: The console is for learning. Production requires code.

### Level 1: Scripted Imperative (AWS CLI and Boto3)

The AWS CLI and SDKs like Boto3 represent the first step toward automation. You can script distribution creation, updates, and deletion using commands like `aws cloudfront create-distribution`.

This approach is imperative—you execute a sequence of commands—but critically, it is **not declarative** and **not idempotent**. Running the same script twice doesn't ensure the same state; it often fails or creates duplicates.

**Where This Succeeds**: For **discrete actions** on existing distributions within CI/CD pipelines, the CLI is ideal. The canonical example is cache invalidation:

```bash
aws cloudfront create-invalidation \
  --distribution-id E2QWRUHAPOMQZL \
  --paths "/*"
```

This is a stateless action. It's perfect for a CLI command in a deployment script.

**Where This Fails**: Managing the **state** of a distribution. If your Boto3 script creates a distribution and you run it again, you've now created a second distribution. If you modify a behavior setting, there's no automatic diffing or rollback. State management is manual, error-prone, and fragile.

**Verdict**: Use the CLI for actions (invalidation, monitoring). Avoid it for managing infrastructure state.

### Level 2: Configuration Management (Ansible)

Ansible's `community.aws` collection provides CloudFront modules like `cloudfront_distribution`, which allows declarative distribution management as part of a larger system configuration workflow.

This is a step up from imperative scripts: Ansible modules are idempotent by design. Running the playbook multiple times converges to the desired state.

**Where This Fits**: Organizations with Ansible-first infrastructure, where CloudFront is managed alongside application deployments, system configuration, and multi-cloud resources.

**The Trade-off**: You're coupling CloudFront infrastructure to an Ansible ecosystem. If the rest of your stack uses Terraform or CDK, you've now introduced a second IaC tool and workflow for this one component.

**Verdict**: Viable for Ansible-centric organizations. Otherwise, it's an architectural mismatch.

### Level 3: Control Plane Abstraction (Crossplane)

Crossplane represents a paradigm shift: it treats cloud infrastructure as native Kubernetes Custom Resource Definitions (CRDs). The Crossplane AWS provider exposes CloudFront resources like `Distribution`, `CachePolicy`, and `OriginAccessIdentity` as `cloudfront.aws.crossplane.io` CRDs that you can `kubectl apply`.

But the real power is **Compositions**—a Crossplane feature that allows platform teams to create high-level abstractions. For example, a platform team could define a `StaticWebsite` CRD that, when applied, automatically provisions an S3 bucket, a CloudFront distribution, an Origin Access Control (OAC), and applies the correct bucket policy—all as a single, opinionated resource.

**Where This Shines**: Platform engineering teams building self-service infrastructure catalogs on Kubernetes. Developers apply a simple YAML manifest, and Crossplane orchestrates the complex multi-resource provisioning behind the scenes.

**The Learning Curve**: You're adopting Kubernetes as your control plane for all infrastructure. This is a strategic, organization-wide decision, not a tactical choice for one CDN.

**Verdict**: Excellent for Kubernetes-native platform teams. Overkill for teams just deploying a CloudFront distribution.

## Production-Grade Infrastructure as Code: The Tool Wars

For production environments, Infrastructure as Code is the baseline. The question is **which** IaC tool. This decision reveals significant trade-offs in abstraction, control, and developer experience.

### CloudFormation: The Low-Level AWS Native

CloudFormation is AWS's foundational IaC service, managing resources via JSON or YAML templates. It's the "compilation target" for higher-level tools like CDK.

**Pros**:
- Native integration with every AWS service
- State managed automatically within AWS (no external state file)
- Template can be reviewed in the console

**Cons**:
- **Verbosity**: A simple CloudFront distribution with a Lambda@Edge function can exceed 50 lines of YAML, filled with repetitive structure and arcane property names.
- **Deployment Speed**: Notoriously slow. Updates can take 5-10 minutes even for minor changes.
- **Rollback Hell**: Deployments can get "stuck" in an unrecoverable state, requiring manual AWS Support intervention.
- **Poor Diff Experience**: "Change sets" exist but are far less readable than Terraform's plan output.
- **Multi-Region Pain**: Managing the us-east-1 certificate dependency requires separate nested stacks or custom resources.

**Verdict**: It works, but it's low-level plumbing. Not designed for humans to write directly.

### AWS CDK: The Abstracted AWS Native

The AWS Cloud Development Kit (CDK) lets you define infrastructure in TypeScript, Python, or other languages. It synthesizes your code into CloudFormation templates and deploys them.

**Pros**:
- **High-Level Constructs**: L2/L3 constructs abstract common patterns. For example, the `S3Origin` construct automatically configures the Origin Access Control and applies the correct S3 bucket policy for you.
- **Superior Edge Function Handling**: This is CDK's killer feature for CloudFront. It automatically handles packaging, zipping, and deploying Lambda@Edge code from a local directory. You point it at your code folder, and it handles the rest. This solves a **massive** developer experience pain point that Terraform struggles with.
- **Type Safety**: If you're using TypeScript, you get autocomplete and compile-time validation of properties.

**Cons**:
- **Still Bound by CloudFormation Speed**: Because CDK synthesizes to CloudFormation, deployments are still slow.
- **Abstraction Leakage**: High-level constructs are convenient until they're not. When you hit an edge case, you're debugging generated CloudFormation templates.

**Verdict**: If you're AWS-only and prioritize developer experience (especially for edge functions), CDK is compelling. Its design philosophy—abstracting common patterns—is a model for what Planton aims to emulate.

### Terraform/OpenTofu: The Cloud-Agnostic Standard

Terraform (and its open-source fork, OpenTofu) is the de facto cloud-agnostic IaC standard. It uses declarative HCL syntax and calls AWS APIs directly. For CloudFront, Terraform and OpenTofu are functionally identical.

**Pros**:
- **Cloud-Agnostic**: Manages AWS, GCP, Azure, and 3,000+ other providers with a unified workflow.
- **Fast Deployments**: Significantly faster than CloudFormation because it calls APIs directly.
- **Readable Diffs**: `terraform plan` output is clear, color-coded, and shows exactly what will change.
- **Mature Ecosystem**: Extensive module library and community support.

**Cons**:
- **Verbosity**: More verbose than CDK. The `aws_cloudfront_distribution` resource requires explicit configuration of every behavior, origin, and policy.
- **Manual State Management**: You must configure remote state storage (e.g., S3 backend) and handle state locking. This is boilerplate for every project.
- **The us-east-1 Certificate Gotcha**: This is Terraform's critical CloudFront pain point. Because CloudFront is global, it requires its ACM certificate to be in `us-east-1`, regardless of your origin's region. If your infrastructure is in `eu-west-1`, you must define a **second** AWS provider instance with `alias` and `region = "us-east-1"`, then explicitly reference this aliased provider in your `aws_acm_certificate` resource. This is non-intuitive, error-prone, and fails at apply-time if you get it wrong.
- **Edge Function Deployment Hell**: Unlike CDK, Terraform doesn't automatically package Lambda@Edge code. You must manually create an `archive_file` data source to zip your code, create the `aws_lambda_function` resource with `publish = true`, configure the IAM role with the `edgelambda.amazonaws.com` principal, and attach the versioned ARN to the cache behavior. This is a complex, multi-step, error-prone workflow that CDK handles transparently.

**Verdict**: Terraform is the industry standard for multi-cloud infrastructure. For CloudFront specifically, it's powerful but demands you handle the multi-region certificate and edge function complexity manually.

### Pulumi: The Middle Ground

Pulumi combines CDK's programming language approach (TypeScript, Python, Go) with Terraform's direct API calls and independent state management. It bypasses CloudFormation entirely.

**Pros**:
- **High-Level Abstraction with Direct Control**: Uses language features (loops, functions, conditionals) without synthesizing to an intermediate template format.
- **Fast Deployments**: Like Terraform, it calls AWS APIs directly.
- **Automatic Asset Handling**: Similar to CDK, it can automatically package and deploy Lambda@Edge code from a local path.

**Cons**:
- **Smaller Ecosystem**: Fewer modules and examples compared to Terraform.
- **State Management**: Requires external state storage (Pulumi Service, S3, etc.).

**Verdict**: A strong middle ground between CDK's abstraction and Terraform's control. Ideal if you value using real programming languages but want to avoid CloudFormation's slowness.

## The IaC Tool Comparison Summary

| Tool | Abstraction Level | us-east-1 ACM Handling | Lambda@Edge Deployment | State Management | Deployment Speed |
|------|------------------|------------------------|------------------------|------------------|------------------|
| **CloudFormation** | Low (Declarative YAML) | Manual (separate stack required) | Manual (zip + S3 or inline) | AWS-Managed (Stack) | Slow (5-10 min) |
| **AWS CDK** | High (Imperative Code) | Abstracted (via Certificate L2) | **Automatic** (from local path) | AWS-Managed (via CFN) | Slow (via CFN) |
| **Terraform** | Low (Declarative HCL) | **Manual** (requires aliased provider) | **Manual** (archive_file + IAM + publish) | External (S3, etc.) | Fast (API direct) |
| **Pulumi** | High (Imperative Code) | Abstracted (via component) | **Automatic** (asset packaging) | External (Pulumi Service, S3) | Fast (API direct) |

The battlegrounds are clear:
1. **us-east-1 Certificate Management**: CDK and Pulumi abstract this. Terraform forces you to handle it manually.
2. **Lambda@Edge Deployment**: CDK and Pulumi automate code packaging. Terraform makes you build the pipeline yourself.

## Production Patterns and the 80/20 Rule

The `aws_cloudfront_distribution` resource in Terraform exposes over 40 top-level fields, with nested objects for origins, behaviors, policies, and error responses. A complete distribution configuration can exceed 200 lines of HCL.

**But here's the key insight**: the vast majority (80%) of production CloudFront deployments use a **small subset** (20%) of these fields.

### The Minimal 80% Configuration

For a production static website served from S3 with a custom domain and WAF protection, the essential fields are:

- `enabled` (always `true`)
- `origins` (one S3 bucket with Origin Access Control)
- `default_cache_behavior` (targets the origin, enforces HTTPS)
- `aliases` (custom domain, e.g., `www.example.com`)
- `viewer_certificate` (ACM certificate ARN from us-east-1)
- `default_root_object` (`"index.html"`)
- `price_class` (`PriceClass_100` for cost optimization)
- `web_acl_id` (AWS WAF WebACL for security)
- `http_version` (`http2`)

This is the "Standard 80%" archetype. It covers the overwhelming majority of use cases.

### The Long-Tail 20%

The remaining 20% of use cases require advanced features:
- `ordered_cache_behavior` (path-based routing, e.g., `/api/*` to an ALB)
- `lambda_function_association` (edge compute for auth or routing)
- `origin_groups` (origin failover for high availability)
- `custom_error_response` (custom 404 pages)
- `restrictions` (geo-blocking)

These are powerful, legitimate features. But they're not the baseline.

## Planton's Approach: Opinionated Simplicity

Planton's `AwsCloudFrontSpec` protobuf API is deliberately designed around the 80/20 principle. It exposes:

- `enabled` (Boolean)
- `aliases` (repeated string, validated as FQDNs)
- `certificate_arn` (string, validated to enforce us-east-1 region)
- `price_class` (enum: `PRICE_CLASS_100`, `PRICE_CLASS_200`, `PRICE_CLASS_ALL`)
- `origins` (repeated message with `domain_name`, `origin_path`, and `is_default`)
- `default_root_object` (string, e.g., `"index.html"`)

### What This Achieves

1. **Eliminates the us-east-1 Gotcha**: The `certificate_arn` field is validated at the protobuf level to ensure it's in `us-east-1`. The controller can even auto-provision the certificate if integrated with an ACM resource.

2. **Secure by Default**: The controller automatically creates an Origin Access Control (OAC) for S3 origins and applies the correct bucket policy. You don't wire this up manually.

3. **Cost-Optimized by Default**: The default `price_class` is `PRICE_CLASS_100` (North America and Europe only), forcing users to consciously opt into higher global costs.

4. **Simplified Multi-Origin Support**: The `origins` list supports multiple origins, with exactly one marked as `is_default`. For advanced path-based routing, users can extend with ordered cache behaviors in future API versions or use the Pulumi/Terraform modules directly for the long-tail 20%.

### What This Doesn't Attempt

Planton does **not** try to expose every CloudFront parameter. It doesn't compete with CloudFormation's exhaustive coverage or Terraform's flexibility.

**Instead, it optimizes for the 80% use case**: deploying a secure, performant, cost-effective static or dynamic site in ~20 lines of protobuf configuration, with escape hatches to Pulumi or Terraform for power users.

## Production Best Practices (Baked Into the Framework)

Based on the research, several anti-patterns emerge repeatedly. Planton's design proactively prevents them:

### Anti-Pattern 1: Public S3 Buckets
**The Problem**: Leaving S3 buckets publicly accessible when serving via CloudFront is a critical security flaw.

**Planton's Solution**: The controller automatically creates an Origin Access Control (OAC) and applies the generated S3 bucket policy. The bucket stays private. This is inspired by CDK's `S3Origin` construct.

### Anti-Pattern 2: Inefficient Caching
**The Problem**: Forwarding all headers, cookies, and query strings destroys cache hit ratio, increases origin load, and inflates costs.

**Planton's Solution**: Future versions will support explicit cache policy configuration using AWS Managed Policies (e.g., `Managed-CachingOptimized`) as the default, with whitelisting for custom requirements.

### Anti-Pattern 3: Relying on Cache Invalidation for Deployments
**The Problem**: Using `aws cloudfront create-invalidation` after every deployment is slow (minutes to propagate globally), costly (after 1,000 free paths/month), and causes "thundering herd" cache refills.

**The Gold Standard**: **File Versioning**. Build pipelines (e.g., Webpack) produce assets with content hashes in filenames (`app.a1b2c3d4.js`). These versioned files are cached "forever" (`max-age=31536000`). Only `index.html` (with a short TTL) is updated to reference the new files. This provides instant, atomic updates with zero invalidation cost.

**Planton's Future Support**: Documentation and example pipelines will demonstrate file versioning as the recommended deployment pattern.

### Anti-Pattern 4: Long TTL on index.html
**The Problem**: Caching `index.html` with a long TTL means users with a cached copy won't see new deployments because their browser won't fetch the updated "manifest" pointing to new assets.

**Best Practice**: `index.html` should have a TTL of 0-60 seconds, or its own cache behavior with `max-age=0`.

## Cost Optimization: Architecture as a Lever

CloudFront's pricing model rewards good architecture and punishes bad decisions:

### The Free Origin Fetch Incentive
**Data transfer from any AWS origin (S3, ALB, EC2) to CloudFront is FREE.** This is a massive financial incentive to keep your entire stack within AWS. Multi-cloud or on-premise origins incur full data transfer costs.

### Price Classes: The Geography-Cost Trade-Off
- `PriceClass_ALL`: Uses all global POPs. Best performance. Highest cost.
- `PriceClass_200`: Excludes expensive regions (South America, Australia). Moderate cost.
- `PriceClass_100`: North America, Europe, Israel only. **Lowest cost.**

If your user base is concentrated in NA/EU, `PriceClass_100` is a direct, significant cost savings with minimal performance impact. **Planton defaults to this.**

### Maximizing Cache Hit Ratio
Every cache **hit** is an origin fetch you didn't pay for. Tight cache policies (whitelisting only required headers/cookies/query strings) are the single most impactful cost optimization.

### Origin Shield: When to Pay to Save
Origin Shield is a **paid** centralized caching layer that collapses origin requests. If your origin is load-sensitive (e.g., just-in-time video transcoding) or multi-cloud (incurring high data transfer), Origin Shield's fee is offset by reduced origin load and transfer costs.

## The Competitive Context: Why CloudFront for AWS-Native Apps

CloudFront operates in a mature, competitive CDN market:

- **Cloudflare**: Security-first, generous free tier, easiest to use. Often chosen as a comprehensive security proxy, not just a CDN.
- **Fastly**: Developer-focused, real-time control, instant purging. Chosen for media and applications requiring sub-second cache updates.
- **Akamai**: Enterprise powerhouse with the largest global network. Chosen for massive, mission-critical scale.
- **CloudFront**: Deep AWS ecosystem integration. Chosen when the rest of your stack is AWS.

**The strategic insight**: CloudFront isn't competing on raw performance (it's comparable) or price (it's competitive). It competes on **integration**. Free origin fetches, native WAF/Shield integration, VPC origins, and seamless IAM permissions make it the path of least resistance for AWS-native applications.

**Planton's value proposition aligns with this**: If you're building on AWS, CloudFront is the default CDN choice. Planton makes deploying and managing it dramatically simpler without sacrificing production-readiness.

## Conclusion: Simplicity as a Strategic Choice

The CloudFront configuration surface is vast. The low-level tools (CloudFormation, Terraform) expose every parameter because they're designed for completeness. The high-level tools (CDK, Pulumi) abstract common patterns because they're designed for developer experience.

Planton chooses **opinionated simplicity**. It assumes you're deploying the 80% use case: a secure, performant, cost-effective static or dynamic site with custom domains, HTTPS, and WAF protection. It encodes AWS best practices—OAC for S3 origins, `PriceClass_100` by default, us-east-1 certificate validation—directly into the API and controller logic.

For the 20% of advanced use cases (multi-origin routing, Lambda@Edge auth, origin failover), Planton provides escape hatches: direct Pulumi module access or Terraform module usage for those who need full control.

The goal isn't to replace Terraform. The goal is to make the default case—deploying a production CloudFront distribution—as simple as defining 20 lines of protobuf configuration and running `planton apply`.

Because the best infrastructure is the infrastructure you don't have to think about.

