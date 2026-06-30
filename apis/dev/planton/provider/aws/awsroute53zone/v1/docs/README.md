# AWS Route53 DNS: The Modern Infrastructure Landscape

## Introduction: DNS as Code is No Longer Optional

DNS has always been critical infrastructure, but for decades it was managed like cattle—manually updated in web consoles, tracked in spreadsheets, and prayed over during cutover windows. A single typo in a DNS record could take down an entire production service, and nobody could tell you what changed or when.

**That era is over.**

Modern infrastructure demands that DNS be treated as code: versioned, reviewed, tested, and deployed through automated pipelines. AWS Route53 provides the foundational DNS service, but **how** you manage it determines whether DNS becomes a strategic asset or an operational liability.

This isn't about choosing between Route53 and another DNS provider—Route53 wins on features, global reach, and AWS integration. The real question is: **Do you manage Route53 records through manual console clicks, brittle scripts, or declarative Infrastructure-as-Code?**

The stakes are higher than you think:

- **Public zones** control how the world reaches your services. A misconfigured MX record means lost email. A forgotten trailing dot in a CNAME can break your entire domain.
- **Private zones** enable split-horizon DNS for internal services across VPCs. Get the VPC associations wrong and your microservices can't find each other.
- **Alias records** are Route53's killer feature for AWS integrations—pointing your apex domain to CloudFront or ALB without the CNAME limitations. But only if you know when and how to use them.
- **Health checks and failover** can automatically route traffic away from failed endpoints. Or they can silently fail if you set up the wrong TTL or forget to attach the health check ID.

This guide explains the deployment landscape—from manual console work to full IaC automation—and why Planton defaults to **declarative configuration with Pulumi** for production-grade DNS management.

---

## The Maturity Spectrum: From Console Clicks to Infrastructure-as-Code

Managing Route53 zones and records requires solving three fundamental challenges:

1. **Consistency**: DNS records must match your desired state across environments. No "drift" from manual changes.
2. **Auditability**: You must know who changed what, when, and why. DNS outages are often traced to undocumented changes.
3. **Repeatability**: Creating a new environment (staging, DR, multi-region) should not require reconstructing DNS from memory or screenshots.

Different approaches solve these problems with varying degrees of success. Here's the progression from amateur to professional:

### Level 0: The Console Cowboy (Don't Do This)

**AWS Management Console** for DNS is the path of least initial resistance. Click "Create Hosted Zone," fill in some records, and you're done in minutes.

**Why It Breaks Down**:

- **No version control**: What was the MX record priority last week? Who changed the TTL from 300 to 60? Nobody knows.
- **No change review**: A junior engineer can delete the zone at 2 AM with zero oversight. It has happened.
- **No repeatability**: Want to spin up a new environment? Screenshot the console and retype everything. Hope you didn't miss the TXT record for SPF.
- **Common mistakes multiply**:
  - Forgetting to update the domain registrar with the Route53 name servers (zone exists but does nothing).
  - Trying to create a CNAME at the zone apex (DNS spec violation—use alias instead).
  - Setting a CNAME value without the trailing dot, accidentally creating `api.example.com.example.com`.
  - Hardcoding Elastic Load Balancer IPs in A records (they change without warning).

**Verdict**: Fine for learning Route53. Disastrous for production. The first DNS outage from an undocumented manual change will convince you to upgrade.

### Level 1: The Scripted Approach (CLI and SDKs)

**AWS CLI and SDKs** (Boto3, AWS SDK for Go, etc.) enable scripting DNS changes. You can write a shell script or Python program to create zones and upsert records.

**Example**:

```bash
aws route53 create-hosted-zone --name example.com --caller-reference $(date +%s)
aws route53 change-resource-record-sets --hosted-zone-id Z1ABC... --change-batch file://records.json
```

**What This Solves**:

- **Repeatability**: Run the script to recreate DNS in a new account.
- **Auditability** (if you commit scripts to Git): Changes are tracked in version control.

**What's Still Broken**:

- **Idempotency is your problem**: Run the script twice, and you might create duplicate records or fail with "already exists" errors. You must manually code UPSERT logic.
- **State management**: Did you already create that record? Is the TTL what the script expects? You have to query Route53 and diff it yourself.
- **JSON wrangling**: The `change-resource-record-sets` API requires precise JSON formatting. Trailing commas, missing quotes, or incorrect structure cause cryptic errors.
- **No dependency management**: If you alias to a CloudFront distribution, you must hardcode the CloudFront zone ID (`Z2FDTNDATAQYW2`) and hope AWS never changes it.

**Verdict**: Better than console, but you're rebuilding a poor version of what Infrastructure-as-Code tools already solved. Use this only for one-off migrations or dynamic updates from application code.

### Level 2: Configuration Management (Ansible, SaltStack)

**Ansible** and similar tools bring declarative DNS management with modules like `amazon.aws.route53`:

```yaml
- name: Create DNS record for app
  amazon.aws.route53:
    zone: "example.com"
    record: "app.example.com"
    type: "A"
    ttl: 300
    value: "203.0.113.45"
    state: present
```

**What This Solves**:

- **Declarative intent**: You describe the desired state. Ansible handles create vs. update logic.
- **Idempotency**: Running the playbook twice doesn't break anything (in theory).

**What's Still Limited**:

- **No real state tracking**: Ansible doesn't maintain a formal state file. It queries Route53 on every run. If the record drifts (someone changed it in the console), Ansible will overwrite it—but you won't know it drifted unless you're watching the logs.
- **No dependency graph**: Ansible runs tasks sequentially. If record B depends on record A, you must manually order them.
- **Pagination quirks**: Route53 zones with >100 records require pagination. Early Ansible versions didn't handle this well, leading to incomplete imports.

**Verdict**: Solid for teams already using Ansible for server configuration. But for pure infrastructure, Infrastructure-as-Code tools are more powerful.

### Level 3: The Production Standard (Infrastructure-as-Code)

**Terraform, Pulumi, CloudFormation, and CDK** treat infrastructure as first-class code with explicit state management, dependency graphs, and change previews.

This is the modern standard for production DNS. All four tools support Route53 comprehensively—creating zones, records, health checks, and even enabling DNSSEC.

**How IaC Solves DNS Problems**:

1. **Explicit State**: The tool knows exactly what it created. If someone manually changes a record, the next plan/preview will detect drift and either revert it or flag it for review.

2. **Dependency Management**: Declare an alias record pointing to an ALB. The tool automatically extracts the ALB's DNS name and zone ID. Update the ALB, and the DNS follows.

3. **Change Preview**: Before applying changes, you see a diff:
   ```
   ~ aws_route53_record.api
     ttl: 300 → 60
   ```
   This prevents "oops, I deleted the MX record" scenarios.

4. **Modular Reusability**: Define a DNS module. Instantiate it for dev, staging, and prod with different parameters. No copy-paste, no drift.

5. **Integration with AWS Features**:
   - **Alias records**: Automatically resolve CloudFront, ALB, S3 website endpoints without hardcoding zone IDs.
   - **Routing policies**: Weighted, latency-based, geolocation, failover—all supported with type-safe configuration.
   - **Health checks**: Attach health checks to failover records. The tool validates the health check exists before applying.

**Verdict**: This is the only approach that provides production-grade DNS management. The question isn't "if" but "which tool."

---

## The IaC Tool Landscape: Terraform, Pulumi, CloudFormation, CDK

All four major IaC tools handle Route53 well. The differences come down to philosophy, language, and ecosystem.

### Terraform: The Open-Source Standard

**Language**: HashiCorp Configuration Language (HCL)—declarative, purpose-built for infrastructure.

**Route53 Support**:
- Comprehensive: zones, records (all types including alias), health checks, delegation sets, DNSSEC.
- Strong community: Thousands of modules, extensive documentation, Stack Overflow answers.

**Example**:

```hcl
resource "aws_route53_zone" "main" {
  name = "example.com"
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.example.com"
  type    = "A"
  alias {
    name                   = aws_lb.app.dns_name
    zone_id                = aws_lb.app.zone_id
    evaluate_target_health = false
  }
}
```

**Strengths**:
- **Mature and stable**: Terraform has managed Route53 for nearly a decade.
- **Multi-cloud**: If you also use GCP Cloud DNS or Cloudflare, Terraform supports them all in one codebase.
- **State management**: Remote backends (S3, Terraform Cloud) enable team collaboration.

**Trade-Offs**:
- **HCL limitations**: Loops and conditionals are less intuitive than general-purpose programming languages.
- **State drift**: If AWS changes something Terraform doesn't track (like SOA serial numbers), you'll see noise in diffs.

**Best For**: Teams standardized on Terraform for multi-cloud infrastructure or those who prefer declarative DSLs over code.

### Pulumi: Infrastructure as Real Code

**Language**: TypeScript, Python, Go, C#, Java—your choice of real programming languages.

**Route53 Support**:
- Full parity with Terraform (uses the same underlying Terraform provider via "bridging").
- First-class TypeScript/Python support with IDE autocomplete, type checking, and refactoring tools.

**Example (TypeScript)**:

```typescript
import * as aws from "@pulumi/aws";

const zone = new aws.route53.Zone("main", { name: "example.com" });

const www = new aws.route53.Record("www", {
  zoneId: zone.zoneId,
  name: "www.example.com",
  type: "A",
  aliases: [{
    name: alb.dnsName,
    zoneId: alb.zoneId,
    evaluateTargetHealth: false,
  }],
});
```

**Strengths**:
- **Real programming languages**: Use loops, functions, classes, and libraries. Generate 100 similar DNS records with a simple `for` loop.
- **Type safety**: Catch configuration errors at compile time (wrong record type, missing required field).
- **Testable**: Write unit tests for your infrastructure code using standard test frameworks.
- **Rich ecosystem**: npm/PyPI packages, linters, formatters—leverage the entire software engineering toolchain.

**Trade-Offs**:
- **Learning curve**: Teams unfamiliar with TypeScript/Python may need ramp-up time.
- **State management**: Requires Pulumi Cloud (managed state) or self-hosted backend (S3, Azure Blob).

**Best For**: Teams that want infrastructure to feel like application code—strongly typed, testable, and expressive. Ideal if you're already using TypeScript or Python for application development.

### AWS CloudFormation: The AWS-Native Choice

**Language**: JSON or YAML templates—declarative, AWS-specific.

**Route53 Support**:
- Native AWS service. Supports zones, records (including alias), health checks, DNSSEC, Key Signing Keys (KSK).
- Deep integration: CloudFormation can output a zone's name servers as stack outputs for cross-stack references.

**Example (YAML)**:

```yaml
Resources:
  MyZone:
    Type: AWS::Route53::HostedZone
    Properties:
      Name: example.com

  WebRecord:
    Type: AWS::Route53::RecordSet
    Properties:
      HostedZoneId: !Ref MyZone
      Name: www.example.com
      Type: A
      AliasTarget:
        DNSName: !GetAtt MyALB.DNSName
        HostedZoneId: !GetAtt MyALB.CanonicalHostedZoneID
        EvaluateTargetHealth: false
```

**Strengths**:
- **No external tools**: If you're all-in on AWS, CloudFormation is built-in.
- **StackSets**: Deploy the same DNS config across multiple AWS accounts/regions with one command.
- **Change sets**: Preview changes before applying (similar to Terraform plan).

**Trade-Offs**:
- **YAML verbosity**: Large templates become unwieldy. No loops or functions (unless you layer AWS CDK on top).
- **Drift detection gaps**: Route53 RecordSet resources don't support CloudFormation's drift detection—manual changes won't be flagged automatically.
- **AWS-only**: Locked into AWS. Can't manage GCP or Azure DNS.

**Best For**: Organizations fully committed to AWS with governance requirements for AWS-native tooling. Common in enterprises with strict compliance policies.

### AWS CDK: CloudFormation with Programming Languages

**Language**: TypeScript, Python, Java, C#—synthesizes to CloudFormation under the hood.

**Route53 Support**:
- High-level constructs like `PublicHostedZone`, `ARecord`, `CnameRecord`.
- Automatic handling of alias targets: `RecordTarget.fromAlias(new CloudFrontTarget(distribution))`.

**Example (TypeScript)**:

```typescript
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as targets from 'aws-cdk-lib/aws-route53-targets';

const zone = new route53.PublicHostedZone(this, 'Zone', {
  zoneName: 'example.com',
});

new route53.ARecord(this, 'Alias', {
  zone,
  recordName: 'www',
  target: route53.RecordTarget.fromAlias(new targets.CloudFrontTarget(distribution)),
});
```

**Strengths**:
- **High-level abstractions**: CDK knows that CloudFront's zone ID is `Z2FDTNDATAQYW2` and fills it in automatically.
- **CloudFormation compatibility**: Deploy via CloudFormation, get drift detection (where supported), StackSets, etc.
- **Cross-stack references**: Easily reference DNS zones created in other stacks.

**Trade-Offs**:
- **CloudFormation limitations**: Inherits CloudFormation's lack of Traffic Policy support and drift detection gaps.
- **AWS-only**: Same vendor lock-in as CloudFormation.
- **Abstraction overhead**: Constructs can hide details. Debugging requires understanding both CDK and the generated CloudFormation.

**Best For**: AWS-native teams that want the benefits of programming languages without abandoning CloudFormation tooling and governance.

---

## Advanced Tools: Crossplane, ExternalDNS, and DNS-as-Code Platforms

Beyond general-purpose IaC, specialized tools solve specific DNS automation problems.

### Crossplane: Kubernetes-Native AWS Management

**What It Does**: Manage Route53 zones and records using Kubernetes Custom Resources.

**Example**:

```yaml
apiVersion: route53.aws.crossplane.io/v1alpha1
kind: HostedZone
metadata:
  name: example-zone
spec:
  forProvider:
    name: example.com
```

**When to Use**: If you're running Kubernetes and want DNS to be part of your Kubernetes control plane. Useful for platform teams building internal self-service portals where developers request DNS via kubectl.

**Trade-Offs**: Adds operational complexity (running Crossplane controllers). Overkill if you're not already deep in Kubernetes-based infrastructure.

### ExternalDNS: Automatic DNS for Kubernetes Services

**What It Does**: Watches Kubernetes Services and Ingresses, automatically creates Route53 records pointing to them.

**Example**:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
```

ExternalDNS sees this annotation and creates an A record for `app.example.com` pointing to the Service's load balancer.

**When to Use**: Dynamic environments where DNS records should appear/disappear as applications deploy. Common in multi-tenant platforms or ephemeral preview environments.

**Trade-Offs**: Records are managed outside your IaC (ExternalDNS owns them). If you also manage some records via Terraform, you must carefully scope ownership to avoid conflicts.

### OctoDNS and DNSControl: Multi-Provider DNS-as-Code

**What They Do**: Unified configuration format for DNS across multiple providers (Route53, Cloudflare, Google Cloud DNS, etc.).

**Example (OctoDNS YAML)**:

```yaml
example.com.:
  - type: A
    value: 192.0.2.1
  - type: MX
    value: 10 mail.example.com.
```

Run `octodns-sync`, and it applies changes to all configured providers.

**When to Use**: Multi-cloud DNS strategies or migrations between providers. Also useful for keeping DNS in sync across Route53 (public internet) and another DNS system (private enterprise network).

**Trade-Offs**: Another tool to learn and maintain. Less AWS-specific than Terraform/Pulumi (no automatic alias zone ID handling).

---

## What Planton Supports: Pulumi for Production DNS

Planton defaults to **Pulumi** for managing AWS Route53 zones and records.

**Why Pulumi?**

1. **Real Code, Real Benefits**: DNS configurations often need logic—loops for creating multiple similar records, conditionals for environment-specific settings, helper functions for normalizing domain names. With Pulumi's TypeScript or Python, this is natural. With HCL or YAML, it's painful.

2. **Type Safety and IDE Support**: Catch errors before deployment. Your editor autocompletes Route53 record types, warns if you set an alias and a TTL (invalid), and highlights missing required fields.

3. **Testable Infrastructure**: Write unit tests to verify your DNS module creates the correct number of records or validates domain name formats. This is standard practice in application code—it should be standard in infrastructure code too.

4. **Multi-Cloud Ready**: While Route53 is AWS-specific, many organizations also use GCP Cloud DNS or Cloudflare. Pulumi handles all of them in one codebase with a unified programming model.

5. **Open Source and Vendor-Neutral**: Pulumi is fully open source (Apache 2.0). The state backend can be self-hosted (S3, Azure Blob, local filesystem) or use Pulumi Cloud. No vendor lock-in on the tooling layer.

**What We Abstract Away**:

- **Boilerplate**: Our Pulumi modules handle common patterns—creating a public zone with default TTLs, configuring alias records for CloudFront/ALB without hardcoding zone IDs, attaching health checks to failover records.
- **Best Practices**: We enforce trailing dots where required, validate CNAME targets, and prevent common mistakes like setting TTL on alias records.
- **Multi-Environment Management**: Define DNS once, deploy to dev/staging/prod with environment-specific overrides (like different ALB targets or TTL values).

**What You Control**:

- **Record Definitions**: You specify the DNS records you need—A, AAAA, CNAME, MX, TXT, alias, routing policies.
- **Health Checks and Failover**: Configure health checks and attach them to failover policies as needed.
- **Private Zones and VPC Associations**: Declare private zones with explicit VPC associations for split-horizon DNS.

---

## The 80/20 Rule: What Most Teams Actually Need

Route53 is feature-rich, but production DNS for most applications uses a **small, predictable subset**:

### Essential Record Types (>80% of Use Cases)

1. **A/AAAA Records**: Map domain names to IPv4/IPv6 addresses. Often replaced by alias records for AWS resources.
2. **Alias Records**: Route53's killer feature—point your apex domain (`example.com`) to CloudFront, ALB, S3 website, or API Gateway without CNAME restrictions. Also free (no query charges for alias queries to AWS resources).
3. **CNAME Records**: Alias subdomains to external services (`support.example.com` → `yourcompany.zendesk.com`).
4. **MX Records**: Email routing. Typically 2-3 records with priorities (`10 mail1.provider.com`, `20 mail2.provider.com`).
5. **TXT Records**: SPF for email validation, domain verification tokens (Google, AWS, etc.), DKIM keys.

### Common Advanced Patterns (Next 15%)

- **Weighted Routing**: Blue/green deployments (90% traffic to old version, 10% to new).
- **Latency-Based Routing**: Global applications serving users from the nearest region.
- **Failover Routing**: Primary/secondary setup with health checks for automatic disaster recovery.
- **Geolocation Routing**: Serve different content based on user location (e.g., GDPR-compliant EU endpoints).

### Rarely Used Features (<5%)

- **Geoproximity and IP-Based Routing**: Niche use cases requiring fine-grained control over traffic distribution.
- **DNSSEC**: Security enhancement, but adoption is low due to complexity and limited client support.
- **Traffic Flow Policies**: Visual policy editor. Powerful but lacks Terraform/Pulumi support, making it unsuitable for IaC workflows.
- **CAA Records**: Restrict which Certificate Authorities can issue certificates for your domain. Good security practice but not universally adopted.
- **SRV Records**: Mainly for legacy services like Active Directory or SIP. Modern cloud-native apps rarely use them.

**Planton's API focuses on the 80%**: zones (public/private), core record types, alias targets, and basic routing policies. Advanced features can be added as needed, but the API remains simple by default.

---

## Production Essentials: What You Must Get Right

### Public vs. Private Zones

- **Public Zones**: Resolve globally on the internet. After creating a zone, you must update your domain registrar with the Route53 name servers. Forget this, and your zone exists but does nothing.
- **Private Zones**: Resolve only within associated VPCs. Critical for internal microservices (`postgres.internal.example.com`). You must enable `enableDnsHostnames` and `enableDnsSupport` in your VPC settings or Route53 resolution won't work.

### Split-Horizon DNS

You can have both a public and private zone for the same domain. Route53 prioritizes the private zone for queries from associated VPCs, falling back to the public zone if no match.

**Use Case**: `api.example.com` points to an internal ALB for VPC traffic (private zone) but resolves to a public CloudFront distribution for internet users (public zone).

### Alias Records: When and Why

Use alias instead of CNAME for:
- **Zone apex**: DNS spec prohibits CNAME at `example.com`. Alias works.
- **AWS targets**: CloudFront, ALB/NLB, S3 website, API Gateway, Global Accelerator. Alias queries to AWS resources are free. CNAME queries cost money.

**Gotcha**: You must know the target's hosted zone ID. For CloudFront, it's always `Z2FDTNDATAQYW2`. For ALB/NLB, it varies by region. Pulumi and CDK handle this automatically. Raw CloudFormation or Terraform requires manual lookup.

### TTL Strategy

- **Default 300 seconds (5 minutes)**: Balances caching efficiency with change propagation speed.
- **Low TTL (60 seconds)**: For records you might change during an incident (failover targets, blue/green weights). Enables faster cutover.
- **High TTL (86400 seconds / 1 day)**: For static records (MX, NS for delegated subdomains) to reduce query costs.

**Pro Tip**: Lower TTL a day before planned changes. After changes are verified, raise it back to reduce query volume.

### Health Checks and Failover

Route53 health checks monitor endpoints (HTTP/HTTPS, TCP, or even CloudWatch alarms). Attach them to failover or multivalue records:

- **Failover**: Primary record with health check. If it fails, Route53 automatically returns the secondary.
- **Multivalue**: Return up to 8 healthy records per query. Clients retry unhealthy ones.

**Cost**: ~$0.50/month per health check for AWS endpoints, ~$0.75/month for external endpoints. First 50 health checks on AWS resources are free.

### Query Logging and Monitoring

Enable **Route53 Query Logging** to send DNS query logs to CloudWatch Logs. Use this to:
- Debug resolution issues ("why isn't my new record resolving?").
- Detect anomalous query patterns (security monitoring).
- Understand query volume for cost optimization.

**Warning**: High-traffic domains generate massive logs. Set CloudWatch Logs retention policies to avoid surprise bills.

---

## Common Mistakes to Avoid

1. **Hardcoding ALB/ELB IPs**: Load balancer IPs change. Always use alias records or CNAME to the DNS name.
2. **Forgetting the Trailing Dot**: Route53 is forgiving, but exporting to BIND or using other DNS tools may not be. `mail.example.com` vs. `mail.example.com.`—one is relative, one is absolute.
3. **CNAME at Apex**: Invalid DNS. Use alias instead.
4. **Ignoring TTL During Changes**: If you update a record with TTL=86400, clients may cache the old value for a day. Plan ahead or accept slow propagation.
5. **Mixing Manual and IaC Changes**: If you manage DNS with Terraform but someone manually adds a record in the console, the next Terraform run will delete it (or fail with a conflict). Decide on one source of truth.
6. **Not Monitoring Health Checks**: A failing health check means Route53 shifted traffic. Set up CloudWatch alarms so you know before customers complain.

---

## Conclusion: DNS is Infrastructure, Treat It Like Code

AWS Route53 is the DNS service. The question is not "which DNS provider" but "how do you manage it."

**Manual console work** might feel fast for the first zone, but it creates technical debt that compounds. Undocumented changes, configuration drift, and the inability to reproduce environments will eventually cause outages.

**Declarative Infrastructure-as-Code** with Pulumi gives you version control, change previews, repeatable deployments, and the ability to test your DNS configurations before they go live. It's the only approach that scales to production.

Planton's Route53 modules abstract the boilerplate while giving you full control over records, routing policies, and health checks. Your DNS becomes part of your GitOps workflow—reviewed in pull requests, tested in CI, and deployed with confidence.

DNS is foundational. When it works, nobody notices. When it breaks, everything breaks.

Treat it like the critical infrastructure it is.

