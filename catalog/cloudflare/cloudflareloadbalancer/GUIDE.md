# CloudflareLoadBalancer guide

Operational judgment for configuring the load balancer well. The README covers what each field is; this covers how the pieces interact.

## The paid add-on gates everything

Cloudflare Load Balancing is an account add-on. Without it, every call under the Load Balancing API -- pools and monitors included -- returns `403`. Enable the subscription before deploying any of the three kinds.

## Steering policies interact with other fields

- `off` uses `defaultPools` order only; geo maps and `randomSteering` are ignored.
- `geo` needs at least one of `regionPools` / `countryPools` / `popPools`; anything unmatched falls back down the chain (PoP -> country -> region -> default). `popPools` is Enterprise-only.
- `random`, `least_outstanding_requests`, and `least_connections` all read `randomSteering` weights -- the weights scale randomness, outstanding requests, or open connections respectively.
- `proximity` needs `latitude`/`longitude` set on the pools; `dynamic_latency` needs health-check RTT data, so pools must have a monitor.
- `locationStrategy` only affects NON-proxied (gray-cloud) traffic; for proxied traffic Cloudflare already knows the client location from the connecting PoP.

## Session affinity constraints

- Affinity works on proxied load balancers; `cookie`/`ip_cookie`/`header` all create per-client sessions with `sessionAffinityTtl` expiry (cookie modes accept 1800-604800s, header mode 30-3600s).
- `header` mode requires `sessionAffinityAttributes.headers`; plan limits apply (1 header on non-Enterprise, 5 on Enterprise).
- A `samesite: None` cookie cannot combine with `secure: Never` -- the spec enforces this at validation time (top level and rule overrides both).
- `zeroDowntimeFailover: sticky` is not supported for header-mode affinity.

## Rules: inherit vs override

Every field inside `rules[].overrides` is optional. Unset means "inherit the load balancer's setting"; an explicit value -- including `none`/`off`/`0` -- is a real override. This matters most for `sessionAffinity: none` (switch affinity OFF for matched traffic) and `steeringPolicy: "off"` (force static failover). Quote `"off"` in YAML: bare `off` parses as boolean false.

A `fixedResponse` rule answers at the edge and always terminates evaluation. Leave `priority` unset everywhere and let list order decide, or set it everywhere -- mixing the two makes ordering hard to reason about.

## The rule count is a plan limit

How many rules one load balancer may carry is set by the account's Load Balancing subscription tier: the entry tier allows exactly ONE rule, and the write fails with `400` code `1002` "rule count N exceeds limit M" the moment the list exceeds the tier's cap (measured live on Basic). Design rule sets against the tier you actually pay for -- splitting behavior across multiple load balancers is the workaround when the cap binds.

## The hostname is the full DNS name

`hostname` must be fully qualified (`app.example.com`), not a bare label (`app`) -- Cloudflare rejects labels with `400` code `1002` "Invalid load balancer name: invalid hostname" (measured live). The name should sit inside the zone `zoneId` points at.

## TTL only applies gray-cloud

`ttl` (and its rule override) is the DNS TTL for NON-proxied load balancers. On a proxied (orange-cloud) load balancer Cloudflare answers with its own edge IPs and the field is rejected by the API.
