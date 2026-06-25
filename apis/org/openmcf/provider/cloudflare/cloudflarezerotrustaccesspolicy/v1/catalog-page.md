# Cloudflare Zero Trust Access Policy

Define who can reach a protected resource, once, and attach it to many applications.

A Cloudflare Zero Trust Access policy is a reusable, account-scoped decision —
allow, deny, non-identity (service tokens), or bypass — combined with the
include/exclude/require rules that decide who it applies to. Policies attach to
Access applications by reference, so one policy can guard many apps and its rules
live in one place.

## Highlights

- **Reusable decisions** — one policy, attached to many applications.
- **Full rule surface** — every Cloudflare Access rule type, including reusable
  group references, IdP groups, device posture, service tokens, and user-risk.
- **Governance built in** — approval workflows, purpose justification, browser
  isolation, and per-policy MFA.
- **Composable** — references `CloudflareZeroTrustAccessGroup`s; referenced by
  `CloudflareZeroTrustAccessApplication`s, all through the resource graph.

## Typical use

Pair with `CloudflareZeroTrustAccessGroup` (reusable rule bundles) and
`CloudflareZeroTrustAccessApplication` (the protected resource) to model your
organization's access control as a composable graph.
