# Preset: OAuth with Hosted UI

**Rank**: 2 (common for web applications)

## When to Use

- Web applications that need OAuth 2.0 / OIDC authentication
- Applications using the Cognito-hosted sign-in/sign-up pages
- Staging and production environments

## What It Provides

- Email as the sign-in identifier with a reasonable password policy
- OAuth Authorization Code flow with OIDC scopes
- Cognito-hosted domain for the sign-in UI
- Token validity configured (1h access/ID, 30d refresh)
- Token revocation and user enumeration protection enabled
- Email recovery

## What You Might Add

- `mfaConfiguration: OPTIONAL` for production
- `emailConfiguration` with DEVELOPER mode for SES
- Additional callback/logout URLs for different environments
- A second client for server-side APIs (`generateSecret: true`, `client_credentials` flow)
- Custom domain with ACM certificate for branded login URLs
