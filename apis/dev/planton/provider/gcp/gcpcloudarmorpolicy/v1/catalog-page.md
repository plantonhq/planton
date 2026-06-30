# GCP Cloud Armor Policy

Deploys a Google Cloud Armor security policy with inline rules for WAF protection, DDoS defense, and traffic filtering. The component supports IP-based allowlists/denylists, CEL expression matching, rate limiting with ban escalation, redirect actions, custom header injection, and preconfigured WAF rule exclusions.

## What Gets Created

When you deploy a GcpCloudArmorPolicy resource, Planton provisions:

- **Security Policy** — a `google_compute_security_policy` with the specified type (CLOUD_ARMOR, CLOUD_ARMOR_EDGE, or CLOUD_ARMOR_INTERNAL_SERVICE)
- **Inline Security Rules** — each rule in the spec becomes an inline rule on the policy, evaluated in priority order (lowest number first)
- **Adaptive Protection** — optional Layer 7 DDoS defense configuration on the policy
- **Advanced Options** — optional JSON body parsing, logging verbosity, and client IP resolution settings

If no rule with priority 2147483647 is specified, the IaC modules auto-add a default "allow all" rule matching the native behavior of the GCP Terraform and Pulumi providers.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the security policy will be created
- **A backend service or load balancer** to attach the policy to (the policy is created independently; attachment is configured on the backend service)

## Quick Start

Create a file `cloud-armor.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudArmorPolicy
metadata:
  name: my-waf-policy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpCloudArmorPolicy.my-waf-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - action: allow
      priority: 1000
      description: Allow internal network
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - 10.0.0.0/8
    - action: "deny(403)"
      priority: 2147483647
      description: Default deny
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
```

Deploy:

```shell
planton apply -f cloud-armor.yaml
```

This creates a Cloud Armor policy that allows traffic from the `10.0.0.0/8` range and denies everything else.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the security policy will be created. | Required. Default kind: `GcpProject` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `policyName` | `string` | `metadata.name` | Name of the security policy in GCP. Must be 1-63 characters, RFC1035-compliant. |
| `description` | `string` | — | Description of the security policy. Max 2048 characters. |
| `type` | `string` | `CLOUD_ARMOR` | Policy type. `CLOUD_ARMOR` for HTTP(S) LB, `CLOUD_ARMOR_EDGE` for CDN/backend buckets, `CLOUD_ARMOR_INTERNAL_SERVICE` for Traffic Director. Immutable. |
| `adaptiveProtectionConfig` | `object` | — | Layer 7 DDoS defense configuration. See below. |
| `advancedOptionsConfig` | `object` | — | JSON parsing, logging, and IP resolution settings. See below. |
| `rules` | `object[]` | `[]` | Security rules evaluated in priority order. If empty, a default "allow all" rule is auto-added. |

### Adaptive Protection Config

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enableLayer7DdosDefense` | `bool` | `false` | Enable Cloud Armor Adaptive Protection for automatic DDoS detection. |
| `ruleVisibility` | `string` | — | `STANDARD` or `PREMIUM` (requires Managed Protection Plus). |

### Advanced Options Config

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `jsonParsing` | `string` | `DISABLED` | `DISABLED`, `STANDARD`, or `STANDARD_WITH_GRAPHQL`. |
| `logLevel` | `string` | `NORMAL` | `NORMAL` or `VERBOSE` (includes matched rule details). |
| `userIpRequestHeaders` | `string[]` | `[]` | Custom headers for client IP resolution behind CDN/proxy. |
| `requestBodyInspectionSize` | `string` | `8KB` | `8KB`, `16KB`, `32KB`, `48KB`, or `64KB`. Pulumi only — not supported in Terraform. |

### Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | `string` | Yes | `allow`, `deny(403)`, `deny(404)`, `deny(502)`, `redirect`, `throttle`, or `rate_based_ban`. |
| `priority` | `int32` | Yes | 0 to 2147483647. Lower = higher priority. Must be unique per policy. |
| `match` | `object` | Yes | Traffic matching condition. See below. |
| `description` | `string` | No | Rule description. Max 64 characters. |
| `preview` | `bool` | No | When `true`, rule is logged but not enforced. |
| `rateLimitOptions` | `object` | No | Required for `throttle` and `rate_based_ban` actions. |
| `redirectOptions` | `object` | No | Required for `redirect` action. `type`: `EXTERNAL_302` or `GOOGLE_RECAPTCHA`. |
| `headerAction` | `object` | No | Custom headers to inject into matching requests. CLOUD_ARMOR only. |
| `preconfiguredWafConfig` | `object` | No | WAF rule exclusions for handling false positives. CLOUD_ARMOR only. |

### Match Fields

| Field | Type | Description |
|-------|------|-------------|
| `versionedExpr` | `string` | Set to `SRC_IPS_V1` for IP-based matching. Mutually exclusive with `expression`. |
| `srcIpRanges` | `string[]` | CIDR ranges. Required when `versionedExpr` is set. Use `*` for all IPs. Max 10 per rule. |
| `expression` | `string` | CEL expression for advanced matching (geo, path, headers). Mutually exclusive with `versionedExpr`. |

## Examples

### Geo-Blocking with CEL Expression

Block traffic from specific countries:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudArmorPolicy
metadata:
  name: geo-blocking
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudArmorPolicy.geo-blocking
spec:
  projectId:
    value: my-gcp-project
  rules:
    - action: "deny(403)"
      priority: 1000
      description: Block restricted regions
      match:
        expression: "origin.region_code == 'CN' || origin.region_code == 'RU'"
    - action: allow
      priority: 2147483647
      description: Default allow
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
```

### Per-IP Rate Limiting

Protect APIs from abuse with per-IP request throttling:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudArmorPolicy
metadata:
  name: api-rate-limit
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudArmorPolicy.api-rate-limit
spec:
  projectId:
    value: my-gcp-project
  rules:
    - action: throttle
      priority: 1000
      description: Rate limit per IP
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
      rateLimitOptions:
        conformAction: allow
        exceedAction: "deny(429)"
        enforceOnKey: IP
        rateLimitThreshold:
          count: 100
          intervalSec: 60
    - action: allow
      priority: 2147483647
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
```

### Production WAF with Adaptive Protection

Full-featured OWASP protection with DDoS defense and rate limiting:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudArmorPolicy
metadata:
  name: prod-waf
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpCloudArmorPolicy.prod-waf
spec:
  projectId:
    value: my-gcp-project
  type: CLOUD_ARMOR
  description: Production WAF policy
  adaptiveProtectionConfig:
    enableLayer7DdosDefense: true
    ruleVisibility: STANDARD
  advancedOptionsConfig:
    jsonParsing: STANDARD
    logLevel: VERBOSE
  rules:
    - action: allow
      priority: 100
      description: Internal network
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - 10.0.0.0/8
    - action: "deny(403)"
      priority: 1000
      description: SQLi protection
      match:
        expression: "evaluatePreconfiguredWaf('sqli-v33-stable')"
    - action: "deny(403)"
      priority: 2000
      description: XSS protection
      match:
        expression: "evaluatePreconfiguredWaf('xss-v33-stable')"
    - action: throttle
      priority: 3000
      description: Per-IP rate limit
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
      rateLimitOptions:
        conformAction: allow
        exceedAction: "deny(429)"
        enforceOnKey: IP
        rateLimitThreshold:
          count: 200
          intervalSec: 60
    - action: "deny(403)"
      priority: 2147483647
      description: Default deny
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `policy_id` | `string` | Fully qualified resource ID (`projects/{project}/global/securityPolicies/{name}`) |
| `policy_name` | `string` | Name of the security policy as it exists in GCP |
| `policy_self_link` | `string` | Self-link URI of the policy, used when attaching to backend services and load balancers |
| `fingerprint` | `string` | Server-computed fingerprint for optimistic concurrency control |

## Related Components

- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the network context for CLOUD_ARMOR_INTERNAL_SERVICE policies
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project referenced by `projectId`
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — provides encryption keys if needed for related services
- [GcpFirewallRule](/docs/catalog/gcp/gcpfirewallrule) — complements Cloud Armor with network-level firewall rules
