# Standard Istio CRD Installation

This preset installs the Istio custom resource definitions on a cluster. The spec is deliberately empty — the component has exactly one job, and it needs no configuration.

## When to Use

- Before deploying any typed Istio kind (DestinationRule, ServiceEntry, PeerAuthentication, RequestAuthentication, AuthorizationPolicy, Telemetry, EnvoyFilter)
- Clusters that consume Istio APIs (for example, ambient mode or CRD-driven tooling) without running the full mesh yet
- As the CRD prerequisite that decouples API installation from mesh installation

## Key Configuration Choices

- **Empty spec** (`spec: {}`) — the component installs the Istio CRD set as-is; versioning follows the component's pinned Istio release, so there is nothing to tune

## Placeholders to Replace

- `metadata.name` — your installation's name
