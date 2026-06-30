---
title: Auth0Action
kind: Auth0Action
provider: auth0
api_version: auth0.planton.dev/v1
id_prefix: a0act
description: Manage Auth0 Actions — custom Node.js functions that execute at specific points in the Auth0 authentication and authorization pipeline.
---

# Auth0Action

Manage Auth0 Actions — custom Node.js functions that execute at specific points in the Auth0 authentication and authorization pipeline. Actions enable token enrichment, registration gating, MFA enforcement, external integrations, and custom provider implementations without modifying core Auth0 configuration.

## Provider

Auth0

## Category

Identity & Access Management

## Use Cases

- Add custom claims to tokens after login
- Restrict registration to allowed email domains
- Send Slack or email alerts on authentication events
- Enforce conditional MFA based on risk signals
- Implement custom SMS/email providers for OTP delivery
- Audit M2M token exchanges for compliance
