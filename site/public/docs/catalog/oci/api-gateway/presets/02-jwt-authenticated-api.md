---
title: "JWT Authenticated API"
description: "This preset creates a public OCI API Gateway with JWT token validation, per-route authorization, rate limiting, CORS, and logging. The gateway validates Bearer tokens against a remote JWKS endpoint..."
type: "preset"
rank: "02"
presetSlug: "02-jwt-authenticated-api"
componentSlug: "api-gateway"
componentTitle: "API Gateway"
provider: "oci"
icon: "package"
order: 2
---

# JWT Authenticated API

This preset creates a public OCI API Gateway with JWT token validation, per-route authorization, rate limiting, CORS, and logging. The gateway validates Bearer tokens against a remote JWKS endpoint (compatible with Auth0, OCI IDCS, Okta, Keycloak, and any OpenID Connect provider) before forwarding requests to the backend. Routes demonstrate three authorization levels: anonymous (health check), authentication-only (user endpoints), and scope-based (admin endpoints).

## When to Use

- Production APIs protected by an identity provider (IdP) that issues JWT access tokens
- APIs with mixed access levels: some endpoints public, some requiring authentication, some requiring specific scopes/roles
- Multi-tenant APIs where rate limiting per client IP prevents abuse
- Any API where token validation should be offloaded from application code to the gateway layer

## Key Configuration Choices

- **JWT authentication with remote JWKS** (`authentication.publicKeys.type: remote_jwks`) -- the gateway fetches the IdP's JSON Web Key Set to verify token signatures. This supports automatic key rotation by the IdP without gateway redeployment. The JWKS response is cached for 1 hour (`maxCacheDurationInHours: 1`). For static key setups, use `static_keys` type instead.
- **Issuer and audience validation** (`issuers`, `audiences`) -- tokens must contain an `iss` claim matching the configured issuer and an `aud` claim matching the configured audience. This prevents tokens issued for other APIs or by other IdPs from being accepted.
- **Clock skew tolerance** (`maxClockSkewInSeconds: 60`) -- allows up to 60 seconds of clock drift when validating `exp`, `nbf`, and `iat` claims. This prevents token rejection due to minor clock differences between the IdP, client, and gateway.
- **Scope verification** (`verifyClaims` with `scope`) -- tokens must contain a `scope` claim. Individual routes further restrict which scopes are required (e.g., the `/admin` route requires the `admin` scope).
- **Three authorization levels** -- demonstrates the three route authorization patterns: `anonymous` (no token required, for health checks), `authentication_only` (valid token required, any scope), and `any_of` (valid token with specific scope required). This is the most common authorization hierarchy for production APIs.
- **Rate limiting** (`rateLimiting` with `client_ip`, 100 req/s) -- limits each client IP to 100 requests per second. This protects the backend from abuse and ensures fair resource sharing. The `client_ip` rate key is appropriate for APIs where clients have distinct IPs. Use `total` rate key for aggregate limiting across all clients.
- **CORS and logging** -- same configuration as the public HTTP proxy preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the gateway will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid>` | OCID of a public subnet for the gateway endpoint | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<gateway-nsg-ocid>` | OCID of the NSG allowing HTTPS ingress to the gateway | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<token-issuer-url>` | JWT issuer URL (e.g., `https://your-tenant.auth0.com/`) | Your IdP tenant configuration |
| `<api-audience>` | Expected audience claim in tokens (e.g., `https://api.example.com`) | Your IdP API/application configuration |
| `<jwks-uri>` | JWKS endpoint URL (e.g., `https://your-tenant.auth0.com/.well-known/jwks.json`) | Your IdP discovery document (`/.well-known/openid-configuration`) |
| `<frontend-domain>` | Origin domain for CORS (e.g., `app.example.com`) | Your frontend deployment configuration |
| `<backend-url>` | Backend service URL (e.g., `https://backend.example.com:8080`) | Your backend service deployment |

## Related Presets

- **01-public-http-proxy** -- Use instead when authentication is handled at the application layer and the gateway only needs CORS and logging
- **03-private-functions-backend** -- Use instead for internal APIs backed by OCI Functions with no public internet exposure
