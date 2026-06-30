# Cloudflare KV Namespace

## Overview

`CloudflareKvNamespace` is a resource for creating and managing Cloudflare Workers KV namespaces—globally distributed key-value storage optimized for read-heavy workloads at the edge.

Cloudflare Workers KV is a globally replicated data store that caches key-value pairs at Cloudflare's 300+ edge locations. It's designed for configuration data, feature flags, session tokens, and cached API responses that change infrequently but need to be accessed quickly from anywhere in the world.

## Key Features

- **Global Distribution**: Data is automatically cached at edge locations worldwide for sub-millisecond read latency
- **Simple API**: Create a namespace, bind it to your Worker, and start reading/writing keys
- **Usage-Based Pricing**: Generous free tier (100K reads/day, 1K writes/day, 1 GB storage) with predictable pay-per-operation pricing beyond that
- **Infrastructure as Code**: Manage namespaces declaratively with Planton's protobuf-based API

## Use Cases

**Ideal for:**
- Feature flags and configuration that changes infrequently
- Cached API responses and computed results
- User session tokens and preferences
- Static data that needs global low-latency access

**Not ideal for:**
- Counters or frequently updated data (use Durable Objects instead)
- Large binary files (use R2 Storage instead)
- Scenarios requiring immediate read-after-write consistency

## API Specification

### CloudflareKvNamespaceSpec

The specification follows the **80/20 principle**—exposing only the most commonly needed fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace_name` | string | Yes | Human-readable name for the KV namespace. Must be unique within the Cloudflare account. Limited to 64 characters. |
| `account_id` | string | Yes | The Cloudflare account ID (32 hex characters) that owns the namespace. |

A KV namespace carries only an account and a title. There is no per-namespace TTL
or description; TTL is applied per write. Seed entries with `CloudflareWorkersKvPair`.

### Stack Outputs

After successful deployment, the following outputs are available:

| Field | Description |
|-------|-------------|
| `namespace_id` | The unique identifier (UUID) of the created KV namespace. Use this to bind the namespace to Worker scripts. |

## How It Works

This resource uses **Pulumi** (with Go) to provision Cloudflare Workers KV namespaces via the Cloudflare API. The implementation:

1. **Creates the namespace** with the specified name
2. **Returns the namespace ID** for use in Worker bindings
3. **Manages lifecycle** (updates, deletes) through Pulumi state

The namespace is empty after creation—you populate keys and values through:
- Cloudflare Workers runtime code (`KV.put()`, `KV.get()`)
- Wrangler CLI (`wrangler kv:bulk put`)
- Cloudflare REST API

## Performance Characteristics

Workers KV uses a tiered caching architecture:

- **Cold reads**: First read from a new region fetches from central storage (tens to hundreds of milliseconds)
- **Hot reads**: Subsequent reads from the same region hit edge cache (sub-millisecond)
- **Write propagation**: Updates go to central storage immediately but propagate to edge caches over 60+ seconds (eventual consistency)

Design your application to tolerate this eventual consistency model.

## Pricing

**Free Tier:**
- 100,000 read operations/day
- 1,000 write operations/day
- 1 GB storage
- Limits reset daily at 00:00 UTC

**Paid Plans (Bundled Workers, $5/month):**
- 10 million reads/month included
- 1 million writes/month included
- $0.50 per million additional reads
- $5.00 per million additional writes
- $0.50/GB-month for storage beyond 1 GB

## Integration with Workers

After creating a namespace, bind it to your Worker in `wrangler.toml`:

```toml
[[kv_namespaces]]
binding = "CONFIG"
id = "<namespace_id>"  # From stack outputs
```

Then access it in your Worker code:

```javascript
// Read a key
const value = await CONFIG.get("feature_flags");

// Write a key with optional TTL
await CONFIG.put("user_session", sessionData, { expirationTtl: 3600 });

// Delete a key
await CONFIG.delete("expired_token");
```

## Related Resources

- **Cloudflare Worker**: Deploy Worker scripts that use KV namespaces
- **Cloudflare Durable Objects**: For data requiring strong consistency or high write rates
- **Cloudflare R2 Storage**: For large binary files and bulk data

## Further Reading

For comprehensive deployment guidance, architecture patterns, and best practices, see [docs/README.md](./docs/README.md).

## References

- [Cloudflare Workers KV Documentation](https://developers.cloudflare.com/kv/)
- [How Workers KV Works](https://developers.cloudflare.com/kv/concepts/how-kv-works/)
- [Workers KV Pricing](https://developers.cloudflare.com/workers/platform/pricing/#workers-kv)
- [Wrangler CLI KV Commands](https://developers.cloudflare.com/workers/wrangler/commands/#kv)

