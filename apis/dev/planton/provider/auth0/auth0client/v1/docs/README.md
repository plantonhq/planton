# Auth0Client: Technical Research Documentation

## Introduction

Auth0 Applications (called "Clients" in the API) are the fundamental building blocks for integrating authentication into your applications. Each application registered in Auth0 represents an OAuth 2.0 client that can authenticate users and request access tokens.

This document provides comprehensive research into Auth0 applications, their configuration options, and the rationale behind Planton's Auth0Client deployment component design.

## Application Types in Auth0

### Native Applications

**Strategy:** Mobile, desktop, or CLI applications

**Characteristics:**
- Cannot securely store client secrets
- Run directly on user devices
- Use Authorization Code flow with PKCE
- Support native social login (Sign in with Apple, Facebook SDK)

**Use Cases:**
- iOS and Android mobile apps
- Desktop applications (Electron, native)
- CLI tools requiring user authentication

**Configuration Requirements:**
- Custom URL scheme callbacks (e.g., `myapp://callback`)
- Mobile-specific settings (bundle identifiers, package names)
- Native social login configuration

### Single Page Applications (SPA)

**Strategy:** JavaScript applications running in the browser

**Characteristics:**
- Cannot securely store client secrets
- All code is visible to users
- Use Authorization Code flow with PKCE
- Require CORS configuration

**Use Cases:**
- React, Angular, Vue.js applications
- Any JavaScript framework running in browser
- Progressive Web Apps (PWAs)

**Configuration Requirements:**
- Callback URLs for redirect after authentication
- Web origins for CORS
- Token handling in browser storage

### Regular Web Applications

**Strategy:** Traditional server-side web applications

**Characteristics:**
- Can securely store client secrets
- Server renders HTML and handles auth flow
- Use Authorization Code flow
- Session-based authentication

**Use Cases:**
- Node.js/Express applications
- Python/Django applications
- Ruby on Rails, PHP applications
- Any server-rendered web app

**Configuration Requirements:**
- Server-side callback URL
- Client secret management
- Session configuration

### Machine-to-Machine (M2M) Applications

**Strategy:** Backend services and APIs

**Characteristics:**
- Can securely store client secrets
- No user interaction
- Use Client Credentials flow
- Service-to-service authentication

**Use Cases:**
- Backend API services
- Scheduled jobs and cron tasks
- CI/CD pipelines
- Microservice communication

**Configuration Requirements:**
- Client credentials grant type
- API permissions (scopes)
- Token lifetime configuration

## OAuth Grant Types

### Authorization Code

The most secure flow for applications that can receive callbacks:

```
User -> App -> Auth0 (login) -> App (with code) -> Auth0 (token) -> App
```

**When to use:**
- Web applications (SPA, regular web)
- Native mobile applications
- Any user-facing application

**PKCE Enhancement:**
Required for public clients (SPA, native) to prevent authorization code interception.

### Client Credentials

For machine-to-machine authentication:

```
Service -> Auth0 (credentials) -> Service (with token)
```

**When to use:**
- Backend services
- Scheduled jobs
- Service-to-service communication

### Implicit (Legacy)

**Deprecated** - Do not use for new applications. Use Authorization Code with PKCE instead.

### Resource Owner Password

**Not recommended** - Use only for legacy system migration.

## API Grants (Client Grants)

### Understanding Grant Types vs API Grants

These are **different concepts** in Auth0:

| Concept | What it does | Field in Auth0ClientSpec |
|---------|--------------|--------------------------|
| **`grant_types`** | Which OAuth flows the client can use (e.g., `authorization_code`, `client_credentials`) | `grant_types` |
| **`api_grants`** | Which APIs the client is authorized to access and with what scopes | `api_grants` |

**Example:** An M2M application needs:
- `grant_types: ["client_credentials"]` — to use client credentials flow
- An **api_grant** — to actually call an API with specific scopes

Without the API grant, the M2M app exists but can't call any APIs.

### Configuring API Grants

Each API grant authorizes the client to call a specific API (Resource Server) with specified permissions:

```yaml
api_grants:
  - audience: "https://api.example.com/"
    scopes:
      - read:resources
      - write:resources
```

**For Auth0 Management API access:**

```yaml
api_grants:
  - audience: "https://your-tenant.us.auth0.com/api/v2/"
    scopes:
      - read:users
      - read:user_idp_tokens
      - update:users
```

### Common Management API Scopes

| Scope | Description |
|-------|-------------|
| `read:users` | Read user profiles |
| `read:user_idp_tokens` | Read identity provider tokens |
| `create:users` | Create new users |
| `update:users` | Update user profiles |
| `delete:users` | Delete users |
| `read:clients` | Read client/application details |
| `update:clients` | Update clients |
| `read:connections` | Read connection configuration |

### Organization Support in API Grants

When using Auth0 Organizations with M2M applications:

- `organization_usage: "deny"` - Organizations cannot be used (default)
- `organization_usage: "allow"` - Organizations can be used optionally
- `organization_usage: "require"` - Organizations must be specified

- `allow_any_organization: true` - Any organization can use this grant
- `allow_any_organization: false` - Must explicitly assign to organizations (default)

## Token Configuration

### JWT Configuration

**Algorithm Options:**
| Algorithm | Type | Secret Storage | Recommendation |
|-----------|------|----------------|----------------|
| RS256 | Asymmetric | Tenant keys | Recommended |
| HS256 | Symmetric | Client secret | Legacy only |
| PS256 | Asymmetric | Tenant keys | High security |

**Lifetime Considerations:**
- Shorter = More secure, more auth requests
- Longer = Better UX, higher risk if compromised
- Recommended: 1-8 hours for access tokens

### Refresh Token Configuration

**Rotation Types:**
- `non-rotating`: Same token reused (less secure)
- `rotating`: New token each refresh (recommended)

**Expiration Types:**
- `non-expiring`: Token never expires (not recommended)
- `expiring`: Token expires based on lifetime settings

**Best Practices:**
- Use rotating refresh tokens
- Set absolute lifetime (e.g., 30 days)
- Set idle timeout (e.g., 7-15 days)
- Enable leeway for clock skew

## Mobile Application Configuration

### iOS Configuration

Required for:
- Universal Links (deep linking)
- Sign in with Apple native integration
- Associated Domains

**Settings:**
- `team_id`: Apple Developer Team ID (10 characters)
- `app_bundle_identifier`: App's bundle ID (e.g., `com.example.app`)

### Android Configuration

Required for:
- App Links (deep linking)
- Intent filters for callbacks

**Settings:**
- `app_package_name`: App's package name
- `sha256_cert_fingerprints`: Signing certificate fingerprints

## Organization Support

Auth0 Organizations enable B2B SaaS scenarios:

**Usage Modes:**
- `deny`: No organization support
- `allow`: Optional organization context
- `require`: Organization must be specified

**Require Behaviors:**
- `no_prompt`: Fail if no org specified
- `pre_login_prompt`: Show org picker before login
- `post_login_prompt`: Show org picker after login

## Design Decisions

### 80/20 Scoping

Based on research, the following features cover 80% of use cases:

**In Scope:**
1. All four application types
2. Standard OAuth grant types
3. JWT configuration (algorithm, lifetime, scopes)
4. Refresh token configuration
5. Mobile app configuration
6. Organization support
7. Cross-origin authentication
8. OIDC backchannel logout
9. API grants (client grants for API access authorization)

**Out of Scope (20% edge cases):**
1. Custom database action scripts
2. Addons (SAML, WS-Fed configuration)
3. Resource server (API) configuration
4. Custom token storage

**Rationale:**
- Addons are better managed separately
- Resource servers are a separate concept
- Custom scripts require runtime

### Validation Rules

Proto validations enforce:
- Required `application_type` with allowed values
- Description max length (140 chars)
- JWT lifetime ranges (0-2592000)
- Refresh token lifetime non-negative
- Organization usage enum values
- JWT algorithm enum values

### Sensible Defaults

All optional fields have production-appropriate defaults:
- `oidc_conformant: true` (modern OIDC behavior)
- `is_first_party: true` (skip consent for own apps)
- Secure refresh token settings when specified

## Production Best Practices

### Security

1. **Use PKCE**: Always enable for SPAs and native apps
2. **Rotate Refresh Tokens**: Enable rotation for security
3. **Short Access Tokens**: 1-8 hour lifetimes
4. **HTTPS Only**: All callback URLs should be HTTPS in production
5. **Limit Scopes**: Request only necessary permissions

### Operations

1. **Environment Separation**: Different clients per environment
2. **Secret Rotation**: Rotate client secrets periodically
3. **Monitoring**: Track token issuance and failures
4. **Audit Trail**: Log all authentication events

### Token Management

1. **Store Tokens Securely**: Use secure storage mechanisms
2. **Handle Expiry**: Implement proper refresh logic
3. **Revoke on Logout**: Clear tokens on user logout
4. **Session Management**: Coordinate with backend sessions

## Common Pitfalls

### 1. Missing Callback URLs

Callbacks must be exact matches. Ensure all URLs are registered:
- Development: `http://localhost:3000/callback`
- Production: `https://app.example.com/callback`

### 2. CORS Issues

For SPAs, ensure `web_origins` includes your application's origin without trailing slashes.

### 3. Grant Type Mismatch

Using wrong grant types causes authentication failures:
- SPAs: Use `authorization_code`, not `implicit`
- M2M: Must use `client_credentials`

### 4. Token Storage in SPAs

Never store tokens in localStorage for sensitive apps. Use:
- In-memory storage with refresh rotation
- Secure HTTP-only cookies (BFF pattern)

### 5. Missing PKCE

SPAs and native apps must use PKCE. Auth0 enforces this for new apps.

## Conclusion

Auth0 Applications are the entry points for authentication in your systems. Planton's Auth0Client component provides a declarative, secure-by-default approach to managing these applications.

By supporting all application types and common OAuth configurations, this component enables teams to:
- Standardize application registration
- Version control authentication configuration
- Maintain consistency across environments
- Apply security best practices automatically

The dual IaC implementation (Pulumi and Terraform) ensures flexibility while the KRM-style manifest provides a familiar interface for platform engineers.


