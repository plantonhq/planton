# CloudflareEmailRoutingRule

Declare a single Email Routing rule for a zone: match incoming mail and drop it,
forward it to verified destinations, and/or hand it to an Email Worker — one rule
carries a LIST of actions, so it can forward AND process in the same match.
The API accepts rules on any zone, but they only take effect once the zone's
Email Routing is enabled (`CloudflareEmailRoutingZone`) — pair the two in any
real setup.

## When to use

- Routing a specific address (e.g. `support@`) to one or more mailboxes.
- Sending matched mail to an Email Worker for custom processing.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingRule
metadata:
  name: support-to-ops
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-com
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

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `zoneId` | yes | Zone ID, or a reference to a `CloudflareDnsZone` |
| `name` | no | Descriptive rule name |
| `enabled` | no | Whether the rule is active (default true) |
| `priority` | no | Evaluation order, lower first (default 0) |
| `matchers` | yes | One or more: `{ type: all\|literal, field: "to", value }` |
| `actions` | yes | One or more, applied in order: `{ type: drop\|forward\|worker, forwardTo[], worker }` |

`forward` requires `forwardTo`; `worker` requires `worker`. A `literal` matcher
requires `field` and `value`; an `all` matcher requires neither.

## Outputs

| Output | Description |
|---|---|
| `rule_id` | The routing rule identifier |
| `zone_id` | The zone the rule belongs to |

## Related kinds

- `CloudflareEmailRoutingZone` — must enable Email Routing on the zone first.
- `CloudflareEmailRoutingAddress` — the verified destinations a rule forwards to.
- `CloudflareWorker` — an Email Worker a rule can hand messages to.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
