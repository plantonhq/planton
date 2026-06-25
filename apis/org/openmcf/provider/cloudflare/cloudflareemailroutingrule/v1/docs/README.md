# CloudflareEmailRoutingRule — Research & Design Notes

## What it is

A per-zone Email Routing rule: a set of matchers plus an action. Rules are
evaluated by priority; the first match wins, falling back to the zone catch-all.

## Typed action (vs the provider's generic shape)

The provider models a rule action as `{type, value[]}` of plain strings. The spec
exposes a typed action — `type` plus `forward_to` (FK →
`CloudflareEmailRoutingAddress`) and `worker` (FK → `CloudflareWorker`) — so the
forwarding targets and the Email Worker are real graph edges. The modules map it
back to `{type, value[]}` (forward → the addresses; worker → the single script
name; drop → no values). The action is modeled as a single message (not a list)
because a rule has exactly one action; the module wraps it in the provider's
one-element actions list.

## Matchers

`all` (every message) or `literal` (a specific recipient via `field: "to"` +
`value`). CEL enforces that a literal matcher carries field+value and an all
matcher carries neither.

## Defaults

`enabled` defaults to true via the spec default option; `priority` defaults to 0.

## Engine parity

Both engines build the same matchers and the same single-element actions list from
the typed action, and read `enabled`/`priority` identically. Outputs map to
`rule_id` / `zone_id`.

## Live-validation note

Requires Email Routing enabled on the zone. A `drop` or `worker` rule needs no
external mailbox; a `forward` rule's target address must be verified out-of-band
before mail is actually delivered.
