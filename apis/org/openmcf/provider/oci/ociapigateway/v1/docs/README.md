# OciApiGateway — Design Notes

## Design Rationale

OciApiGateway bundles the API Gateway resource and a single API Deployment into one component. This is the most complex OCI component in terms of API surface, reflecting the gateway's role as the single entry point for API traffic.

### Why bundle gateway and deployment?

In practice, a gateway almost always has exactly one deployment. The deployment cannot exist without a gateway, and a gateway without a deployment is useless. Bundling them avoids the boilerplate of two separate manifests and ensures the deployment is always created with the correct gateway reference and compartment. Users who need multiple deployments on one gateway (rare) can create additional deployments via the OCI API directly.

### Why include only JWT authentication?

OCI API Gateway supports multiple authentication types (JWT, Custom/Functions-based, Token). JWT is by far the most common pattern for API authentication — it covers Auth0, Okta, OCI IDCS, and any OAuth2/OIDC provider that exposes a JWKS endpoint. Functions-based authentication adds significant complexity (function invocation, custom response mapping) and is better suited for a v2 iteration if demand exists.

### Why support both remote JWKS and static keys?

Remote JWKS is preferred for production (automatic key rotation) but requires network access to the IdP. Static keys are useful for testing, air-gapped environments, or IdPs that don't expose a JWKS endpoint. Supporting both covers the full range of real-world deployment scenarios.

### Why is route order significant?

OCI API Gateway evaluates routes in the order they appear in the deployment specification. First match wins. This is a deliberate design choice that gives operators explicit control over routing precedence — especially important when paths overlap (e.g., `/users/{id}` vs `/users/me`).

### Why are CORS, rate limiting, and authentication at the deployment level?

These are deployment-level policies in the OCI API. They apply to all routes before route matching occurs. Individual routes can override authorization behavior but not authentication or rate limiting. This matches how most APIs work: one authentication scheme, one rate limit, applied uniformly.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Bundle gateway + deployment | Single manifest; no orphaned resources | Cannot have multiple deployments per gateway via this component |
| JWT-only authentication | Covers 90%+ of use cases; simpler spec | Custom (Functions-based) auth requires separate tooling |
| Route order matters | Explicit control over precedence | Users must understand ordering semantics |
| Deployment-level policies | Consistent behavior across all routes | Cannot have different rate limits per route |
| Flat Backend message | One message for all backend types | Irrelevant fields visible for each type |

## Resource Graph

```
OciApiGateway
├── oci_apigateway_gateway (always)
│   └── network_security_group_ids (0..N)
└── oci_apigateway_deployment (always, DependsOn gateway)
    ├── specification
    │   ├── logging_policies (optional)
    │   ├── request_policies (optional)
    │   │   ├── authentication (JWT, optional)
    │   │   ├── cors (optional)
    │   │   └── rate_limiting (optional)
    │   └── routes (1..N)
    │       ├── backend (http | oracle_functions | stock_response)
    │       ├── authorization (optional per-route)
    │       └── logging_policies (optional per-route)
    └── outputs: gateway_id, hostname, deployment_endpoint
```

## Deferred from v1

- **ca_bundles, ip_mode, IPv4/IPv6 config** — advanced networking options with very low adoption.
- **response_cache_details** — requires external Redis; infrastructure concern outside the gateway.
- **Custom (Functions-based) authentication** — complex; Functions-based auth invokes a function for each request.
- **Token authentication** — newer JWT variant; overlaps with existing JWT support.
- **validation_failure_policy** — OAuth2 redirect flows; adds significant complexity.
- **dynamic_authentication, mutual_tls, usage_plans** — advanced deployment policies.
- **body_validation, header/query transformations** — route-level policies that add spec complexity.
- **dynamic_routing, oauth2_logout backends** — niche backend types.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags on both the gateway and deployment resources:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciApiGateway` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
