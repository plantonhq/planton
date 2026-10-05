# Cloudflare Email Routing Rule

Defines a single Cloudflare Email Routing rule for a zone: match inbound mail -- by a specific recipient or all messages -- and then drop it, forward it to verified destinations, or hand it to an Email Worker. Rules deploy on any zone but only take effect once the zone's Email Routing is enabled (a `CloudflareEmailRoutingZone`), and every forwarding destination must be a verified `CloudflareEmailRoutingAddress` -- the API rejects the rule otherwise. Rules are evaluated in priority order, so specific rules can take precedence over broad ones.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Routing Rule** -- a matcher set plus an ordered action list, evaluated at the configured priority against the zone's inbound mail

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Email Routing edit access (and Workers Scripts access when routing to an Email Worker). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Email Routing enabled** -- the zone must already have Email Routing enabled (`CloudflareEmailRoutingZone`); the API rejects rules on a zone where it is off.
- **Verified destinations (only for forward actions)** -- a forwarding rule can only deliver to verified `CloudflareEmailRoutingAddress` mailboxes, and verification happens through an emailed link, out-of-band of any deploy.

## Deploy

### Console

Open the deployment store, find **Cloudflare Email Routing Rule**, and click **Deploy**. The creation wizard captures the zone and rule metadata (name, enabled, priority), at least one matcher (a specific recipient or all messages), and the ordered actions (drop / forward / Email Worker -- combinable in one rule). Start from the **Forward an Address** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingRule
metadata:
  name: forward-support
  org: acme-corp
  env: prod
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  name: forward-support
  matchers:
    - type: literal
      field: to
      value: support@example.com
  actions:
    - type: forward
      forwardTo:
        - value: ops@example.com
```

```shell
planton apply -f cloudflare-email-routing-rule.yaml
```

This forwards mail addressed to `support@example.com` to `ops@example.com`. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone and destination address are deployed in the same InfraPipeline, wire the rule to both with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  matchers:
    - type: literal
      field: to
      value: support@example.com
  actions:
    - type: forward
      forwardTo:
        - valueFrom:
            kind: CloudflareEmailRoutingAddress
            name: ops-mailbox
            fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, creates the zone and address first, then the rule -- keeping in mind the address still needs its human verification click before Cloudflare accepts the forward.

## Key Configuration

These are the most important decisions when configuring a routing rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone (`zoneId`)** -- the zone whose inbound mail this rule filters. Immutable -- changing it replaces the rule.

**Matchers (`matchers`)** -- at least one; a message matches the rule if ANY matcher matches. A `literal` matcher targets exactly one recipient (`field: to` is the only supported field). An `all` matcher makes the rule catch-all-shaped -- but prefer the zone's real catch-all (`CloudflareEmailRoutingZone.catchAll`) for fallback policy: it is evaluated after every rule by design, while an `all` rule competes on priority.

**Actions (`actions`)** -- one or more, applied in order. One rule can forward to mailboxes AND hand the message to an Email Worker -- the common "human inbox + automation" shape. `drop` stands alone by nature: combining it with delivery actions is contradictory even though the API does not statically reject it.

**Priority (`priority`)** -- lower runs first; `0` is fine for a single rule. With several rules, space priorities (10, 20, 30, ...) so later inserts need no renumbering, and keep specific-recipient rules at lower numbers than any broad rule.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareEmailRoutingAddress** (per forward action) | `actions[].forwardTo[]` | `status.outputs.email` |
| **CloudflareWorker** (worker actions) | `actions[].worker` | `status.outputs.script_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_id` | The Cloudflare-assigned identifier of the rule | Verification tooling and imports -- paired with `zone_id`, since the rule's API identity is the (zone, rule) tuple |
| `zone_id` | The zone the rule belongs to | Completes the rule's API identity for tooling that composes on it |

## Common Patterns

**Role-address forwarding** -- a `literal` matcher on `support@` with a forward action routes a shared alias to a real verified inbox. Start from the **Forward an Address** preset.

**Programmatic handling** -- a `worker` action hands matched mail to an Email Worker for parsing, ticket creation, or auto-responses. Start from the **Route to an Email Worker** preset.

**Inbox plus automation** -- one rule with both a forward action and a worker action: humans see the mail, the Worker processes it. Start from the **Forward AND Process with a Worker** preset.

## Works With

- [**Cloudflare Email Routing Zone**](/infra-catalog/cloudflare-email-routing-zone) -- enables Email Routing on the zone this rule belongs to
- [**Cloudflare Email Routing Address**](/infra-catalog/cloudflare-email-routing-address) -- the verified mailboxes a forwarding rule delivers to
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- the Email Worker a worker rule hands matched mail to
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone whose inbound mail is filtered; `zoneId` references its output
