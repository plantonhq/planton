# Auth0Client

## Overview

`Auth0Client` is a Planton deployment component that enables declarative management of Auth0 Applications. In the Auth0 dashboard, these are shown as "Applications" - they represent OAuth 2.0 clients that can authenticate users and request access to APIs.

This component abstracts the complexity of Auth0 application configuration into a simple, declarative YAML manifest that follows the Kubernetes Resource Model (KRM) structure.

## Purpose

Auth0Client addresses the need to:

- **Register OAuth Applications**: Configure client applications for authentication
- **Support Multiple Application Types**: SPAs, native apps, web apps, and M2M services
- **Configure OAuth Settings**: Callbacks, logout URLs, grant types, and token settings
- **Enable Mobile Integration**: Configure iOS and Android app settings
- **Manage Token Configuration**: Control JWT and refresh token behavior
- **Version Control**: Replace manual Auth0 dashboard configuration with declarative manifests

## Key Features

- **All Application Types**: Native, SPA, Regular Web, and Machine-to-Machine
- **OAuth Configuration**: Grant types, callbacks, CORS origins, logout URLs
- **JWT Settings**: Signing algorithms, token lifetimes, custom scopes
- **Refresh Token Control**: Rotation, expiration, and idle timeout settings
- **Mobile Support**: iOS Team ID, Android package names, native social login
- **Organization Support**: Multi-tenant application configuration
- **OIDC Compliance**: Strict OIDC-conformant mode

## Example Usage

### Single Page Application (SPA)

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: my-spa-app
  org: my-organization
  env: production
spec:
  application_type: spa
  description: My React SPA
  callbacks:
    - https://myapp.com/callback
    - http://localhost:3000/callback
  allowed_logout_urls:
    - https://myapp.com
    - http://localhost:3000
  web_origins:
    - https://myapp.com
    - http://localhost:3000
  grant_types:
    - authorization_code
    - refresh_token
  oidc_conformant: true
```

### Machine-to-Machine (M2M) Application

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: backend-api-client
  org: my-organization
spec:
  application_type: non_interactive
  description: Backend service for API access
  grant_types:
    - client_credentials
  jwt_configuration:
    lifetime_in_seconds: 86400
    alg: RS256
```

### Deploy

```bash
planton apply --manifest auth0-client.yaml \
  --auth0-provider-config auth0-creds.yaml
```

## Best Practices

1. **Use OIDC Conformant Mode**: Enable `oidc_conformant: true` for new applications
2. **Limit Grant Types**: Only enable grant types your application needs
3. **Use PKCE**: SPAs and native apps should use Authorization Code with PKCE
4. **Rotate Refresh Tokens**: Enable `rotation_type: rotating` for security
5. **Set Token Expiry**: Use appropriate token lifetimes for your security requirements
6. **Store Secrets Securely**: Client secrets should be managed securely
7. **Version Control**: Store client manifests in version control for audit trails

## Application Types

| Type | Use Case | Can Store Secret | Recommended Flow |
|------|----------|------------------|------------------|
| `native` | Mobile/Desktop apps | No | Authorization Code + PKCE |
| `spa` | Single Page Apps | No | Authorization Code + PKCE |
| `regular_web` | Server-side apps | Yes | Authorization Code |
| `non_interactive` | M2M/Backend services | Yes | Client Credentials |

## Related Documentation

- [Examples](./examples.md) - Complete working examples for all application types
- [Research Documentation](./docs/README.md) - In-depth analysis of Auth0 applications


