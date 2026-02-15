# Preset: Enterprise OIDC

Configures a generic OIDC provider for enterprise single sign-on. Works with
Okta, Auth0, Azure AD, Keycloak, and any OIDC-compliant identity provider.

## What This Creates

- An identity provider registration for a generic OIDC provider
- Attribute mapping for email, username, and name fields

## Variables to Replace

- `${USER_POOL_ID}` -- Cognito User Pool ID (e.g., `us-east-1_Ab1Cd2EfG`)
- `${OIDC_CLIENT_ID}` -- OIDC client ID registered with your identity provider
- `${OIDC_CLIENT_SECRET}` -- OIDC client secret (omit for public clients using PKCE)
- `${OIDC_ISSUER_URL}` -- OIDC issuer URL (e.g., `https://dev-123456.okta.com`)

## After Deployment

Add the provider name (`"CorpSSO"`) to the `supportedIdentityProviders` list
in your User Pool Client configuration.

## Auto-Discovery

Cognito automatically discovers authorization, token, userinfo, and JWKS
endpoints from the issuer's `.well-known/openid-configuration` document.
Override individual endpoints only if your provider does not support standard
discovery.
