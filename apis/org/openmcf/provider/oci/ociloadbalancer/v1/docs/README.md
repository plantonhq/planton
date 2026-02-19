# OCI Application Load Balancer: Design Rationale and Research

## Introduction

The OciLoadBalancer component manages OCI's Application Load Balancer (Layer 7) — the most complex component in the OCI provider, bundling 7 sub-resource types (load balancer, backend sets, backends, listeners, certificates, hostnames, rule sets) into a single atomic deployment unit. The spec defines 17 nested message types with 3 enums (~608 lines of proto), reflecting the inherent complexity of a full-featured Layer 7 load balancer. This document explains the design decisions that shaped the component.

## Why Bundle All Sub-Resources

The bundling decision follows the same philosophy as OciVcn (which bundles gateways). The key argument: **a load balancer's sub-resources are useless independently.**

1. **Backend sets are scoped to a load balancer.** A backend set without a load balancer cannot exist. The Terraform resource `oci_load_balancer_backend_set` requires a `load_balancer_id`. Making backend sets a separate OpenMCF component would require a mandatory foreign key reference to the load balancer, and the backend set would have no outputs useful to anything other than the load balancer's own listeners.

2. **Listeners depend on everything else.** A listener references a backend set, optionally references hostnames, rule sets, and certificates — all of which are scoped to the same load balancer. Creating listeners as a separate component would require foreign key references to the load balancer plus knowledge of the backend set names, hostname names, and rule set names within that load balancer. This creates a fragile coupling that is better expressed as a single manifest.

3. **Atomic deployment is the common case.** Platform engineers deploying a load balancer expect to define the full stack (LB + backends + listeners + SSL) in one pass, not orchestrate 5+ separate resources with cross-references. The "I just want an HTTPS load balancer" scenario should be one manifest, one command.

4. **The Pulumi module already manages ordering.** Resources are created in sequence (LB → certificates → backend sets → hostnames → rule sets → listeners) with explicit `DependsOn` relationships. This ordering is an implementation detail that users should not need to manage.

**Trade-off acknowledged:** The bundled approach means the spec is large (17 nested messages). This is mitigated by progressive disclosure in documentation — the Quick Start uses only the required fields, and optional features (certificates, hostnames, rule sets, session persistence) are documented in their own sub-sections.

## Why Session Persistence Uses oneof

The `session_persistence_config` field on BackendSet uses a proto `oneof`:

```protobuf
oneof session_persistence_config {
    LbCookieSessionPersistenceConfig lb_cookie_session_persistence = 7;
    SessionPersistenceConfig app_cookie_session_persistence = 8;
}
```

This enforces mutual exclusivity at the schema level. The alternatives considered:

1. **Two separate optional fields without oneof.** Simpler proto, but allows users to set both fields simultaneously. The OCI API rejects this, so the error surfaces at deployment time rather than validation time. Oneof catches the mistake earlier.

2. **A single message with a "mode" enum.** Would require a union of all cookie configuration fields in one message. LB-managed cookies have fields (domain, isHttpOnly, isSecure, maxAgeInSeconds, path) that application-managed cookies do not. A single message would have confusing "only applies when mode is X" documentation for half its fields.

3. **Chosen: oneof.** The proto `oneof` maps cleanly to YAML — users set either `lbCookieSessionPersistence` or `appCookieSessionPersistence`, never both. Validation rejects manifests with both fields set. Each message has only the fields relevant to its mode.

## Why SslConfiguration Is a Shared Message

The `SslConfiguration` message is used in two contexts:

- **Backend set SSL** — encrypts traffic between the load balancer and backend servers (re-encryption)
- **Listener SSL** — terminates SSL from clients to the load balancer

The OCI API uses nearly identical fields for both contexts. Sharing one message type avoids duplicating 9 fields across two messages with only one difference: `has_session_resumption` is only relevant for listener SSL (it controls whether the load balancer caches TLS sessions for client connections). The field is documented as "Only applicable in listener SSL context. Ignored for backend set SSL."

The alternative — two separate messages `BackendSetSslConfiguration` and `ListenerSslConfiguration` — would duplicate 8 identical fields. The shared message trades one "ignored in this context" footnote for avoiding significant proto duplication. The Pulumi module does use separate Go types (`BackendSetSslConfigurationArgs` vs `ListenerSslConfigurationArgs`) internally, mapping from the shared proto message.

## Why RuleSetItem Uses a Flat Structure

Rule sets contain items, where each item has an `action` enum and a set of fields where applicability depends on the action:

```
add_http_request_header → header, value
redirect → redirectUri, responseCode, conditions
ip_based_max_connections → defaultMaxConnections, ipMaxConnections
... (11 action types total)
```

The alternatives considered:

1. **Separate message types per action.** Would create 11 additional message types (AddHttpRequestHeaderRule, RedirectRule, etc.) and require a proto `oneof` with 11 arms. This is technically cleaner but makes the YAML authoring experience verbose — users would need to know the exact message type name for each action.

2. **Chosen: Flat structure with action enum.** Matches the OCI API model directly. Users set `action: redirect` and then fill in `redirectUri`, `responseCode`, and `conditions`. Irrelevant fields are ignored. This is the same pattern used by the OCI Terraform and Pulumi providers.

The flat structure makes the action-to-field mapping important documentation (provided as a table in the catalog page). The trade-off is that the proto does not enforce which fields are valid for which action at the schema level — the IaC modules pass through what the user provides, and the OCI API rejects invalid combinations at deployment time.

## Why idle_timeout_in_seconds Is int64 in Proto

The `ConnectionConfiguration.idle_timeout_in_seconds` field is `int64` in the proto, but the OCI Terraform provider accepts it as a `string`. This is a known quirk of the Terraform provider where numeric values are encoded as strings.

The proto uses `int64` because:
1. It is the semantically correct type — the value represents a duration in seconds.
2. Users write `idleTimeoutInSeconds: 300` in YAML, not `idleTimeoutInSeconds: "300"`.
3. The Pulumi module converts to string internally: `pulumi.String(fmt.Sprintf("%d", cc.IdleTimeoutInSeconds))`.

This keeps the user-facing API clean while handling the provider quirk in the implementation.

## What's Excluded and Why

### path_route_set

Oracle has deprecated `path_route_set` in favor of routing policies. The `oci_load_balancer_path_route_set` Terraform resource still exists for backward compatibility, but new deployments should use routing policies instead. The OciLoadBalancer component does not create path route sets.

### Routing Policies

Routing policies (`oci_load_balancer_routing_policy`) provide content-based routing with conditions on URL path, headers, and other request attributes. They are not bundled into this component because:

1. **Complexity scope.** The spec already has 17 nested messages. Adding routing policies with their own condition types and action types would significantly increase the spec surface area.
2. **Independent lifecycle.** Routing policies can be shared across listeners and may be managed by different teams than the load balancer itself.
3. **Reference support.** Listeners can reference externally managed routing policies by name via `routingPolicyName`.

If routing policy management becomes a common request, a future version could either bundle them or create a separate OciLoadBalancerRoutingPolicy component.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Defined Tags** — OCI defined tags (namespace-scoped, schema-validated) require a tag namespace to be created first. Freeform tags (from OpenMCF labels) cover the majority of tagging use cases. Defined tag support can be added when the tag namespace pattern is established across OCI components.

- **WAF Integration** — OCI Web Application Firewall can be attached to an Application Load Balancer for advanced threat protection. WAF policies are a separate OCI resource with significant configuration surface. Integration would be via a WAF policy OCID field on the load balancer spec.

- **SSL Cipher Suite Management** — OCI supports creating custom cipher suites (`oci_load_balancer_ssl_cipher_suite`). The current implementation references cipher suites by name (e.g., `"oci-default-ssl-cipher-suite-v1"`). Custom cipher suite creation is not bundled — the OCI-provided defaults cover most use cases.

- **Load Balancer Shape Update** — Changing the shape of an existing load balancer (e.g., from a fixed shape to flexible) is a separate OCI API operation with constraints. The current implementation handles this through standard Pulumi/Terraform update semantics.

## Research Notes

### Load Balancer Limits

| Resource | Limit | Notes |
|----------|-------|-------|
| Load balancers per compartment | 50 (default) | Can be increased via service limit request. |
| Backend sets per load balancer | 128 | |
| Backends per backend set | 512 | |
| Listeners per load balancer | 128 | |
| Certificates per load balancer | 128 | Includes both uploaded and OCI Certificate Service references. |
| Hostnames per load balancer | 16 | |
| Rule sets per load balancer | 128 | |
| Rules per rule set | 50 | |

### Flexible Shape Bandwidth

| Attribute | Minimum | Maximum | Notes |
|-----------|---------|---------|-------|
| `minimumBandwidthInMbps` | 10 | 8000 | Billed bandwidth floor. |
| `maximumBandwidthInMbps` | 10 | 8000 | Must be >= minimum. Load balancer scales up to this during traffic spikes. |

When both minimum and maximum are set to the same value, the load balancer operates at a fixed bandwidth (similar to deprecated fixed shapes, but using the flexible shape infrastructure).

### Health Check Behavior

- Health checks run from the load balancer's subnet to the backend's IP and port.
- HTTP health checks expect a response within `timeoutInMillis`. If the response matches `returnCode` (or any 2xx when `returnCode` is omitted) and optionally matches `responseBodyRegex`, the backend is considered healthy.
- TCP health checks establish a TCP connection to the backend's IP and port. If the connection succeeds within `timeoutInMillis`, the backend is healthy.
- After `retries` consecutive failures, the backend is marked unhealthy and removed from the rotation. After the same number of consecutive successes, it is re-added.
- Health check probes do not count toward backend connection limits.

### Listener Protocol and SSL Interaction

The OCI load balancer does not have a separate "HTTPS" protocol. Instead, SSL termination is achieved by combining the `http` protocol with an `sslConfiguration`:

| Desired Behavior | `protocol` | `sslConfiguration` |
|-----------------|-----------|-------------------|
| HTTP (plaintext) | `http` | Not set |
| HTTPS (SSL termination) | `http` | Set with certificate |
| HTTP/2 over TLS | `http2` | Set with certificate |
| TCP passthrough | `tcp` | Not set |
| TLS passthrough | `tcp` | Set with certificate |
| gRPC (always over TLS) | `grpc` | Set with certificate |

### Backend SSL vs Listener SSL

| Aspect | Listener SSL (client-facing) | Backend Set SSL (backend-facing) |
|--------|----------------------------|--------------------------------|
| **Purpose** | Terminate SSL from clients | Re-encrypt traffic to backends |
| **Certificate source** | Load balancer certificate or OCI Certificate Service | Backend server's certificate |
| **Certificate validation** | Not applicable (load balancer presents its own cert to clients) | `verifyPeerCertificate` validates the backend's cert |
| **Session resumption** | `hasSessionResumption` supported | Not applicable |
| **When to use** | Almost always for HTTPS | When compliance requires end-to-end encryption |

### Rule Set Execution Order

When a listener has multiple rule sets, they are applied in the order specified in `ruleSetNames`. Within a rule set, items are evaluated in order. The first matching `redirect` or `allow` rule terminates evaluation — subsequent rules are not processed. Header manipulation rules (`add_http_*`, `remove_http_*`, `extend_http_*`) are all applied regardless of order.
