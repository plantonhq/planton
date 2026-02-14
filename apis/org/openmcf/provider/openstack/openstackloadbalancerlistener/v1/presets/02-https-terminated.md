# HTTPS Listener with TLS Termination

This preset creates an HTTPS listener on port 443 that terminates TLS at the load balancer. Traffic is decrypted at the Octavia amphora and forwarded to backend pools as plain HTTP. The `X-Forwarded-For` and `X-Forwarded-Proto` headers are inserted so backends know the original client IP and protocol.

## When to Use

- Public-facing web applications that need HTTPS
- APIs requiring TLS encryption from clients to the load balancer
- Any production HTTP service that should be encrypted in transit

## Key Configuration Choices

- **TERMINATED_HTTPS** (`protocol: TERMINATED_HTTPS`) -- TLS termination at the LB; backends receive plain HTTP
- **Port 443** -- standard HTTPS port
- **TLS certificate** (`defaultTlsContainerRef`) -- references a Barbican secret container with the certificate and private key
- **Forwarded headers** -- `X-Forwarded-For` and `X-Forwarded-Proto` inserted for backend visibility

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<loadbalancer-id>` | ID of the load balancer to attach this listener to | OpenStack console or `OpenStackLoadBalancer` status outputs |
| `<barbican-tls-container-uri>` | URI of the Barbican TLS secret container (e.g., `https://barbican.example.com/v1/containers/<uuid>`) | Barbican API or Horizon Secrets panel |

## Related Presets

- **01-http** -- Use instead for unencrypted HTTP traffic
- **03-tcp-passthrough** -- Use instead to pass encrypted traffic through without termination
