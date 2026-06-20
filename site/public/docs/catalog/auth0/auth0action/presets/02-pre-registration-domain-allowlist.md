---
title: "Preset: Pre-Registration Domain Allowlist"
description: "Restrict user registration to specific email domains. Users with disallowed domains receive a clear denial message and are blocked from creating an account."
type: "preset"
rank: "02"
presetSlug: "02-pre-registration-domain-allowlist"
componentSlug: "auth0action"
componentTitle: "Auth0Action"
provider: "auth0"
icon: "package"
order: 2
---

# Preset: Pre-Registration Domain Allowlist

## Pattern

Restrict user registration to specific email domains. Users with disallowed domains receive a clear denial message and are blocked from creating an account.

## What It Does

- Reads the comma-separated list of allowed domains from the `ALLOWED_DOMAINS` secret.
- Extracts the domain from the registering user's email.
- Denies registration with a user-friendly message if the domain is not in the allow list.

## When to Use

- B2B applications where only employees of specific companies should register.
- Internal tools where registration should be limited to corporate email addresses.
- Any scenario where you want to gate registration without modifying the Universal Login page.

## Customization

- Update the `ALLOWED_DOMAINS` secret value to include your organization's email domains.
- Add subdomains as separate entries (e.g., `eng.example.com,sales.example.com`).
- For more complex logic, replace the domain check with a regex or external API call.
- Combine with `api.user.setUserMetadata()` to tag approved users during registration.
