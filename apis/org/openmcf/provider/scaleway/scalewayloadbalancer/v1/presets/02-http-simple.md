# Simple HTTP Load Balancer

This preset creates a minimal Scaleway Load Balancer with a single HTTP frontend and TCP health checks. No TLS certificates or Private Network attachment are configured, making this the simplest possible LB setup. Suitable for development, internal services, or scenarios where TLS is terminated elsewhere.

## When to Use

- Development and testing environments where HTTPS is not required
- Internal services behind a VPN or within a Private Network that do not need public TLS
- Quick prototyping to validate backend server connectivity

## Key Configuration Choices

- **Small tier** (`type: LB-S`) -- minimal cost; upgrade as traffic grows
- **No Private Network** -- the LB reaches backends via public IPs; add `privateNetworkId` for production topologies
- **Single backend server** -- sufficient for development; add more IPs for availability
- **TCP health check** (`healthCheck.type: tcp`) -- simplest check; upgrade to `http` with a `/health` endpoint for production
- **Backend port 8080** -- common development application port; adjust to match your application's listen port
- **Frontend port 80** -- standard HTTP; no certificates needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<backend-server-ip>` | IP address of the backend server | Scaleway console or `ScalewayInstance` status outputs |

## Related Presets

- **01-https-letsencrypt** -- Use instead for production with automatic TLS, Private Network attachment, and HTTP health checks
