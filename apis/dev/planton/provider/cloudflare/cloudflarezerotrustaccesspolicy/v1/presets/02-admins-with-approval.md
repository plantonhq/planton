# Preset: Admins with approval and MFA

An `allow` policy for privileged access: it references a reusable admins group,
requires explicit approval and a purpose justification, enforces hardware-key MFA,
and uses a short 1-hour session.

## When to use

- Just-in-time privileged access with an approval workflow and step-up MFA.

## Key choices

- `include.group`: reference a `CloudflareZeroTrustAccessGroup` instead of repeating
  rules.
- `approvalRequired` + `approvalGroups`: who must approve, and how many approvals.
- `purposeJustificationRequired` + `purposeJustificationPrompt`: require a reason.
- `mfaConfig.allowedAuthenticators`: e.g. `security_key`.

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |
