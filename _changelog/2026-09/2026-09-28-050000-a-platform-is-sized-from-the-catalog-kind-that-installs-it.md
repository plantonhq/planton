# A Platform Is Sized from the Catalog Kind That Installs It

**Date**: September 28, 2026
**Type**: Feature
**Components**: `KubernetesPlantonPlatform` (spec, both IaC modules, docs, fixtures), `KubernetesPlantonOperator` (default chart), the operator's exported facts

## Summary

The operator can now size every component of a self-hosted platform (operator chart 0.23.0). The catalog kind that installs a platform now exposes the same sizing: every component takes `resources`. The store also gains `max_memory`, its dataset ceiling, which raising its memory limit depends on. Temporal and OpenFGA gain blocks of their own. Each field states the operator's measured default as `default_container_resources`, and a gate reads the operator's own published registry, so this kind can never again lag the operator on sizing or show a stale default.

## What Changed

- **`spec.proto`:**
  - `resources` (`dev.planton.kubernetes.ContainerResources`) on `control_plane`, `console`, `runner`, `gateway`, `identity`, `database.postgresql`, `database.redis`, `vault` and `components.graph`;
  - `database.redis.max_memory`, with a Valkey-units CEL rule;
  - new `temporal` (the `frontend`, `history`, `matching` and `worker` services) and `openfga`.

  Every field says what it sizes, its default, the per-quantity merge, the restart it causes, and that it needs chart 0.23.0. One message serves Temporal's four services, and their defaults differ, so each service field carries its own default.
- **Both modules** render only the quantities set, in lockstep:
  - Pulumi `resourcesMap`, with render tests in `sizing_test.go`;
  - Tofu `sizing` locals, checked with `tofu console` on a sized and an empty spec, which render byte-for-byte what Pulumi renders.

  The gateway block now renders without `local_port`.
- **The gate** (`v1alpha1/sizing_gate_test.go`) reads `operator/api/v1/component_sizing.json` and the operator's CRD. It fails when:
  - a sized workload has no field here, or none in the CRD;
  - a stated default differs from the registry.

  The two operator files are exported for Bazel by data-only `BUILD.bazel` files; the operator itself stays outside this workspace.
- **`KubernetesPlantonOperator`:** its default chart moves from 0.15.0 to 0.23.0 in the proto, both modules, the README, both presets and `controls.yaml`.
- **Docs:**
  - the README's inaccurate "Sizing" row is split into replicas and a `<component>.resources` row;
  - GUIDE.md gains "Size a component, and only what you mean to change";
  - the catalog page gains the sizing paragraph;
  - `cost.yaml` lists resources as a cost driver;
  - the full-surface fixture sizes four components.

## Verification

- `go test` passes for both kinds, and so do the Bazel targets `//catalog/kubernetes/kubernetesplantonplatform/v1alpha1:v1alpha1_test` and `.../iac/pulumi/module:module_test`.
- The gate is red when the published default is re-measured (5Gi against the stated 4Gi) and when the registry sizes a workload this kind lacks.
- `planton validate` passes on the fixture, and `make protos` passes (the Java CEL gate included).
- Known quirk: after a previous generation, `make protos`'s Java step fails once reading `_test` stubs and passes on the next run.
