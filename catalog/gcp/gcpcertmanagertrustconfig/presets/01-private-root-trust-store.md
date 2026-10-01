# Private Root Trust Store

This preset creates a global trust config holding one private root CA and its issuing intermediate — the standard shape for mutual TLS on an Application Load Balancer, where every client certificate your CA issues is accepted and nothing else is.

## When to Use

- Clients (services, devices, partners) carry certificates from a CA you control
- A server TLS policy on a global external or cross-region internal load balancer needs a trust config to name
- Clients that send only their leaf certificate, so the load balancer must complete the chain itself

## Key Configuration Choices

- **Root as trust anchor, issuing CA as intermediate** — the chain completes even when clients omit their intermediates
- **Global location (default)** — pairs with global load balancers; set `location` to the region for a regional one
- **`deletionPolicy: PREVENT`** — destroy fails rather than locking every client out of a live service

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<root-ca-pem-body>` | The base64 body of the root CA certificate | Your CA's published root certificate |
| `<intermediate-ca-pem-body>` | The base64 body of the issuing CA certificate | Your CA's published issuing certificate |

## Related Presets

- **02-allowlisted-device-certificates** — accept a few self-signed certificates without trusting any CA
