# Preset: Google OAuth

Configures Google as a social identity provider for a Cognito User Pool.

## What This Creates

- An identity provider registration for Google OAuth 2.0
- Standard attribute mapping (email, username)

## Variables to Replace

- `${USER_POOL_ID}` -- Cognito User Pool ID (e.g., `us-east-1_Ab1Cd2EfG`)
- `${GOOGLE_CLIENT_ID}` -- Google OAuth 2.0 client ID from Google Cloud Console
- `${GOOGLE_CLIENT_SECRET}` -- Google OAuth 2.0 client secret

## After Deployment

Add `"Google"` to the `supportedIdentityProviders` list in your User Pool Client configuration to enable Google sign-in.

## Scopes

The `email profile openid` scopes request:
- `email` -- user's email address
- `profile` -- basic profile information (name, picture)
- `openid` -- OpenID Connect standard claims
