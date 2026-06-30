# GcpCloudArmorPolicy

A GCP Cloud Armor security policy that provides WAF (Web Application Firewall) and DDoS protection for HTTP(S) load balancers, Cloud CDN, and backend services. This component provisions a Cloud Armor policy with inline rules—prioritized traffic-matching rules that evaluate incoming requests and apply actions (allow, deny, rate-limit, redirect).

## When to Use

Use `GcpCloudArmorPolicy` when you need:

- **WAF protection** — Block common web attacks (SQL injection, XSS, LFI, RCE) using preconfigured OWASP ModSecurity rules
- **DDoS defense** — Rate limiting, Layer 7 DDoS mitigation, and optional Adaptive Protection for anomaly detection
- **Rate limiting** — Throttle or ban traffic that exceeds per-IP, per-path, or custom-key request thresholds
- **IP allowlisting** — Restrict access to specific CIDR ranges or default-deny policies

## What This Component Creates

This component provisions a single Google Cloud Armor security policy with inline rules. The policy can be attached to backend services, load balancers, or CDN configurations. It does not create or modify backends—you attach the policy to your existing infrastructure.

## Key Features

- **IP-based rules** — Match traffic by source IP ranges (CIDR) using `versioned_expr: SRC_IPS_V1` and `src_ip_ranges`
- **CEL expressions** — Advanced matching via Common Expression Language: `origin.region_code`, `request.path`, `request.headers`, `inIpRange()`, and more
- **Rate limiting** — Throttle or rate-based ban with configurable thresholds and enforce-on-key (IP, HTTP_HEADER, HTTP_PATH, etc.)
- **Redirect** — Send users to reCAPTCHA challenge or custom URL (`EXTERNAL_302`, `GOOGLE_RECAPTCHA`)
- **Header injection** — Add custom headers to matching requests before forwarding to backends
- **Preconfigured WAF exclusions** — Carve out false positives (e.g., SQL in search params, HTML in rich text) via exclusions per rule set
- **Adaptive Protection** — Enable Layer 7 DDoS anomaly detection and auto-mitigation
- **JSON parsing** — Inspect JSON and GraphQL request bodies for WAF rules
- **Preview mode** — Log matched traffic without enforcing actions to test rules safely

## Policy Types

Three policy types determine where the policy can be attached and which features are available:

| Type | Use Case | Features |
|------|----------|----------|
| `CLOUD_ARMOR` (default) | HTTP(S) load balancer backends | Full feature set: WAF, rate limit, redirect, header injection |
| `CLOUD_ARMOR_EDGE` | Cloud CDN, backend buckets | IP and geo-based rules only |
| `CLOUD_ARMOR_INTERNAL_SERVICE` | Internal Traffic Director | Limited feature set |

The policy type is immutable after creation.

## Default Rule Behavior

Every Cloud Armor policy must have a rule at priority `2147483647` (the default rule). If you do not specify one, the IaC modules automatically add an "allow all" rule at that priority—matching the behavior of the GCP Terraform and Pulumi providers. This ensures unmatched traffic is allowed by default.

## Quick Start

Minimal policy with a single allow rule:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudArmorPolicy
metadata:
  name: my-waf-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Allow all traffic (default rule auto-added at 2147483647 if omitted)
```

## StringValueOrRef: projectId

The `projectId` field uses `StringValueOrRef`. You can pass a literal value or a reference to a `GcpProject` resource:

```yaml
# Literal value
projectId:
  value: my-gcp-project

# Reference to GcpProject
projectId:
  valueFrom:
    kind: GcpProject
    name: my-project
    fieldPath: status.outputs.project_id
```

## Outputs

| Output | Description |
|--------|-------------|
| `policy_id` | Fully qualified resource ID (`projects/{project}/global/securityPolicies/{name}`) |
| `policy_name` | Name as it exists in GCP |
| `policy_self_link` | Self-link URI (used when attaching to backend services) |
| `fingerprint` | Server-computed fingerprint for concurrency control |
