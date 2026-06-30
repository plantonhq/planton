# Preset: Production Multi-Client

**Rank**: 3 (production-grade with security hardening)

## When to Use

- Production environments serving real users
- Applications requiring both browser-based (SPA) and server-to-server (M2M) authentication
- Organizations that need MFA, SES email, and deletion protection

## What It Provides

- Email as the sign-in identifier with strong password policy (12 chars, all character types)
- Optional MFA with TOTP authenticator app support
- SES-based email delivery (no 50/day sandbox limit)
- Deletion protection enabled
- Two app clients:
  - **web-spa**: Public client for browser-based SPA (Authorization Code flow, no secret)
  - **api-server**: Confidential client for server-to-server auth (Client Credentials flow, with secret)
- Token revocation and user enumeration protection
- Cognito-hosted domain for OAuth endpoints

## What You Might Add

- `lambdaConfig.preSignUp` for custom signup validation
- `lambdaConfig.preTokenGeneration` to inject custom claims (tenant_id, roles)
- `customAttributes` for application-specific user data
- Custom domain with ACM certificate
- Additional clients for mobile applications
