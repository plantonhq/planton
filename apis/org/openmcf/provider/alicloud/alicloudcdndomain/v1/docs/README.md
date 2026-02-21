# AliCloudCdnDomain — Research & Design Document

## 1 Introduction

This document captures the research and design rationale behind the
**AliCloudCdnDomain** OpenMCF component (R26). The component provides a
declarative, Kubernetes-style manifest for managing an accelerated domain
within Alibaba Cloud CDN.

Alibaba Cloud CDN is a globally distributed content distribution network.
Operators register an accelerated *domain name* (e.g., `cdn.example.com`) and
attach one or more *origin sources*. Edge nodes closest to end users cache
responses from the origins and serve them with lower latency, higher
throughput, and reduced load on the origin infrastructure.

After the CDN service accepts the domain, Alibaba Cloud returns a `cname`
value. The operator must create a DNS CNAME record pointing the accelerated
domain to that value before edge acceleration takes effect.

The AliCloudCdnDomain component wraps this lifecycle into a single manifest
that can be version-controlled, reviewed, and deployed through CI/CD.

### 1.1 Terminology

| Term | Meaning |
|------|---------|
| **Accelerated domain** | The user-facing domain name registered in CDN (e.g., `cdn.example.com`). |
| **Origin source** | The backend server that CDN edge nodes pull content from. |
| **CNAME** | The DNS alias issued by CDN that routes traffic through edge nodes. |
| **CDN type** | The content profile (`web`, `download`, `video`) that tells CDN how to optimize caching. |
| **Scope** | The geographic reach of edge acceleration (`domestic`, `overseas`, `global`). |
| **CAS** | Alibaba Cloud Certificate Management Service, used for HTTPS certificates. |

### 1.2 Scope of This Document

This document covers:

- How CDN domain configuration is typically done manually and with IaC tools.
- The production architecture concerns: origin configuration, HTTPS, caching.
- What the OpenMCF component covers and why it makes certain trade-offs.

---

## 2 Deployment Landscape

This section surveys the ways operators provision CDN domains on Alibaba Cloud,
from manual workflows to infrastructure-as-code.

### 2.1 Manual Console Configuration

The Alibaba Cloud console provides a step-by-step wizard:

1. Navigate to the CDN console and click "Add Domain Name."
2. Enter the accelerated domain name and select the CDN type.
3. Choose a geographic scope.
4. Add one or more origin sources with type, address, port, priority, and weight.
5. Optionally configure HTTPS by uploading a certificate or selecting one from CAS.
6. Submit and wait for domain verification.
7. Copy the returned CNAME and create a DNS record.

Manual configuration is suitable for one-off experiments but introduces several
risks:

- No audit trail for configuration changes.
- Inconsistency when multiple operators configure domains independently.
- HTTPS certificates can expire without automated rotation workflows.
- No automated rollback path if a misconfiguration degrades service.

### 2.2 Alibaba Cloud CLI

The `aliyun` CLI offers imperative commands:

```shell
# Add a CDN domain
aliyun cdn AddCdnDomain \
  --DomainName cdn.example.com \
  --CdnType web \
  --Sources '[{"content":"origin.example.com","type":"domain","port":443,"priority":"20"}]'

# Check the domain status
aliyun cdn DescribeCdnDomainDetail --DomainName cdn.example.com

# Enable HTTPS
aliyun cdn SetCdnDomainSSLCertificate \
  --DomainName cdn.example.com \
  --SSLProtocol on \
  --CertType cas \
  --CertId cas-cn-abc123 \
  --CertRegion cn-hangzhou
```

CLI usage is scriptable but still imperative. Scripts accumulate ad-hoc
conditionals, error handling becomes inconsistent, and there is no built-in
state tracking to detect drift between the intended and actual configuration.

### 2.3 Terraform

The `alicloud_cdn_domain_new` resource in the `aliyun/alicloud` Terraform
provider is the primary IaC path:

```hcl
resource "alicloud_cdn_domain_new" "main" {
  domain_name = "cdn.example.com"
  cdn_type    = "web"
  scope       = "domestic"

  sources {
    type     = "domain"
    content  = "origin.example.com"
    port     = 443
    priority = 20
    weight   = 10
  }

  certificate_config {
    cert_type                 = "cas"
    cert_id                   = "cas-cn-abc123"
    cert_region               = "cn-hangzhou"
    server_certificate_status = "on"
  }

  tags = {
    team = "platform"
  }
}

output "cname" {
  value = alicloud_cdn_domain_new.main.cname
}
```

Key characteristics of the Terraform approach:

- **State file** tracks the current configuration and allows plan/apply diffs.
- **Import** (`terraform import`) can adopt existing CDN domains.
- **Lifecycle meta-arguments** can prevent accidental destruction.
- **Dynamic blocks** support variable numbers of origin sources.
- HTTPS certificate fields are part of the same resource, so cert rotation is
  a normal `apply` cycle.

The provider version `~> 1.200` is the minimum recommended for the
`alicloud_cdn_domain_new` resource to include all certificate_config
sub-arguments.

#### 2.3.1 Terraform Caveats

| Area | Detail |
|------|--------|
| ForceNew fields | `domain_name` and `cdn_type` trigger destroy+recreate. |
| Certificate rotation | Changing `cert_id` is an in-place update; no downtime. |
| Scope restrictions | Domains scoped to `domestic` require ICP filing. |
| Tag propagation | Tags are metadata-only; they do not affect CDN behavior. |

### 2.4 Pulumi

The `pulumi-alicloud` SDK provides the `cdn.DomainNew` resource:

```go
domain, err := cdn.NewDomainNew(ctx, "cdn.example.com", &cdn.DomainNewArgs{
    DomainName: pulumi.String("cdn.example.com"),
    CdnType:    pulumi.String("web"),
    Scope:      pulumi.String("domestic"),
    Sources: cdn.DomainNewSourceArray{
        cdn.DomainNewSourceArgs{
            Type:    pulumi.String("domain"),
            Content: pulumi.String("origin.example.com"),
            Port:    pulumi.Int(443),
            Priority: pulumi.Int(20),
        },
    },
    CertificateConfig: &cdn.DomainNewCertificateConfigArgs{
        CertType:                 pulumi.String("cas"),
        CertId:                   pulumi.String("cas-cn-abc123"),
        CertRegion:               pulumi.String("cn-hangzhou"),
        ServerCertificateStatus:  pulumi.String("on"),
    },
    Tags: pulumi.StringMap{
        "team": pulumi.String("platform"),
    },
}, pulumi.Provider(alicloudProvider))
```

Pulumi mirrors Terraform's resource model for CDN domains. The Go SDK used by
OpenMCF modules provides compile-time type safety on argument names and
output access. State management, diff previews, and stack references all work
the same as any other Pulumi resource.

### 2.5 Comparison Table

| Criterion | Console | CLI | Terraform | Pulumi |
|-----------|---------|-----|-----------|--------|
| Audit trail | Manual screenshots | Script logs | State file + VCS | State file + VCS |
| Drift detection | None | None | `terraform plan` | `pulumi preview` |
| Multi-origin failover | Click-through wizard | JSON in CLI args | `dynamic "sources"` | `DomainNewSourceArray` |
| HTTPS setup | Certificate picker | Separate API call | `certificate_config` block | `CertificateConfigArgs` |
| Reusability | None | Shell functions | Modules | Component resources |
| Rollback | Manual revert | Re-run script | `terraform apply` previous state | `pulumi up` previous commit |

---

## 3 Production Architecture

### 3.1 Origin Configuration

CDN domains support multiple origin sources for high availability and load
distribution. Two configuration patterns are common in production:

**Active/Standby (Priority-Based Failover)**

| Source | Type | Priority | Weight | Role |
|--------|------|----------|--------|------|
| `origin-a.example.com` | domain | 20 | 10 | Primary |
| `origin-b.example.com` | domain | 30 | 10 | Standby |

Edge nodes send all traffic to the priority-20 source. If it becomes
unreachable, CDN fails over to the priority-30 source. This is the recommended
pattern for applications with a primary region and a DR region.

**Weighted Round-Robin (Same Priority)**

| Source | Type | Priority | Weight | Role |
|--------|------|----------|--------|------|
| `origin-a.example.com` | domain | 20 | 60 | 60% traffic |
| `origin-b.example.com` | domain | 20 | 40 | 40% traffic |

When multiple sources share the same priority, CDN distributes requests by
weight. This pattern is used for load balancing across co-equal origins.

**OSS as Origin**

Using an OSS bucket as the origin is the standard pattern for static websites
and asset hosting. The source type is `oss` and the content field uses the
bucket's public domain (e.g., `my-bucket.oss-cn-hangzhou.aliyuncs.com`). The
`checkUrl` field should point to a known-good object in the bucket for origin
health verification during domain creation.

### 3.2 HTTPS and Certificates

CDN edge nodes terminate TLS on behalf of the origin. Three certificate
strategies are available:

| Strategy | `certType` | When to Use |
|----------|-----------|-------------|
| CAS-managed | `cas` | Certificate is stored in Alibaba Cloud Certificate Management Service. |
| Upload | `upload` | PEM cert and key are provided directly in the manifest. |
| Free DV | `free` | Alibaba Cloud issues a free Domain-Validated certificate. |

**CAS-managed certificates** are the recommended approach for production:

- Certificates are managed centrally in CAS and referenced by `certId`.
- CAS supports automatic renewal for certificates it manages.
- Certificate rotation is a manifest change (`certId` value), applied with no
  edge downtime.
- `certRegion` should be `cn-hangzhou` for domains scoped to `domestic` and
  `ap-southeast-1` for `overseas` or `global` scopes.

**Uploaded certificates** are used when the certificate is managed externally
(e.g., Let's Encrypt, internal CA). The PEM content is provided in
`serverCertificate` and `privateKey`. This requires the operator to manage
certificate renewal and redeploy when the certificate changes.

**Free DV certificates** are limited to single-domain coverage and do not
support wildcard or multi-domain. They are suitable for non-critical or
development CDN domains.

The `serverCertificateStatus` field defaults to `on` when a
`certificateConfig` block is present, enabling HTTPS on the edge. Setting it
to `off` disables HTTPS while preserving the certificate association.

### 3.3 Caching Behavior

Alibaba Cloud CDN applies default caching rules based on file extensions and
HTTP cache headers from the origin. Operators commonly customize caching via
CDN domain config rules:

- File-extension-based TTLs (e.g., `.jpg` cached for 30 days, `.html` for 5 minutes).
- Directory-based rules (e.g., `/api/` paths not cached).
- Query-string handling (include or ignore query parameters in cache keys).
- HTTP header additions (e.g., `Access-Control-Allow-Origin`).

These CDN function configurations are *not* part of the `alicloud_cdn_domain_new`
resource. They require separate Terraform resources
(`alicloud_cdn_domain_config`) or Pulumi resources. This is by design — the
domain resource manages the domain lifecycle and origin routing, while function
configs manage edge behavior.

### 3.4 DNS Integration

After CDN registers the domain and returns a CNAME, the operator must create
a DNS CNAME record. Without this record, traffic continues to flow directly to
the origin, bypassing CDN entirely.

Typical workflow:

1. Deploy the AliCloudCdnDomain manifest; obtain the `cname` output.
2. Create a CNAME record: `cdn.example.com → cdn.example.com.w.cdngslb.com`.
3. Wait for DNS propagation (minutes to hours depending on TTL).
4. Verify with `dig cdn.example.com CNAME` or `nslookup`.

If the domain is managed in Alibaba Cloud DNS, an AliCloudDnsRecord component
can reference the CDN domain's `cname` output to automate step 2.

### 3.5 ICP Filing Requirements

For domains with `scope` set to `domestic` or `global`, the accelerated domain
must have a valid ICP (Internet Content Provider) filing with the Chinese
Ministry of Industry and Information Technology. Domains without an ICP filing
will fail CDN verification. This requirement is enforced by the CDN service,
not by the IaC tool.

### 3.6 Resource Groups

The `resourceGroupId` field associates the CDN domain with an Alibaba Cloud
resource group for access control (RAM policies scoped to a resource group)
and cost tracking. If omitted, the domain is placed in the account's default
resource group.

---

## 4 Best Practices

| # | Practice | Rationale |
|---|----------|-----------|
| 1 | Use CAS-managed certificates for HTTPS | Centralized renewal; no PEM in manifests. |
| 2 | Set up active/standby origins for production | Automatic failover without manual intervention. |
| 3 | Use OSS origin for static assets | Lower cost, higher throughput, no server maintenance. |
| 4 | Set `scope` explicitly | Avoids surprise ICP requirements from default `domestic`. |
| 5 | Use `checkUrl` for OSS origins | Validates origin reachability during domain creation. |
| 6 | Apply resource group IDs | Enables per-team cost attribution and scoped IAM policies. |
| 7 | Tag every domain | Tags propagate to billing reports and resource inventories. |
| 8 | Separate CDN function configs from domain lifecycle | Domain changes are infrequent; caching rules change often. |
| 9 | Automate DNS CNAME creation | Reference the `cname` output in an AliCloudDnsRecord component. |
| 10 | Store manifests in version control | Enables code review, audit trail, and rollback. |
| 11 | Use `priority` 20/30 convention | Consistent primary/standby distinction across all CDN domains. |
| 12 | Monitor domain status via outputs | `status` output exposes `online`, `offline`, `configuring`, `check_failed`. |

---

## 5 What OpenMCF Supports

### 5.1 The 80/20 Design Principle

The AliCloudCdnDomain component covers the 80% of CDN domain configuration
that nearly every deployment needs:

- **Domain registration** — accelerated domain name, CDN type, geographic scope.
- **Origin configuration** — one or more sources with type, address, port,
  priority, and weight.
- **HTTPS certificates** — CAS-managed, uploaded, or free DV certificates
  with configurable certificate status.
- **Operational metadata** — tags, resource group, health check URL.

The remaining 20% — per-path caching rules, referer whitelists, HTTP header
injection, IP blacklists, URL signing, WebSocket support, HTTP/2 settings,
and real-time log delivery — is intentionally outside the component scope.
These features vary widely across deployments and change at different cadences
than the domain lifecycle. They are better managed through raw Terraform
resources (`alicloud_cdn_domain_config`) or dedicated OpenMCF components if
the pattern stabilizes.

### 5.2 Field Coverage

The following table maps Alibaba Cloud CDN domain fields to their OpenMCF
coverage status:

| CDN API Field | OpenMCF Field | Covered |
|--------------|---------------|---------|
| DomainName | `domainName` | Yes |
| CdnType | `cdnType` | Yes |
| Scope | `scope` | Yes |
| Sources | `sources` | Yes |
| CertificateConfig | `certificateConfig` | Yes |
| CheckUrl | `checkUrl` | Yes |
| ResourceGroupId | `resourceGroupId` | Yes |
| Tags | `tags` | Yes |
| TopLevelDomain | — | No (rarely used, mainland China specific) |
| DomainConfig functions | — | No (separate resource) |
| RealTimeLogDelivery | — | No (separate resource) |
| WAF integration | — | No (separate resource) |

### 5.3 Foreign Keys

The component accepts foreign key references to other Alibaba Cloud resources:

| Field | References | Purpose |
|-------|-----------|---------|
| `resourceGroupId` | Resource Group ID | Cost attribution, IAM scope |
| `certificateConfig.certId` | CAS Certificate ID | HTTPS termination |
| `sources[].content` | Origin IP, domain, or OSS bucket | Content origin |

These are string-valued identifiers. OpenMCF does not validate their existence
at plan time — the CDN API performs that validation during deployment.

### 5.4 Implementation Landscape

The component is implemented as both a Pulumi module and a Terraform module,
using the same protobuf-defined manifest schema.

#### Pulumi Module

| File | Purpose |
|------|---------|
| `main.go` (entrypoint) | Loads stack input, calls `module.Resources()`. |
| `module/locals.go` | Computes tags from metadata (name, id, org, env) merged with user tags. |
| `module/main.go` | Creates the alicloud provider and `cdn.DomainNew` resource. |
| `module/outputs.go` | Defines output constants: `domain_name`, `cname`, `status`. |

The Pulumi module creates a single `cdn.DomainNew` resource with:

- An explicit alicloud provider scoped to `spec.region`.
- Dynamic source list construction from the spec's repeated `sources` field.
- Optional `certificateConfig` when the spec includes certificate configuration.
- Tags derived from metadata fields plus user-defined tags.

#### Terraform Module

| File | Purpose |
|------|---------|
| `provider.tf` | Configures the `aliyun/alicloud` provider with the spec's region. |
| `variables.tf` | Defines `metadata` and `spec` input variables with validation. |
| `locals.tf` | Computes final tags from metadata and spec. |
| `main.tf` | Creates `alicloud_cdn_domain_new` with dynamic source and certificate blocks. |
| `outputs.tf` | Exports `domain_name`, `cname`, and `status`. |

Both modules produce identical outputs for the same input manifest, ensuring
operators can switch provisioners without changing their manifest files.

### 5.5 Outputs

| Output | Source | Description |
|--------|--------|-------------|
| `domain_name` | `alicloud_cdn_domain_new.main.domain_name` | The registered accelerated domain name. |
| `cname` | `alicloud_cdn_domain_new.main.cname` | The CNAME value for DNS configuration. |
| `status` | `alicloud_cdn_domain_new.main.status` | Current domain status (`online`, `offline`, `configuring`, etc.). |

The `cname` output is the critical integration point — it must be used to
create a DNS CNAME record before CDN acceleration is active.

### 5.6 Validation

The protobuf schema enforces validation rules at deserialization time:

| Field | Rule |
|-------|------|
| `region` | Required, non-empty string. |
| `domainName` | Required, 1-63 characters. |
| `cdnType` | Required, one of `web`, `download`, `video`. |
| `scope` | Optional, one of `domestic`, `overseas`, `global`. |
| `sources` | At least one item required. |
| `sources[].type` | Required, one of `ipaddr`, `domain`, `oss`, `common`. |
| `sources[].content` | Required, non-empty string. |
| `sources[].priority` | 0-100 range. |
| `sources[].weight` | 0-100 range. |
| `certificateConfig.certType` | Optional, one of `upload`, `cas`, `free`. |
| `certificateConfig.serverCertificateStatus` | Optional, one of `on`, `off`. |

The Terraform module adds equivalent `validation` blocks on the `spec`
variable for `cdn_type`, `scope`, and source count.

---

## 6 Design Decisions

### 6.1 Single Resource, No Domain Configs

The component creates one `alicloud_cdn_domain_new` resource. CDN function
configurations (caching rules, headers, redirects) are separate API resources
in both Terraform and Pulumi. Bundling them into this component would create a
monolithic resource with high change frequency and complex diff semantics.

### 6.2 Certificate Config as Inline Block

The certificate configuration is an inline block on the domain resource rather
than a separate resource. This matches the Alibaba Cloud provider's data model
where `certificate_config` is a nested attribute of `alicloud_cdn_domain_new`.
An alternative design using `alicloud_cdn_domain_config` for HTTPS was
rejected because it would split the HTTPS lifecycle from the domain lifecycle
and introduce ordering dependencies.

### 6.3 Explicit Provider per Deployment

Both the Pulumi and Terraform modules create an alicloud provider scoped to
`spec.region`. CDN is a global service, but the provider requires a region for
API endpoint routing. Using an explicit provider avoids reliance on ambient
environment configuration and ensures the module is self-contained.

### 6.4 Tag Computation in Locals

System tags (`resource`, `resource_name`, `resource_kind`, `resource_id`,
`organization`, `environment`) are computed from metadata and merged with
user-provided `spec.tags`. User tags override system tags with the same key.
This convention is consistent across all OpenMCF Alibaba Cloud components.

### 6.5 ForceNew Awareness

The `domainName` and `cdnType` fields trigger a destroy-and-recreate cycle if
changed. The component documents these fields as immutable after creation.
Operators who need to change these values must accept the service disruption
or create a new domain and migrate DNS.

---

## 7 Operational Considerations

### 7.1 Domain Status Lifecycle

After creation, the CDN domain transitions through several states:

```
configuring → online
configuring → check_failed (if origin health check fails)
online → offline (if manually stopped or ICP revoked)
```

The `status` stack output reflects the current state. Operators should monitor
this output in CI/CD pipelines and alert on `check_failed` or unexpected
`offline` transitions.

### 7.2 Propagation Timing

CDN domain configuration changes (scope, sources, certificates) take effect
at edge nodes within 5-30 minutes after the API accepts the change. This is
an Alibaba Cloud platform behavior, not an IaC limitation. During propagation,
some edge nodes may serve content with the old configuration.

### 7.3 Certificate Rotation

For CAS-managed certificates, rotation is a manifest update:

1. Upload or renew the certificate in CAS.
2. Update `certificateConfig.certId` in the manifest.
3. Deploy the updated manifest.

CDN performs a rolling certificate update across edge nodes. There is no
downtime during rotation. For uploaded certificates, the same flow applies
but the `serverCertificate` and `privateKey` fields are updated instead.

### 7.4 Cost Model

Alibaba Cloud CDN billing is based on:

- **Bandwidth** — peak bandwidth per region per billing cycle.
- **Traffic** — total data transferred from edge nodes.
- **HTTPS requests** — additional charge for TLS handshakes.
- **Value-added features** — real-time logging, WAF, QUIC.

The component itself incurs no additional cost beyond the underlying CDN
usage. Tags can be used to allocate CDN costs to teams or projects in billing
reports.

---

## 8 Testing and Verification

### 8.1 Pre-Deployment Checks

- Validate the manifest schema: `openmcf validate -f manifest.yaml`.
- Preview the deployment: `openmcf pulumi preview` or `openmcf tofu plan`.
- Verify ICP filing status for `domestic` or `global` scope.

### 8.2 Post-Deployment Checks

```shell
# Check domain status
aliyun cdn DescribeCdnDomainDetail --DomainName cdn.example.com

# Verify CNAME is assigned
dig cdn.example.com CNAME

# Test edge caching
curl -I https://cdn.example.com/index.html
# Look for: X-Cache: HIT, Age: > 0

# Verify HTTPS certificate
openssl s_client -connect cdn.example.com:443 -servername cdn.example.com </dev/null 2>/dev/null | openssl x509 -noout -dates
```

### 8.3 Drift Detection

Both Pulumi and Terraform detect drift between the manifest and the actual
CDN configuration:

- `openmcf pulumi refresh` / `openmcf tofu plan` will show differences.
- Common drift sources: manual console changes, CDN service auto-updates.

---

## 9 Migration Paths

### 9.1 Importing Existing CDN Domains

Both provisioners support importing existing CDN domains into state:

**Terraform:**
```shell
openmcf tofu import alicloud_cdn_domain_new.main cdn.example.com
```

**Pulumi:**
```shell
pulumi import alicloud:cdn/domainNew:DomainNew cdn.example.com cdn.example.com
```

After import, align the manifest fields with the actual CDN configuration to
avoid an immediate update on the next deployment.

### 9.2 Migrating Between Provisioners

Because both the Pulumi and Terraform modules consume the same protobuf
manifest, migrating between provisioners requires:

1. Destroy the stack in the current provisioner (or remove from state).
2. Import the CDN domain into the new provisioner's state.
3. Change the provisioner label in the manifest metadata.
4. Deploy with the new provisioner.

The CDN domain itself is not affected — only the IaC state changes.

---

## 10 Comparison with Other Cloud Providers

| Feature | Alibaba Cloud CDN | AWS CloudFront | Azure CDN | GCP Cloud CDN |
|---------|------------------|----------------|-----------|---------------|
| Origin types | IP, domain, OSS, common | S3, ALB, custom origin | Storage, App Service, custom | GCS, backend service |
| HTTPS | CAS, upload, free DV | ACM, custom | Managed, custom | Google-managed, custom |
| Geographic scope | domestic/overseas/global | Price class selection | Standard/Premium tiers | Global by default |
| Cache config | Separate domain config resource | Distribution behaviors | Rule engine | Backend service config |
| ICP requirement | Yes (domestic/global) | No | No | No |

---

## 11 Conclusion

The AliCloudCdnDomain component provides a focused, declarative interface for
the most common CDN domain operations on Alibaba Cloud: registering an
accelerated domain, configuring origin sources with failover, enabling HTTPS
with certificate management, and applying operational metadata.

By limiting scope to the domain resource and its directly nested certificate
configuration, the component avoids the complexity of edge behavior rules
(caching, headers, redirects) that change at a different cadence and are better
managed as separate resources or components.

The dual Pulumi/Terraform implementation ensures operators can use their
preferred provisioner while sharing the same manifest format, validation
rules, and output contract.

Key integration points for production deployments:

- Use the `cname` output to automate DNS record creation.
- Reference CAS certificate IDs for HTTPS with automated renewal.
- Pair with AliCloudDnsRecord and AliCloudStorageBucket components for
  complete static site delivery architectures.

---

## References

- [Alibaba Cloud CDN Documentation](https://www.alibabacloud.com/help/en/cdn/)
- [Terraform alicloud_cdn_domain_new](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/cdn_domain_new)
- [Pulumi alicloud cdn.DomainNew](https://www.pulumi.com/registry/packages/alicloud/api-docs/cdn/domainnew/)
- [Alibaba Cloud CAS Documentation](https://www.alibabacloud.com/help/en/certificate-management-service/)
- [ICP Filing Overview](https://www.alibabacloud.com/help/en/icp-filing/)
