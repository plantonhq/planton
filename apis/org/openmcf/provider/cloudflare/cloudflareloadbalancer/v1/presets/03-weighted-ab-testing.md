# Weighted A/B Testing

Two origins with steering=random and different weights: traffic is distributed by weight (e.g., 70% control, 30% variant). Use for A/B tests, canary deployments, or gradual rollouts.

## When to Use

- A/B testing (control vs variant)
- Canary deployments (e.g., 90% old, 10% new)
- Gradual traffic shift between versions

## Key Configuration Choices

- **steeringPolicy: random** (`steeringPolicy: random`) -- Distributes traffic by origin weight.
- **weight** (`origins[].weight`) -- Relative traffic share; e.g., 70 and 30 = 70%/30% split.
- **zoneId** (`zoneId`) -- Zone ID; use value wrapper or reference to CloudflareDnsZone.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID containing hostname | CloudflareDnsZone status.outputs.zone_id |
| `app.example.com` | Hostname for the load balancer | Your application FQDN |
| `control.example.com`, `variant.example.com` | Control and variant origin addresses | Your A/B backend servers |
| `70`, `30` | Traffic weights (must sum for proportion) | Desired split (e.g., 90/10 for canary) |

## Related Presets

- **01-active-passive-failover** -- Use for failover instead of weighted split
- **02-geographic-routing** -- Use when routing by geography
