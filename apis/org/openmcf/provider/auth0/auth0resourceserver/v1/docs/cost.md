# Auth0 Resource Server - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Resource servers (APIs) are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited resource servers (APIs)
- Unlimited scopes per resource server

## Cost Impact

Creating, updating, or deleting Auth0 resource server resources has no direct billing impact. There is no charge per API definition or per scope/permission defined.

The only cost driver is the number of monthly active users authenticating through your tenant. API definitions and their associated scopes are purely configuration objects.

## Token Volume Considerations

While resource servers are free to define, the access tokens issued for these APIs consume tenant resources. High-volume API usage patterns should consider:

| Factor | Impact |
|--------|--------|
| Token issuance rate | Contributes to tenant rate limits |
| Token lifetime | Shorter tokens increase issuance frequency |
| Number of scopes per token | Increases token size but not cost |
| M2M token grants | Counted against plan M2M token quota |

## Rate Limits

The Auth0 Management API enforces rate limits on resource server operations. Creating or updating resource servers counts against the Management API rate limit (varies by plan tier). This affects automation speed but not cost.
