# HTTPS Load Balancer with SSL Termination

This preset creates a load balancer that terminates TLS on port 443 and forwards traffic to backend Droplets over HTTP on port 80. Uses tag-based targeting so any Droplet with the `web` tag in the VPC is automatically added. Health checks ensure only healthy backends receive traffic.

## When to Use

- Production web applications requiring HTTPS
- SSL termination at the load balancer (certificate managed separately)
- Tag-based scaling: add/remove Droplets with the tag to scale
- Multiple backend Droplets behind a single public endpoint

## Key Configuration Choices

- **HTTPS to HTTP** (`entryPort: 443`, `entryProtocol: https`, `targetPort: 80`) -- TLS terminates at the LB; backends serve plain HTTP for simplicity.
- **Certificate** (`certificateName`) -- use the name of a `DigitalOceanCertificate` (Let's Encrypt or custom) resource; prefer name over ID for stable IaC.
- **Tag-based targeting** (`dropletTag: web`) -- all Droplets with tag `web` in the VPC are attached; no manual droplet ID management.
- **Health check** (`path: /health`, `checkIntervalSec: 10`) -- backends must respond 2xx on `/health`; adjust path to match your app.
- **VPC required** (`vpc`) -- load balancer must be in a VPC; same VPC as your Droplets.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `<certificate-name>` | Name of the SSL certificate in DigitalOcean | `DigitalOceanCertificate` resource `certificate_name` or status |
| `nyc3` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **02-http-basic** -- Use when HTTPS is not required (dev/staging)
