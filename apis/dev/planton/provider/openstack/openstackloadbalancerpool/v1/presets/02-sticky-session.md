# Sticky Session Pool (HTTP Cookie)

This preset creates a backend pool with round-robin distribution and HTTP cookie-based session persistence. Octavia inserts and tracks a cookie so that subsequent requests from the same client are routed to the same backend member. Use this for applications with server-side session state.

## When to Use

- Web applications with server-side sessions (shopping carts, user state)
- WebSocket connections that must maintain affinity to a specific backend
- Any HTTP service where session continuity matters

## Key Configuration Choices

- **Round-robin with persistence** -- new sessions are distributed evenly; subsequent requests stick
- **HTTP_COOKIE** (`persistence.type: HTTP_COOKIE`) -- Octavia manages the cookie automatically (no application changes needed)
- **HTTP protocol** -- pool communicates with backends over HTTP

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<listener-id>` | ID of the listener this pool serves as the default backend for | OpenStack console or `OpenStackLoadBalancerListener` status outputs |

## Related Presets

- **01-round-robin** -- Use instead when no session persistence is needed
