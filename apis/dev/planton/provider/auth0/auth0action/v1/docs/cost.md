# Auth0 Action - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Actions are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited Actions
- Unlimited Action triggers (post-login, pre-registration, etc.)

## Cost Impact

Creating, updating, or deleting Auth0 Action resources has no direct billing impact. There is no charge per Action definition, per deployment, or per trigger binding.

The only cost driver is the number of monthly active users authenticating through your tenant. Actions execute as part of the authentication pipeline at no additional cost.

## Execution Considerations

While Actions are free to define, they affect authentication latency. Each Action in a trigger's pipeline adds execution time to the authentication flow:

| Factor | Impact |
|--------|--------|
| Action code complexity | Directly increases login latency |
| External API calls in Actions | Adds network round-trip time |
| Number of Actions per trigger | Cumulative latency increase |
| Action timeout | 20 seconds maximum per Action |

## Rate Limits

The Auth0 Management API enforces rate limits on Action operations. Deploying Actions (creating new versions) counts against the Management API rate limit. Action executions during authentication flows are governed by separate per-tenant limits.

## npm Dependencies

Actions can include npm packages. There is no cost for npm dependencies, but each Action deployment must resolve and bundle its dependencies, which affects deployment time.
