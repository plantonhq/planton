# KubernetesPeerAuthentication deployment component

**Date**: June 2, 2026
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, IaC Modules (Pulumi + Terraform), E2E

## Summary

Forged `KubernetesPeerAuthentication` (kind 863), the first typed Istio API
deployment component, at 100% upstream fidelity to Istio 1.26.8. It provisions a
namespaced `security.istio.io/v1` `PeerAuthentication` -- the mesh policy that
sets mutual TLS (mTLS) requirements for incoming connections to selected
workloads. This is the first of seven typed Istio API components and the first to
be proven green on a live `kind` cluster across both IaC engines.

## What's new

- **API**: `kubernetespeerauthentication/v1` -- `spec.proto`, `api.proto`,
  `stack_input.proto`, `stack_outputs.proto`, plus generated stubs. The spec
  flattens the upstream fields after the namespaced envelope:
  - `namespace` -- `StringValueOrRef` foreign key to `KubernetesNamespace`.
  - `selector` -- reuses the shared `KubernetesIstioApiWorkloadSelector`.
  - `mtls.mode` -- closed string set `UNSET | DISABLE | PERMISSIVE | STRICT`
    (DD-008, UPPERCASE external-standard exception). Modeled as required-in-set
    (no empty value): omitting the whole `mtls` block is how inheritance is
    expressed; `UNSET` is the explicit inherit value.
  - `port_level_mtls` -- `map<uint32, MutualTls>` per-workload-port overrides,
    keyed by port number, with a message-level rule requiring a selector.
- **Shared type**: grew `KubernetesIstioApiWorkloadSelector.match_labels` (in
  `istio_api.proto`) with the upstream CEL rules its first consumer needs --
  non-empty keys, no wildcards in keys or values.
- **IaC**: typed Pulumi module (`istiosecurityv1.NewPeerAuthentication`) and a
  null-pruned Terraform module (`kubernetes_manifest`), feature-equal across both
  engines.
- **E2E**: tier-4 profile (green, pulumi + terraform). The component declares
  `prerequisites: [KubernetesIstioBaseCrds]`, so the harness installs the Istio
  CRDs before each scenario. Verification dispatches through a new `istioApiKinds`
  map in the Kubernetes verifier (mirror of `gatewayApiKinds`), asserting the
  `peerauthentications.security.istio.io` CR exists after apply and is gone after
  destroy.

## Validation

Full ladder, all green:

- Static: `make protos` (incl. Java gate), `make generate-cloud-resource-kind-map`,
  `go build ./...`, `go vet`, `terraform fmt`/`validate`.
- Codified: 23 protovalidate spec tests (accept + reject for every CEL rule).
- **Live E2E on a `kind` cluster** for both Pulumi and Terraform -- all eight
  lifecycle phases pass (dependency install of `KubernetesIstioBaseCrds` -> apply
  -> verify exists -> destroy -> verify gone). Live runs caught three issues that
  static checks could not: the crd2pulumi typed `Spec`/`Mtls`/`Selector` fields
  must be assigned the Args **value** (not the `*SpecPtr()` wrapper, which
  marshals to the wrong element type), and a Terraform null-attribute access in a
  selector conditional (`&&` does not short-circuit; switched to a `?:` guard).

## Notes

- Only `STRICT` (and the no-`mtls` inherit path) was exercised live; `UNSET` is
  carried per the upstream proto (source of truth) and is accepted by the
  protovalidate layer. The CRD's OpenAPI enum, generated from the same
  `release-1.26`, includes `UNSET`.
- Also closed a pre-existing T01 gap discovered during wiring:
  `KubernetesIstioBaseCrds` (868) had a green E2E profile but no `discover.go`
  PascalCase entry and no test functions, so its CI matrix regex matched nothing.
  Added its `knownPrefixes` entry, `TestKubernetesIstioBaseCrds_{Pulumi,Terraform}`
  functions, and tier-4 wiring; it is now proven live as the PeerAuthentication
  prerequisite.
