# Cloudflare Email Routing Zone

Enables Cloudflare Email Routing on a DNS zone -- the anchor of the Email Routing family. Enabling provisions the records inbound mail needs (MX, SPF, DKIM) automatically and configures the single per-zone catch-all rule that decides what happens to mail no other routing rule matched: leave Cloudflare's default (drop), forward to verified destinations, or hand it to an Email Worker. Enabling replaces the zone's existing mail delivery path, so never enable it on a domain whose mail another system (Google Workspace, O365) must keep handling. Individual routing rules and destination addresses are separate Infra Components that build on top of an enabled zone.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Email Routing enablement** -- turns Email Routing on for the zone; destroying it disables routing (a toggle, not a deletion)
- **DNS records** -- the required MX, SPF, and DKIM records, created automatically on enable; with `lockDnsRecords` the module manages them explicitly so they cannot be edited out-of-band
- **Catch-all rule** -- created only when `catchAll` is set; the single per-zone rule for unmatched mail: drop, forward to addresses, or send to an Email Worker

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Email Routing edit access (and DNS edit access for the auto-provisioned records). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **An active zone** -- the domain must already exist as a Cloudflare zone (manage it with a `CloudflareDnsZone`, or reference an externally-managed zone ID).
- **No competing mail system** -- enabling rewrites the zone's MX/SPF/DKIM records and takes over inbound delivery for the domain.
- **Verified destinations (only for a forwarding catch-all)** -- a forwarding catch-all can only deliver to verified `CloudflareEmailRoutingAddress` mailboxes.

## Deploy

### Console

Open the deployment store, find **Cloudflare Email Routing Zone**, and click **Deploy**. The creation wizard captures the zone, whether the auto-created DNS records are locked, and the optional catch-all action. Leave the catch-all as **None** to keep Cloudflare's default (drop, disabled). Start from the **Forward-All Email Routing** or **Drop Catch-All Email Routing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingZone
metadata:
  name: acme-com-email-routing
  org: acme-corp
  env: prod
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  catchAll:
    enabled: true
    actions:
      - type: forward
        forwardTo:
          - value: catch-all@example.com
```

```shell
planton apply -f cloudflare-email-routing-zone.yaml
```

This enables Email Routing on the `acme-com` zone and forwards all otherwise-unmatched mail to a catch-all mailbox. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone and the catch-all's destination address are deployed in the same InfraPipeline, wire both with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  catchAll:
    enabled: true
    actions:
      - type: forward
        forwardTo:
          - valueFrom:
              kind: CloudflareEmailRoutingAddress
              name: catch-all-mailbox
              fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, creates the zone and address first, then enables routing -- the address still needs its human verification click before Cloudflare accepts the forward.

## Key Configuration

These are the most important decisions when enabling Email Routing. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone (`zoneId`)** -- the zone to enable Email Routing on. Immutable -- changing it replaces the resource. Enabling is the takeover moment: Cloudflare rewrites the zone's mail records and starts accepting inbound mail.

**Catch-all Actions (`catchAll.actions`)** -- what happens to unmatched mail, as a list applied in order: `drop` discards it silently (Cloudflare's default), `forward` delivers to `forwardTo` destinations, `worker` hands it to an Email Worker -- and one catch-all can forward AND invoke a Worker. Omit `catchAll` to leave the default.

**Catch-all deletion is a no-op on Cloudflare's side** -- destroying the resource (or removing `catchAll` and re-applying) leaves the zone's last catch-all configuration in place. To actually neutralize a catch-all without disabling routing, set it to a disabled drop (`enabled: false`, `actions: [{type: drop}]`) rather than deleting it.

**Lock DNS Records (`lockDnsRecords`)** -- manage the auto-created MX/SPF/DKIM records explicitly so they cannot be edited out-of-band. Destroying the managed-DNS resource removes those records.

**Subdomain routing (`dnsName`)** -- targets the managed records at a subdomain (`mail.example.com`) instead of the apex, making Cloudflare route mail addressed to `*@mail.example.com`. It requires `lockDnsRecords: true` -- the subdomain choice lives on the managed-DNS resource, and the spec rejects `dnsName` without it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareEmailRoutingAddress** (per forwarding catch-all action) | `catchAll.actions[].forwardTo[]` | `status.outputs.email` |
| **CloudflareWorker** (worker catch-all actions) | `catchAll.actions[].worker` | `status.outputs.script_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_id` | The zone Email Routing was enabled on | Referenced by `CloudflareEmailRoutingRule`s for the same zone |
| `enabled` | Whether Email Routing is enabled | Gating rule deployment on the zone being ready to route |
| `status` | The Email Routing configuration status (`ready`, `unconfigured`, `misconfigured`) | Diagnosing a zone that accepts but misroutes mail |

## Common Patterns

**Forward-all** -- a forwarding catch-all funnels every inbound message to one or more real mailboxes; add specific rules later for per-address routing. Start from the **Forward-All Email Routing** preset.

**Strict routing** -- keep the catch-all at drop and rely on explicit `CloudflareEmailRoutingRule`s, so only addresses you deliberately route are accepted. Start from the **Drop Catch-All Email Routing** preset.

**Full mail setup, in order** -- register a `CloudflareEmailRoutingAddress` for each real mailbox (owners click the verification links), then this kind to enable routing with the catch-all as fallback policy, then a `CloudflareEmailRoutingRule` per address you route explicitly -- rules always run before the catch-all regardless of creation order.

## Works With

- [**Cloudflare Email Routing Rule**](/infra-catalog/cloudflare-email-routing-rule) -- per-recipient rules that take precedence over the catch-all
- [**Cloudflare Email Routing Address**](/infra-catalog/cloudflare-email-routing-address) -- the verified mailboxes a forwarding catch-all delivers to
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- the Email Worker a worker catch-all hands unmatched mail to
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone Email Routing is enabled on
