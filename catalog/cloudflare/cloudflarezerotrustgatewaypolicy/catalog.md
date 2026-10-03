# Cloudflare Zero Trust Gateway Policy

Deploys one Cloudflare Gateway policy: a wirefilter expression over employee traffic (DNS queries, HTTP requests, or network connections) plus the action taken on a match — block, allow, isolate, redirect, override, resolve, and kin. Identity and device-posture expressions can narrow who and which devices the rule applies to, and `ruleSettings` carries the action-specific behavior (block pages, session checks, custom resolvers, egress IPs, isolation controls). One behavior deserves a loud warning: `enabled` defaults to false at Cloudflare, so a policy authored without `enabled: true` deploys disabled and filters nothing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Gateway Policy** — a `cloudflare_zero_trust_gateway_policy` on the account, carrying the action, the singular `filter` (sent as a one-element list, which is all Cloudflare accepts), the three wirefilter expressions, and optional expiration or per-weekday schedule for DNS rules
- **Rule Settings** — always sent: an empty object when the spec configures nothing (the provider's own drift workaround), otherwise the full settings tree

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **The matching add-on** (only for `isolate` and `egress` actions) — Browser Isolation for isolate, dedicated egress IPs for egress. Those actions fail the apply on an account that lacks the entitlement; nothing is billed or upgraded through this component.
- **A Zero Trust list** (only when `traffic` references a list by ID) — deploy the list first and use its ID in the expression.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Gateway Policy**, and click **Deploy**. The creation wizard walks you through the owning account, the action and filter, the traffic/identity/device-posture expressions, and the action-specific rule settings. Start from the **DNS block** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewayPolicy
metadata:
  name: block-gambling
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: block-gambling-domains
  action: block
  filter: dns
  enabled: true
  precedence: 1000
  traffic: any(dns.domains[*] == "gambling.example.com")
  ruleSettings:
    blockPageEnabled: true
    blockReason: Blocked by company policy
```

```shell
planton apply -f gateway-policy.yaml
```

This creates an active DNS policy blocking the named domain for every WARP-enrolled device, with the block page shown and a reason attached. Without `enabled: true` the same manifest would deploy a rule that filters nothing. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire a private resolver's virtual network to a resource managed by another Infra Component:

```yaml
spec:
  action: resolve
  filter: dns_resolver
  enabled: true
  ruleSettings:
    dnsResolvers:
      ipv4:
        - ip: 10.0.0.53
          routeThroughPrivateNetwork: true
          vnetId:
            valueFrom:
              kind: CloudflareZeroTrustTunnelVirtualNetwork
              name: corp-vnet
              fieldPath: status.outputs.virtual_network_id
```

The InfraPipeline deploys the virtual network first, then provisions the policy against its real ID. `routeThroughPrivateNetwork: true` is required whenever `vnetId` is set — validation rejects the manifest without it.

## Key Configuration

These are the most important decisions when configuring a Gateway policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`enabled` defaults to false** — Cloudflare treats a missing value as disabled. Write `enabled: true` on every policy you intend to enforce; a code review that does not see that field should bounce the change.

**Precedence is yours to manage** — omit `precedence` and Cloudflare assigns a number; two policies created that way can evaluate in an order you did not choose. Set it explicitly (lower runs earlier) and pick numbers with room to insert policies between them.

**Action and filter must pair** — which actions are valid depends on the filter: dns takes allow/block/override/safesearch/ytrestricted, http takes allow/block/isolate/redirect/quarantine/scan, l4 takes allow/block/audit-ssh/l4_override, egress rules take egress, dns_resolver rules take resolve. Leave `filter` empty to let Cloudflare infer it from the action. Each `ruleSettings` field likewise applies only to specific filter/action combinations.

**Wirefilter expressions are reformatted by the API** — `traffic`, `identity`, and `devicePosture` are rewritten before storing (whitespace, function spelling, list syntax). If a plan shows drift on an expression you did not change, the stored form is the API's — copy it back into the spec rather than fighting the formatter.

**`addHeaders` and `overrideIps` drift on first apply** — at provider v5.23.0, a policy carrying either shows computed-field drift even when applied correctly; the provider's own tests expect it. `blockPage`, `checkSession`, `blockReason`, and `dnsResolvers` do not have that defect — prefer them for policies you need to prove idempotent.

**Prefer disable over delete for reversibility** — destroy is a real delete, and a high-precedence block you might have to restore in a hurry is safer flipped to `enabled: false` than destroyed and re-created.

**Resolver mechanisms are exclusive** — a resolve policy uses at most one of `dnsResolvers` (custom upstreams), `resolveDnsInternally` (internal DNS views), or `resolveDnsThroughCloudflare` (public 1.1.1.1); validation rejects combinations.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.ruleSettings.dnsResolvers.ipv4[].vnetId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.ruleSettings.dnsResolvers.ipv6[].vnetId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | The UUID of the created policy | API automation and policy audits |
| `precedence` | The evaluation order (Cloudflare-assigned when the spec left it unset) | Slotting new policies deliberately around existing ones |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**DNS block** — the first Gateway policy on most accounts: block a domain (or a Zero Trust list of domains) for every WARP-enrolled device, with the block page and a reason. Start from the **DNS block** preset.

**HTTP allow with a session check** — allowlist an internal web app and require a fresh Access session every 24 hours; uses `checkSession` and `blockPage`, the settings that stay idempotency-safe. Start from the **HTTP allow with a session check** preset.

**Private DNS resolution** — a `resolve` policy routing matched internal hostnames to an on-premises resolver reached through a Zero Trust tunnel's virtual network, so corp domains answer correctly for roaming devices.

**Posture-gated egress** — an l4 or http policy whose `devicePosture` expression requires a passing posture rule (disk encryption, EDR presence) before traffic leaves, pairing this kind with device posture rules.

## Works With

- [**Cloudflare Zero Trust List**](/infra-catalog/cloudflare-zero-trust-list) — reusable domain/IP/email sets referenced by ID from the `traffic` expression.
- [**Cloudflare Zero Trust Tunnel Virtual Network**](/infra-catalog/cloudflare-zero-trust-tunnel-virtual-network) — the private network `ruleSettings.dnsResolvers.*.vnetId` routes resolver traffic through.
- [**Cloudflare Zero Trust Device Posture Rule**](/infra-catalog/cloudflare-zero-trust-device-posture-rule) — the health checks the `devicePosture` expression matches on.
- [**Cloudflare Zero Trust DNS Location**](/infra-catalog/cloudflare-zero-trust-dns-location) — the per-site entry points DNS policies can match on.
- [**Cloudflare Zero Trust Gateway Settings**](/infra-catalog/cloudflare-zero-trust-gateway-settings) — the account-wide posture (TLS decryption, logging) these policies run under.
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) — website WAF and firewall for zone traffic; Gateway policies filter employee traffic. Different product, different kind.
