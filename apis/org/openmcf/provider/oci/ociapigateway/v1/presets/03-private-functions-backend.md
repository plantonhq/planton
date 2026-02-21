# Private Functions Backend

This preset creates a private (VCN-internal) OCI API Gateway that routes requests to OCI Functions backends. The gateway is accessible only from within the VCN (or via peered VCNs, VPN, and FastConnect), making it suitable for internal microservice APIs where the gateway provides routing, logging, and timeout management while OCI Functions handle the business logic.

## When to Use

- Internal APIs consumed by other services within the same VCN (service-to-service communication)
- Serverless microservice architectures where OCI Functions implement individual API operations
- Backend-for-frontend (BFF) patterns where a private gateway aggregates multiple functions behind a unified API surface
- Event-driven processing where HTTP requests trigger serverless functions with managed routing and timeouts

## Key Configuration Choices

- **Private endpoint** (`endpointType: endpoint_type_private`) -- the gateway has no public IP and is reachable only from within the VCN. This eliminates internet exposure entirely. Clients must be in the same VCN, a peered VCN, or connected via VPN/FastConnect. The gateway's private IP is available in status outputs.
- **OCI Functions backends** (`type: oracle_functions`) -- requests are forwarded to individual OCI Functions identified by their OCIDs. The gateway handles HTTP routing, timeout enforcement, and error responses while the functions handle business logic. Each route can target a different function, enabling a microservice decomposition pattern.
- **Per-route timeouts** -- the `/process` route has a generous 120-second read timeout for long-running operations (data transformations, external API calls), while the `/query` route uses a 60-second timeout for read-heavy operations. Tune these based on the actual execution time of your functions. OCI Functions have a maximum execution time of 300 seconds.
- **NSG protection** (`networkSecurityGroupIds`) -- even though the gateway is private, NSG rules control which subnets and services can reach it. Configure ingress rules allowing TCP port 443 from specific application subnets.
- **No authentication or CORS** -- intentionally omitted for internal APIs. Service-to-service communication within a VCN typically relies on network-level isolation (NSGs, private subnets) rather than token-based authentication. Add `requestPolicies.authentication` if your internal services use JWTs for service identity.
- **No rate limiting** -- omitted because internal service-to-service traffic is typically controlled at the application level or through service mesh patterns. Add `requestPolicies.rateLimiting` if internal callers need throttling.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the gateway will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of a private subnet for the gateway endpoint | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<gateway-nsg-ocid>` | OCID of the NSG controlling access to the gateway | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<process-function-ocid>` | OCID of the function handling POST /process requests | OCI Console > Developer Services > Functions, or inspect function resources via CLI |
| `<query-function-ocid>` | OCID of the function handling GET /query requests | OCI Console > Developer Services > Functions, or inspect function resources via CLI |

## Related Presets

- **01-public-http-proxy** -- Use instead for internet-facing APIs with HTTP backends
- **02-jwt-authenticated-api** -- Use instead for public APIs requiring JWT token validation
