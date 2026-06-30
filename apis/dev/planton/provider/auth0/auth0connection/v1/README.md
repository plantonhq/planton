# Auth0Connection

## Overview

`Auth0Connection` is a Planton deployment component that enables declarative management of Auth0 identity connections. Connections in Auth0 are the bridge between your applications and identity providers, allowing users to authenticate using various methods including databases, social providers, and enterprise identity systems.

This component abstracts the complexity of Auth0 connection configuration into a simple, declarative YAML manifest that follows the Kubernetes Resource Model (KRM) structure.

## Purpose

Auth0Connection addresses the need to:

- **Unify Authentication Sources**: Configure multiple identity providers (Google, GitHub, SAML, OIDC, Azure AD) through a single interface
- **Enforce Security Policies**: Apply consistent password policies, MFA requirements, and brute force protection
- **Enable Enterprise SSO**: Integrate with corporate identity providers using SAML, OIDC, or Azure AD
- **Simplify Configuration**: Replace manual Auth0 dashboard configuration with version-controlled manifests
- **Maintain Consistency**: Ensure identical authentication configuration across environments

## Key Features

- **Multiple Strategy Support**: Configure database, social, SAML, OIDC, and Azure AD connections
- **Password Policies**: Enforce password complexity, history, and dictionary checks
- **Security Controls**: Built-in brute force protection and MFA support
- **Enterprise Integration**: Full support for SAML, OIDC, and Azure AD federation
- **Client Scoping**: Control which Auth0 applications can use each connection
- **Foreign Key References**: Link to Auth0Client components by name instead of hardcoding IDs
- **Domain Discovery**: Enable identifier-first authentication flows

## Example Usage

### Database Connection (Auth0 Hosted)

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: user-database
  org: my-organization
  env: production
spec:
  strategy: auth0
  display_name: Email Sign Up
  enabled_clients:
    # Reference managed Auth0Client by name
    - value_from:
        kind: Auth0Client
        name: my-web-app
  database_options:
    password_policy: good
    brute_force_protection: true
    password_history_size: 5
    password_no_personal_info: true
    password_dictionary: true
```

### Google OAuth Connection

```yaml
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: google-login
  org: my-organization
  env: production
spec:
  strategy: google-oauth2
  display_name: Continue with Google
  enabled_clients:
    # Direct client ID
    - value: "my-web-app-client-id"
  social_options:
    client_id: "your-google-client-id.apps.googleusercontent.com"
    client_secret: "your-google-client-secret"
    scopes:
      - openid
      - profile
      - email
```

### Deploy

```bash
planton apply --manifest auth0-connection.yaml \
  --auth0-provider-config auth0-creds.yaml
```

## Best Practices

1. **Use Strong Password Policies**: For database connections, use "good" or "excellent" password policies
2. **Enable Brute Force Protection**: Always enable brute force protection for database connections
3. **Limit Enabled Clients**: Only enable connections for applications that need them
4. **Use Domain Connections**: For enterprise SSO, enable domain-based discovery for seamless user experience
5. **Store Secrets Securely**: Keep OAuth client secrets in secure credential management systems
6. **Version Control**: Store connection manifests in version control for audit trails

## Connection Strategies

| Strategy | Use Case | Required Options |
|----------|----------|------------------|
| `auth0` | Email/password authentication | `database_options` |
| `google-oauth2` | Google login | `social_options` |
| `facebook` | Facebook login | `social_options` |
| `github` | GitHub login | `social_options` |
| `samlp` | SAML enterprise SSO | `saml_options` |
| `oidc` | OpenID Connect SSO | `oidc_options` |
| `waad` | Azure AD / Entra ID | `azure_ad_options` |

## Related Documentation

- [Examples](./examples.md) - Complete working examples for all connection types
- [Research Documentation](./docs/README.md) - In-depth analysis of Auth0 connection landscape

