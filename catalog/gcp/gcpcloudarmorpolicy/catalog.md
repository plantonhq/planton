# GCP Cloud Armor Policy

Deploys a Cloud Armor security policy with configurable rules for IP allowlisting/denylisting, rate limiting, ban escalation, OWASP WAF protection, and Layer 7 DDoS defense — global (the default) or, with `region` set, regional, where the same block also builds a `CLOUD_ARMOR_NETWORK` policy that filters packets and enables network DDoS protection for passthrough Network Load Balancers. The policy attaches to HTTP(S) load balancers, Cloud CDN backends, or internal Traffic Director services — the attachment itself lives on the consuming backend service or backend bucket, which references this policy's self-link, and the scopes must match.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine API enablement** (`compute.googleapis.com`) on the target project (never disabled on destroy)
- **Security Policy** -- a `compute.SecurityPolicy` (global, `region` empty) or a `compute.RegionSecurityPolicy` (`region` set) in the specified GCP project, configured with the chosen policy type, rules, and advanced options; exactly one of the two exists
- **Security Rules** -- one rule per entry in `rules`, each with a priority, action (allow, deny, throttle, rate_based_ban, redirect), match condition (IP ranges or CEL expression, or a packet-level `networkMatch` on a regional `CLOUD_ARMOR_NETWORK` policy), and optional rate limiting, redirect, header injection, or WAF exclusion configuration
- **Network DDoS protection and user-defined fields** -- on a regional `CLOUD_ARMOR_NETWORK` policy, the `ddosProtectionConfig` level and the custom packet fields its rules match
- **Network Edge Security Service** -- created only when a regional network policy declares `networkEdgeSecurityService`; enrolls the region in advanced network DDoS protection with this policy attached (one per region per project)
- **Adaptive Protection** -- created only when `adaptiveProtectionConfig` is present; enables automatic Layer 7 DDoS detection and alerting
- **Advanced Options** -- created only when `advancedOptionsConfig` is present; configures JSON body parsing (with optional custom content types), logging verbosity, and client IP resolution headers
- **Default Rule** -- if no rule at priority 2147483647 is provided, the IaC module auto-adds a default "allow all" rule
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically to a global policy for tracking and governance (the regional collection carries no labels)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the security policy will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef.
- **An HTTP(S) load balancer** or backend service to attach the policy to (configured outside this Cloud Resource via the backend service's `securityPolicy` field).

## Deploy

### Console

Open the deployment store, find **GCP Cloud Armor Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic IP Allowlist** preset in the [Presets](#presets) tab to pre-populate a deny-by-default policy with IP-based allow rules.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudArmorPolicy
metadata:
  name: api-protection
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  policyName: api-protection
  type: CLOUD_ARMOR
  rules:
    - action: allow
      priority: 1000
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges: ["10.0.0.0/8"]
    - action: "deny(403)"
      priority: 2147483647
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges: ["*"]
```

```shell
planton apply -f gcp-cloud-armor-policy.yaml
```

This creates a CLOUD_ARMOR policy that allows traffic from RFC 1918 ranges and denies everything else with 403. The policy must be attached to a backend service separately. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the policy to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the Cloud Armor policy with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a Cloud Armor policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- Leave `region` empty for a global policy (global external ALB backends, backend buckets, CDN). Set `region` for a regional policy (regional external and internal ALB backends, or a `CLOUD_ARMOR_NETWORK` policy for passthrough Network Load Balancers). A regional backend service accepts only a regional policy; the global-only levers (labels, Adaptive Protection, reCAPTCHA, `requestBodyInspectionSize`, `redirect`, header injection) are rejected when `region` is set. Immutable.

**Policy type** -- Set `type` to CLOUD_ARMOR (default) for backend security policies with full WAF, rate limiting, and (globally) header injection. Use CLOUD_ARMOR_EDGE for CDN and backend bucket protection (IP/geo rules only). Use CLOUD_ARMOR_INTERNAL_SERVICE (global only) for internal Traffic Director services. Use CLOUD_ARMOR_NETWORK (regional only) for packet-level filtering and network DDoS protection: `ddosProtectionConfig.ddosProtection` STANDARD (free) or ADVANCED / ADVANCED_PREVIEW (Cloud Armor Enterprise, the region enrolled through `networkEdgeSecurityService`), `userDefinedFields`, and rules that match through `networkMatch` instead of `match`. Type is immutable after creation.

**Rule priority and evaluation** -- Rules are evaluated from lowest priority number (highest precedence) to highest. Priority 2147483647 is reserved for the default rule. Plan priority numbering with gaps (e.g., 1000, 2000, 3000) to allow inserting rules later without renumbering.

**Rate limiting and ban escalation** -- Use `throttle` action for simple rate limiting or `rate_based_ban` for two-tier protection. Rate limiting uses `rateLimitThreshold` to cap requests per interval; ban escalation adds `banThreshold` and `banDurationSec` to fully block persistent abusers. Set `enforceOnKey` to `IP` for per-source limiting.

**WAF rules and exclusions** -- Match against preconfigured OWASP rule sets (e.g., `sqli-v33-stable`, `xss-v33-stable`) using CEL expressions. Add `preconfiguredWafConfig` exclusions for request fields that trigger false positives -- critical for production APIs that accept user-generated content.

**Adaptive Protection** -- Enable `adaptiveProtectionConfig.enableLayer7DdosDefense` for automatic anomaly detection. STANDARD visibility is available to all Cloud Armor users. PREMIUM requires Managed Protection Plus.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | Fully qualified resource ID (`projects/{p}/global/securityPolicies/{name}`, or `projects/{p}/regions/{r}/securityPolicies/{name}` for a regional policy) | Backend service security policy references |
| `policy_name` | Name of the security policy in GCP | Audit logs, monitoring dashboards |
| `policy_self_link` | Self-link URI of the security policy | Attaching to backend services, load balancers, CDN configurations |
| `fingerprint` | Server-computed fingerprint for optimistic concurrency | Out-of-band policy updates |
| `region` | Region of a regional policy; empty for a global one | Telling the scope from the outputs alone |
| `network_edge_security_service_self_link` | Self-link of the region's network edge security service when declared; empty otherwise | Confirming the region's advanced DDoS enrollment |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic IP allowlist** -- Deny-by-default policy that allows traffic only from specified CIDR ranges (corporate networks, VPNs). All other traffic receives 403 Forbidden. Suitable for internal dashboards and admin APIs. Start from the **Basic IP Allowlist** preset.

**Rate limiting for APIs** -- Per-IP rate limiting with ban escalation. Throttles traffic beyond 100 requests per minute, then bans persistent abusers exceeding 500 requests over 5 minutes for 1 hour. Start from the **Rate Limiting for API Endpoints** preset.

**WAF OWASP protection** -- OWASP WAF rules blocking SQL injection and XSS attacks, with adaptive Layer 7 DDoS protection, JSON body parsing, and verbose logging. Suitable for internet-facing web applications. Start from the **WAF OWASP Protection** preset.

**Regional WAF for a regional ALB** -- The same allowlist, geo-block, and rate-limit rules on a regional policy a regional backend service attaches. Start from the **Regional WAF for a Regional Load Balancer** preset.

**Network DDoS protection for a passthrough NLB** -- A regional `CLOUD_ARMOR_NETWORK` policy with STANDARD DDoS protection and a packet-level allowlist on a user-defined field. Start from the **Network DDoS Protection for a Passthrough Load Balancer** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the security policy is created
- [**GCP Backend Service**](/cloud-catalog/gcp-backend-service) -- consumes `policy_self_link` as its `securityPolicy` (backend WAF) or `edgeSecurityPolicy` (edge filtering); a regional backend service takes a regional policy from the same region
- [**GCP Backend Bucket**](/cloud-catalog/gcp-backend-bucket) -- consumes a CLOUD_ARMOR_EDGE policy's `policy_self_link` as its `edgeSecurityPolicy`