# Auth0 Client - Cost

## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Auth0 Applications (clients) are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited social connections
- Unlimited applications (clients)

## Cost Impact

Creating, updating, or deleting Auth0 client resources has no direct billing impact. There is no charge per application regardless of type (SPA, Web, M2M, Native).

The only cost driver is the number of monthly active users authenticating through your tenant. If your MAU count stays within the free tier, all client resources remain free.

## M2M Token Considerations

Machine-to-machine (M2M) applications consume M2M token grants. The free plan includes 1,000 M2M tokens per month. Exceeding this limit requires a paid plan. Each M2M client credentials grant counts against this quota regardless of which M2M application issues the request.

## Paid Plan Thresholds

| Plan | MAU Limit | M2M Tokens/Month |
|------|-----------|-------------------|
| Free | 25,000 | 1,000 |
| Essentials | Custom | 5,000 |
| Professional | Custom | 10,000 |
| Enterprise | Custom | Negotiated |
