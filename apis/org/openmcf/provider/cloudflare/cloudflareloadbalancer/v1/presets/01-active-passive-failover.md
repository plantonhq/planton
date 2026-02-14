# Active-Passive Failover

Two origins with steering=off: traffic goes to the first healthy origin; if it fails, traffic fails over to the second. Proxied through Cloudflare for DDoS protection and CDN. Use for high-availability with primary/backup origins.

## When to Use

- Primary + backup servers (e.g., main DB, standby)
- DR failover when primary becomes unhealthy
- Simple active-passive redundancy

## Key Configuration Choices

- **steeringPolicy: off** (`steeringPolicy: off`) -- Static/failover; first healthy origin gets all traffic.
- **proxied: true** (`proxied: true`) -- Recommended; traffic through Cloudflare.
- **healthProbePath** (`healthProbePath: /`) -- Path for health checks; default is /.
- **zoneId** (`zoneId`) -- Zone ID; use value wrapper or reference to CloudflareDnsZone.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID containing hostname | CloudflareDnsZone status.outputs.zone_id |
| `app.example.com` | Hostname for the load balancer | Your application FQDN |
| `192.0.2.1`, `192.0.2.2` | Primary and secondary origin IPs | Your server IPs or hostnames |

## Related Presets

- **02-geographic-routing** -- Use when routing by geography instead of failover
- **03-weighted-ab-testing** -- Use when splitting traffic by weight
