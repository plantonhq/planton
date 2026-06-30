# ProviderE2EProfile

A KRM-style API that declares how E2E tests are executed for an entire cloud
provider. Every provider that participates in E2E testing has exactly one of
these, located at `{provider}/aa_e2e/profile.yaml`.

## Purpose

The CI workflow reads this profile to determine:

- What credentials the runner needs (OIDC, API token, none for kind)
- What test substrate to use (kind cluster, real cloud, local container)
- What tools to install (kind, kubectl, pulumi, tofu, aws-cli, gcloud, ...)
- How many parallel test cells to run
- Which GitHub Actions environment holds the secrets
- What default cost class and schedule lane apply to the provider's components

This keeps CI workflows thin and provider-agnostic. Adding a new provider to E2E
means creating one `aa_e2e/profile.yaml` -- no workflow YAML changes.

## Where It Lives

```
apis/dev/planton/provider/{provider}/aa_e2e/
  profile.yaml       <-- this API
  harness.go          <-- provider test harness (Go code)
  verify/             <-- resource verification logic
```

The `aa_e2e/` directory is the home for all E2E infrastructure belonging to a
provider. The profile sits alongside the harness code that implements the
testing logic.

## KRM Shape

```yaml
apiVersion: qa.planton.dev/v1
kind: ProviderE2EProfile
metadata:
  name: <provider-name>
spec:
  credential_approach: <none|oidc|api_token|service_account_key|self_hosted>
  test_substrate: <kind|real_cloud|local_container>
  default_cost_class: <free_local|free_cloud|cheap_ephemeral|paid_ephemeral|expensive>
  default_schedule_lane: <pr|nightly|weekly|monthly>
  required_tools: [...]
  github_environment: <string>
  max_concurrent_tests: <int>
```

Field semantics are documented in the proto comments in `spec.proto`. The
profile YAML is loaded by `pkg/e2e/profile/` and consumed by the `planton e2e
discover` CLI command.

## Relationship to ComponentE2EProfile

The provider profile sets defaults. Individual components can override
`cost_class` in their own profile when they are more expensive than typical
components in the same provider (e.g., an EKS cluster component in the AWS
provider).
