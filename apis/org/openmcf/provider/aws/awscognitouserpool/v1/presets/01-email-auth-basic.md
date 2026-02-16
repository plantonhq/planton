# Preset: Email Auth Basic

**Rank**: 1 (most common starting point)

## When to Use

- Getting started with Cognito
- Development and testing environments
- Simple applications needing email-based sign-up and sign-in

## What It Provides

- Email as the sign-in identifier
- Auto-verified email addresses
- Password recovery via email
- Single app client with SRP auth (secure) and refresh token support
- No MFA, no domain, no custom attributes

## What You Might Add

- `passwordPolicy` for stronger password requirements
- `mfaConfiguration: OPTIONAL` with `softwareTokenMfaEnabled: true`
- `domain` for hosted UI login pages
- Additional clients for mobile or server-side applications
