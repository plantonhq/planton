# CatalogKindE2EProfile

A KRM-style API that declares a single kind's E2E testing readiness,
classification, and runtime configuration. Every kind that participates
in E2E testing has exactly one of these, located at
`{kind}/v1/e2e/profile.yaml`.

## Purpose

The CI workflow reads kind profiles to:

- Determine which kinds are testable (status: green) vs deferred/skipped
- Group kinds into matrix cells by tier and provisioner
- Calculate per-cell timeouts from individual kind timeouts
- Construct `go test -run` regexes that target the right test functions
- Skip deferred kinds with self-documenting reasons in the profile

This makes CI discovery mechanical. The `planton e2e discover` command scans
all profiles under a provider, filters by status, and produces the GitHub
Actions matrix JSON. Adding a new kind to CI means creating one
`v1/e2e/profile.yaml` -- no workflow YAML or Go code changes.

## Where It Lives

```
catalog/{provider}/{kind}/v1/e2e/
  profile.yaml       <-- this API
  scenarios/          <-- test scenario manifests (minimal.yaml, with-probes.yaml, ...)
  fixtures/           <-- prerequisite deployments for operator-dependent kinds
```

The profile sits at the root of the kind's E2E directory, alongside the
scenarios and fixtures that the E2E framework executes.

## KRM Shape

```yaml
apiVersion: qa.planton.dev/v1
kind: CatalogKindE2EProfile
metadata:
  name: <kind-name>
spec:
  tier: <int>
  status: <green|deferred|skip|stub>
  deferred_reason: <string> # when status is deferred or skip
  validated_provisioners: [pulumi, terraform]
  timeout_minutes: <int>
  cost_class: <enum> # optional override of provider default
  limitations: [...] # known constraints (e.g., "requires 4+ GB RAM")
```

Field semantics are documented in the proto comments in `spec.proto`. The
`validated_provisioners` field uses the shared `IacProvisioner` enum from
`shared/iac.proto`.

## Status Lifecycle

A kind progresses through statuses as its E2E support matures:

```
stub -> skip -> deferred -> green
```

- **stub**: The IaC module exists but has no real deployment logic yet.
- **skip**: The module works but E2E is intentionally excluded (needs cloud
  credentials, database dependencies, or design decisions not yet made).
- **deferred**: E2E was attempted but hits a known limitation (kind resource
  constraints, provider quirks). The `deferred_reason` documents why.
- **green**: E2E passes on CI. Included in scheduled runs.

Kinds can also move backward (green -> deferred) when a regression or
platform change causes failures. The profile is the single source of truth.

## Tier Classification

Tiers group kinds by dependency complexity within a provider. For
Kubernetes:

| Tier | Category           | Dependencies         | Example                             |
| ---- | ------------------ | -------------------- | ----------------------------------- |
| 1    | Native K8s         | None                 | Namespace, Deployment, Secret       |
| 2    | Helm-based         | Self-contained chart | Redis, Grafana, ArgoCD              |
| 3    | Operator-dependent | Needs operator CRDs  | Postgres, Kafka, Elasticsearch      |
| 4    | Operators/addons   | Cluster-level infra  | CertManager, Istio, ExternalSecrets |

Other providers will define their own tier semantics as E2E expands.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
