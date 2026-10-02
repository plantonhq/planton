# GcpCloudArmorPolicy

A GCP Cloud Armor security policy that provides WAF (Web Application Firewall) and DDoS protection for HTTP(S) load balancers, Cloud CDN, backend services, and — as a regional network policy — the passthrough Network Load Balancers, protocol forwarding rules, and public-IP VMs in a region. This component provisions a Cloud Armor policy with inline rules—prioritized traffic-matching rules that evaluate incoming requests (or packets) and apply actions (allow, deny, rate-limit, redirect).

## When to Use

Use `GcpCloudArmorPolicy` when you need:

- **WAF protection** — Block common web attacks (SQL injection, XSS, LFI, RCE) using preconfigured OWASP ModSecurity rules
- **DDoS defense** — Rate limiting, Layer 7 DDoS mitigation, and optional Adaptive Protection for anomaly detection
- **Rate limiting** — Throttle or ban traffic that exceeds per-IP, per-path, or custom-key request thresholds
- **IP allowlisting** — Restrict access to specific CIDR ranges or default-deny policies
- **Regional load balancers** — The same WAF in front of a regional external or internal Application Load Balancer (set `region`)
- **Network DDoS protection** — Packet-level filtering and Google's network DDoS protection for passthrough Network Load Balancers (`region` + `type: CLOUD_ARMOR_NETWORK`)

## What This Component Creates

This component provisions a single Google Cloud Armor security policy with inline rules — global when `region` is empty, regional when it is set — and, when a regional network policy declares `networkEdgeSecurityService`, the region's network edge security service that enrolls it in advanced network DDoS protection. The policy can be attached to backend services, load balancers, or CDN configurations. It does not create or modify backends—you attach the policy to your existing infrastructure.

## One kind, two scopes

| `region` | Resource | Attached by | Types |
|---|---|---|---|
| empty (default) | global security policy | a global backend service or backend bucket (global external ALB, Cloud CDN) | `CLOUD_ARMOR`, `CLOUD_ARMOR_EDGE`, `CLOUD_ARMOR_INTERNAL_SERVICE` |
| set (e.g. `us-central1`) | regional security policy | a regional backend service (regional external / internal ALB); or, as `CLOUD_ARMOR_NETWORK`, the region's passthrough NLBs, protocol forwarding, and public-IP VMs | `CLOUD_ARMOR`, `CLOUD_ARMOR_EDGE`, `CLOUD_ARMOR_NETWORK` |

Scopes must match: a regional backend service accepts only a regional policy. The global-only levers (labels, Adaptive Protection, reCAPTCHA options, `requestBodyInspectionSize`, the `redirect` action, header injection, reCAPTCHA token options) are rejected when `region` is set; the network-policy levers (`CLOUD_ARMOR_NETWORK`, `ddosProtectionConfig`, `userDefinedFields`, `rules[].networkMatch`) are rejected when it is empty. A policy cannot move between scopes.

## Key Features

- **IP-based rules** — Match traffic by source IP ranges (CIDR) using `versioned_expr: SRC_IPS_V1` and `src_ip_ranges`
- **CEL expressions** — Advanced matching via Common Expression Language: `origin.region_code`, `request.path`, `request.headers`, `inIpRange()`, and more
- **Rate limiting** — Throttle or rate-based ban with configurable thresholds and enforce-on-key (IP, HTTP_HEADER, HTTP_PATH, etc.)
- **Redirect** — Send users to reCAPTCHA challenge or custom URL (`EXTERNAL_302`, `GOOGLE_RECAPTCHA`)
- **Header injection** — Add custom headers to matching requests before forwarding to backends
- **Preconfigured WAF exclusions** — Carve out false positives (e.g., SQL in search params, HTML in rich text) via exclusions per rule set
- **Adaptive Protection** — Enable Layer 7 DDoS anomaly detection and auto-mitigation
- **JSON parsing** — Inspect JSON and GraphQL request bodies for WAF rules, with `requestBodyInspectionSize` controlling how much of each body the WAF reads (8KB default, up to 64KB)
- **Preview mode** — Log matched traffic without enforcing actions to test rules safely
- **Labels** — User labels merged with the platform attribution labels (platform wins on key conflicts); global policies only — the regional collection carries no labels
- **Deletion policy** — `DELETE` (default), `PREVENT` (destroy refuses), or `ABANDON` (the policy keeps enforcing but leaves management); applies to the network edge security service too
- **Network DDoS protection** (regional `CLOUD_ARMOR_NETWORK`) — `ddosProtectionConfig.ddosProtection` `STANDARD` (free, always on), `ADVANCED`, or `ADVANCED_PREVIEW` (Cloud Armor Enterprise; the region enrolled through `networkEdgeSecurityService`)
- **Custom packet fields** (regional `CLOUD_ARMOR_NETWORK`) — `userDefinedFields` read up to 4 bytes at a fixed offset from the IPv4, IPv6, TCP, or UDP header, optionally masked; rules match them by name
- **Packet-level rules** (regional `CLOUD_ARMOR_NETWORK`) — `rules[].networkMatch` on source/destination ranges and ports, IP protocols, source country codes, source ASNs, and user-defined fields

## Policy Types

Four policy types determine where the policy can be attached and which features are available:

| Type | Scope | Use Case | Features |
|------|-------|----------|----------|
| `CLOUD_ARMOR` (default) | global or regional | HTTP(S) load balancer backends | WAF, rate limit; globally also redirect and header injection |
| `CLOUD_ARMOR_EDGE` | global or regional | Cloud CDN, backend buckets | IP and geo-based rules only |
| `CLOUD_ARMOR_INTERNAL_SERVICE` | global only | Internal Traffic Director | Limited feature set |
| `CLOUD_ARMOR_NETWORK` | regional only | Passthrough Network Load Balancers, protocol forwarding, public-IP VMs | Network DDoS protection, user-defined fields, L3/L4 `networkMatch` rules |

The policy type is immutable after creation.

## Default Rule Behavior

Every Cloud Armor policy carries a default rule at priority `2147483647`. Creating a policy with NO rules lets the GCP API add a default "allow all" rule automatically; providing ANY rules requires the set to include that default explicitly — the spec enforces this before deploy, mirroring the API's own rejection.

## Quick Start

Minimal policy with a single allow rule:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
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
      description: Default rule - required whenever any rules are provided
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

## Regional network policy

A `CLOUD_ARMOR_NETWORK` policy filters packets, not HTTP requests, so its rules use `networkMatch` instead of `match`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudArmorPolicy
metadata:
  name: nlb-shield
spec:
  region: us-central1
  type: CLOUD_ARMOR_NETWORK
  ddosProtectionConfig:
    ddosProtection: STANDARD
  userDefinedFields:
    - name: SIG1_AT_0
      base: TCP
      offset: 8
      size: 2
      mask: "0x8F00"
  rules:
    - priority: 100
      action: allow
      preview: true
      networkMatch:
        srcIpRanges: ["10.10.0.0/16"]
        userDefinedFields:
          - name: SIG1_AT_0
            values: ["0x8F00"]
    - priority: 2147483647
      action: deny(403)
      networkMatch: {}
```

## Outputs

| Output | Description |
|--------|-------------|
| `policy_id` | Fully qualified resource ID (`projects/{project}/global/securityPolicies/{name}`, or `projects/{project}/regions/{region}/securityPolicies/{name}` for a regional policy) |
| `policy_name` | Name as it exists in GCP |
| `policy_self_link` | Self-link URI (used when attaching to backend services; carries `regions/{region}` for a regional policy) |
| `fingerprint` | Server-computed fingerprint for concurrency control |
| `region` | Region of a regional policy; empty for a global one |
| `network_edge_security_service_self_link` | Self-link of the network edge security service when `networkEdgeSecurityService` is declared; empty otherwise |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
