# KubernetesBackendTlsPolicy Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API
`BackendTLSPolicy` on a target cluster using the typed crd2pulumi SDK
(`gatewayv1.NewBackendTLSPolicy`). BackendTLSPolicy is a standard-channel
resource served as `gateway.networking.k8s.io/v1` (the `v1alpha3` version
is deprecated upstream and no longer served). The typed approach catches
field-name and structure errors at compile time.

## Prerequisites

- The Gateway API standard-channel CRDs must already be installed on the
  cluster (`KubernetesGatewayApiCrds`).
- The backend Service(s) the policy targets, in the same namespace (the
  policy cannot target across namespaces).
- For the bring-your-own-CA arm: a same-namespace ConfigMap carrying the
  PEM CA bundle in a key named `ca.crt`.
- Go toolchain and the Pulumi CLI.
- Access to the target Kubernetes cluster.

## Rendering Notes

- **`group` is always emitted, even when empty**: the CRD requires the
  `group` KEY on targetRefs and caCertificateRefs but allows the empty
  value (core-group referents -- Service, ConfigMap). The spec models it
  as a presence-required `optional`, and the module always sets `Group:`
  from the resolved value instead of dropping the empty string.
- **The trust-anchor arms are exactly-one-of** (protovalidate CEL mirrors
  the CRD's own rules): `caCertificateRefs` and `wellKnownCACertificates`
  are each set only when present, so exactly one appears in the rendered
  CR. The typed SDK field carries the CRD's exact key casing
  (`WellKnownCACertificates` -- capital CA), which the spec pins via
  `json_name`.
- **SAN pairing is pre-enforced**: `type` is a closed enum
  (`Hostname` | `URI`) and `hostname`/`uri` are each set only when
  present -- protovalidate guarantees exactly the matching value field
  appears.
- **No await/wait logic is attached**: the per-ancestor
  Accepted/ResolvedRefs conditions belong to the Gateway controller's
  reconciliation, not to applying the resource.

## Local Development

```bash
make deps
make build
```

## Usage

### With the Planton CLI

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesBackendTlsPolicyIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to a manifest) or
`IAC_INPUT_YAML` (inline YAML content):

```bash
export IAC_INPUT_YAML_FILE=../../e2e/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `policy_name` | Name of the created BackendTLSPolicy (equals `metadata.name`) |
| `namespace` | Namespace the BackendTLSPolicy was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewBackendTLSPolicy, v1); options mapped inline
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Output constant names
    ├── target_refs.go   # targetRefs (same-namespace Services, optional sectionName) mapping
    └── validation.go    # Trust anchor, hostname, and SAN mapping
```

The policy's `StringValueOrRef` foreign keys (`namespace`,
`targetRefs[].name`, `caCertificateRefs[].name`) arrive resolved to literal
strings in the IaC input; the module reads their final values directly.

## References

- [Gateway API BackendTLSPolicy](https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)
