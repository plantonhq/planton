---
title: "Team Access with Google Groups + MFA"
description: "Restricts access to a hostname to specific Google Workspace groups and requires multi-factor authentication. Use for sensitive internal tools where only certain teams should have access and MFA is..."
type: "preset"
rank: "02"
presetSlug: "02-team-google-groups"
componentSlug: "zero-trust-access-application"
componentTitle: "Zero Trust Access Application"
provider: "cloudflare"
icon: "package"
order: 2
---

# Team Access with Google Groups + MFA

Restricts access to a hostname to specific Google Workspace groups and requires multi-factor authentication. Use for sensitive internal tools where only certain teams should have access and MFA is mandated.

## When to Use

- Admin dashboards or sensitive tools for engineering or ops teams
- Per-team access control via Google Workspace groups
- Environments requiring MFA for compliance or security

## Key Configuration Choices

- **allowedGoogleGroups** (`allowedGoogleGroups`) -- List of Google Workspace group emails; only members get access.
- **requireMfa: true** (`requireMfa: true`) -- MFA required before access is granted.
- **sessionDurationMinutes 480** (`sessionDurationMinutes: 480`) -- 8 hours; shorter for high-sensitivity apps.
- **zoneId** (`zoneId`) -- Zone ID; use value wrapper or reference to CloudflareDnsZone.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the protected hostname's domain | CloudflareDnsZone status.outputs.zone_id |
| `<company-domain>` | Company domain for group emails | Your Google Workspace domain |
| `engineering@`, `admins@` | Google Workspace group emails | Google Admin Console → Groups |
| `admin.example.com` | Hostname to protect | Your application's FQDN |

## Related Presets

- **01-company-wide-email** -- Use when allowing all company emails instead of specific groups
