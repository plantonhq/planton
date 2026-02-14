# Round-Robin HTTP Pool

This preset creates a backend pool using the round-robin algorithm over HTTP. Traffic is distributed equally across all healthy members. This is the most common pool configuration for stateless web applications and APIs.

## When to Use

- Stateless web applications where any backend can handle any request
- REST APIs with no server-side session state
- Microservices behind a load balancer

## Key Configuration Choices

- **Round-robin** (`lbMethod: ROUND_ROBIN`) -- equal distribution across all members
- **HTTP protocol** -- pool communicates with backends over HTTP (match to your listener protocol)
- **No session persistence** -- each request is independently routed; add `persistence` if needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<listener-id>` | ID of the listener this pool serves as the default backend for | OpenStack console or `OpenStackLoadBalancerListener` status outputs |

## Related Presets

- **02-sticky-session** -- Use instead when client sessions must stick to the same backend
