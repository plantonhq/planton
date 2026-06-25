# Cloudflare Zero Trust Access Group

Define a reusable set of Access rules once and reference it from many policies.

Cloudflare Zero Trust Access groups are named, reusable bundles of membership
criteria — email domains, IdP groups, IP ranges, countries, device posture,
service tokens, user-risk levels, and more. Instead of repeating the same rules in
every Access policy, you define a group (for example "engineering-team" or
"platform-admins") and reference it wherever it applies. Groups can even be
composed out of other groups.

## Highlights

- **Reusable membership** — one group, referenced by many policies and groups.
- **Full rule surface** — every Cloudflare Access rule type (identity, network,
  device, service-token, risk, and external evaluation).
- **Composable** — `group` rules reference other groups by output ID; policies
  reference groups; everything wires through the resource graph.
- **Account or zone scoped** — account-level for broad reuse, or scoped to a zone.

## Typical use

Model your organization's recurring access criteria as groups, then assemble Access
policies from them. Pair with `CloudflareZeroTrustAccessPolicy` (the decision) and
`CloudflareZeroTrustAccessApplication` (the protected resource).
