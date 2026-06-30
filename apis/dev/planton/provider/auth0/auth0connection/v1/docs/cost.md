# Auth0 Connection - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Authentication connections are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited social connections
- Unlimited database connections

## Cost Impact

Creating, updating, or deleting Auth0 connection resources has no direct billing impact. There is no charge per connection regardless of type (Database, Social, Enterprise, SAML, OIDC, Azure AD).

The only cost driver is the number of monthly active users authenticating through your tenant.

## Enterprise Connection Considerations

While connections themselves are free objects, certain enterprise connection types (SAML, OIDC, Azure AD) are only available on paid plans:

| Connection Type | Minimum Plan |
|----------------|--------------|
| Database | Free |
| Social (Google, GitHub, etc.) | Free |
| SAML | Essentials |
| OIDC | Essentials |
| Azure AD | Essentials |
| LDAP/AD | Enterprise |

## Social Connection Limits

The free plan allows unlimited social connections, but each social provider requires its own OAuth app registration (e.g., Google Cloud Console, GitHub OAuth App). Auth0 provides default development keys for testing, but production deployments should use custom keys.
