# GcpCloudArmorPolicy Guide

The judgment this guide protects: a WAF policy is production traffic
filtering — a wrong rule blocks customers, a missing one admits attacks.
Preview mode and the priority ladder are what make changes safe; use
both, every time.

## Preview before enforce, always

Every rule supports `preview: true` — matched traffic is logged but the
action is not enforced. Ship new deny/throttle rules in preview, read
the logs for a representative window, then flip preview off. This is the
single highest-value habit with Cloud Armor: false positives found in
logs cost nothing; found in production they are an outage.

## The priority ladder is the architecture

Rules evaluate lowest priority number first, first match wins, and the
default rule at 2147483647 catches everything else. Leave deliberate
gaps (1000, 2000, 3000...) so emergency rules can slot between existing
ones. Creating with NO rules lets the API add an allow-all default;
providing ANY rules requires including that default explicitly — decide
its action consciously (allow-by-default perimeter vs deny-by-default
allowlist), because it IS your security posture.

## Rate limiting needs its key thought through

`throttle` and `rate_based_ban` are only as good as `enforceOnKey`: IP
alone punishes everyone behind a corporate NAT; HTTP_HEADER/HTTP_COOKIE
keys need the header to actually exist on abusive traffic. The
conform/exceed action pair is validated to provider truth (conform is
always `allow`). Start with generous thresholds in preview and tighten
from measured traffic, not intuition.

## WAF exclusions carve, they do not disable

Preconfigured WAF rules (SQLi, XSS...) false-positive on legitimate rich
content. The `exclusions` block removes a specific cookie, header, query
param, or URI from inspection for a named rule set — a scalpel. Reaching
for "remove the rule entirely" when one search field trips it is the
wrong tool; exclude that field instead. `requestBodyInspectionSize`
bounds how deep the WAF reads bodies (8KB default, 64KB max) — larger
catches deeper payloads at higher processing cost, and bodies beyond the
limit pass uninspected either way.

## Adaptive Protection is a second pair of eyes

Layer-7 DDoS detection learns baselines and can auto-deploy mitigations
via `thresholdConfigs` (confidence, impacted-baseline, expiration).
Start with detection only (`enable` without auto-deploy thresholds) and
graduate to auto-deploy once you trust its verdicts on your traffic.
Adaptive Protection watches global Application Load Balancers only; a
regional policy cannot carry it.

## One kind, two scopes

`region` empty builds Google's GLOBAL security policy — the one a global
backend service or backend bucket attaches. `region` set builds the
REGIONAL policy — the one a regional backend service (regional external
or internal Application Load Balancer) attaches. Scopes must match: a
regional backend service refuses a global policy and vice versa, and the
`policy_self_link` output carries `regions/{region}` so a chart can tell
them apart. A policy cannot move between scopes; changing `region`
recreates it, and the backend service must be re-pointed.

The regional collection is a narrower product. It has no labels (a
regional policy carries no platform labels either), no Adaptive
Protection, no reCAPTCHA (neither the policy-level site key nor
per-rule token options), no `requestBodyInspectionSize` (the WAF
inspects the default 8KB), no `redirect` action, and no header
injection; rate limits exceed to `deny(STATUS)` only. Every one of those
is rejected before deploy when `region` is set, with the message naming
the lever, so a manifest moved from global to regional fails loudly
instead of silently dropping protection.

## Network policies filter packets, not requests

`type: CLOUD_ARMOR_NETWORK` (regional only) is a different product under
the same name: it sits in front of passthrough Network Load Balancers,
protocol forwarding rules, and VMs with public IPs, and it sees packets.
Its rules match through `networkMatch` — source and destination ranges
and ports, IP protocols, source country codes, source ASNs, and
`userDefinedFields` (up to 4 bytes read at a fixed offset from the IPv4,
IPv6, TCP, or UDP header, optionally masked) — never through the HTTP
`match`. Every listed field must match (AND); within a field any listed
value matches (OR); an empty `networkMatch: {}` matches every packet and
is how the default rule is written. A rule that names a user-defined
field the policy never defined is rejected before deploy.

`ddosProtectionConfig.ddosProtection` is the reason most network
policies exist. `STANDARD` is Google's always-on protection, free with
the load balancer. `ADVANCED` adds the network-layer mitigations of
Cloud Armor Enterprise (Managed Protection Plus) and needs two things:
the project enrolled in Enterprise, and the region enrolled through a
network edge security service — declare `networkEdgeSecurityService` on
exactly one policy per region per project (Google allows one). Use
`ADVANCED_PREVIEW` first: Google logs what it would mitigate without
mitigating, the same preview habit as the rules. Without an Enterprise
subscription, `ADVANCED` is accepted by the API and silently does
nothing more than `STANDARD`; the module cannot check the subscription
for you.

## Priority 0 is a real rule

Priorities run 0 to 2147483647 and 0 is the HIGHEST — the rule Google
evaluates first. It is a legal, common choice for an emergency block. The
spec treats priority by presence, so 0 is accepted; only an omitted
priority is rejected.

## Teardown discipline

Detaching the policy from a backend service leaves that traffic
UNFILTERED — deletion is a security event, not a cleanup. GCP refuses to
delete a policy still attached, and `PREVENT` also covers the window
after detachment. `ABANDON` keeps the policy enforcing while dropping
management.
