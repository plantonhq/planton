# OpenMCF E2E Test Framework

End-to-end tests that deploy real infrastructure using OpenMCF IaC modules and
verify the results against real providers.

## What This Framework Does

Every OpenMCF component ships with Pulumi and Terraform modules that create
cloud infrastructure. These E2E tests prove that those modules actually work by
executing the full lifecycle against real providers:

1. **VALIDATE** -- load the manifest and build the stack input
2. **DEPLOY** -- run the IaC module (Pulumi up or Terraform apply)
3. **VERIFY-OUT** -- check that stack outputs are populated
4. **VERIFY-RES** -- confirm resources exist using provider-native tools
5. **DESTROY** -- tear down all created resources
6. **VERIFY-CLN** -- confirm resources are gone

If any phase fails, the framework still attempts DESTROY to avoid leaking
resources.

## How Test Scenarios Are Organized

Test manifests live **next to their components**, not in a central test
directory. Each component has an `e2e/` folder at the `v1` level:

```
apis/org/openmcf/provider/{provider}/{component}/v1/
  e2e/                    <-- test manifests live here
    minimal.yaml
    with-probes.yaml
    with-hpa.yaml
    ...
  iac/
    hack/manifest.yaml    <-- the canonical example manifest
    pulumi/               <-- Pulumi module
    tf/                   <-- Terraform module
  spec.proto
```

Each YAML file in `e2e/` is a complete OpenMCF manifest representing one test
scenario. The framework discovers them automatically and runs each through the
full 6-phase lifecycle as an independent sub-test.

## Provider Harnesses

Each cloud provider has a harness that manages test infrastructure and
verification. The harness for a provider lives under that provider's directory
in an `aa_e2e/` folder (the `aa_` prefix ensures it sorts first in the file
explorer, ahead of component directories):

```
apis/org/openmcf/provider/{provider}/aa_e2e/
  e2e_harness.go          <-- provider lifecycle (setup/teardown)
  e2e_verify.go           <-- resource verification logic
```

For Kubernetes, the harness creates a `kind` cluster and uses `kubectl` for
verification. Future cloud provider harnesses will manage credentials and use
provider SDKs or CLIs for verification.

## Running Tests

Prerequisites: Docker running, `kind`, `pulumi`, and `kubectl` installed.

```bash
# All Kubernetes E2E tests (all components, all scenarios)
make e2e-test-kubernetes

# Single component, all scenarios
make e2e-test-component component=KubernetesNamespace

# Single scenario via Go test
go test -tags=e2e -timeout=30m -v -count=1 \
  -run "TestKubernetesNamespace_Pulumi/minimal$" ./e2e/...
```

## Build Tag Isolation

All E2E test files use `//go:build e2e`. This means:

- `go test ./...` and `make test` never trigger E2E tests
- `go build ./...` never compiles E2E test binaries
- You must pass `-tags=e2e` explicitly to run them

The framework packages under `e2e/framework/` have **no** build tag -- they are
ordinary Go libraries that get compiled normally. Only the test files that
create real infrastructure are gated.

## How Verification Works

The framework does not hardcode resource names. Instead, it parses each test
manifest at runtime to extract the resource name, namespace, and kind, then
builds the appropriate verification dynamically. This means adding a new test
scenario is as simple as dropping a YAML file into the component's `e2e/`
folder -- no Go code changes needed.

## Adding a New Test Scenario

1. Create a YAML manifest in `{component}/v1/e2e/` with a descriptive filename
2. Use a unique `metadata.name` (and unique namespace if the component creates
   one) to avoid collisions with other scenarios
3. Run `make e2e-test-component component={ComponentName}` to verify it works
4. That's it -- the framework discovers and runs it automatically

## Adding a New Provider

1. Create `apis/org/openmcf/provider/{provider}/aa_e2e/` with harness and
   verify files
2. Implement the `provider.Harness` interface (Setup, Teardown, VerifyDeployed,
   VerifyDestroyed)
3. Add a test entry point in `e2e/` that creates the harness and discovers
   scenarios for that provider
4. Add Makefile targets

## Architecture

```
e2e/
  e2e_test.go             -- TestMain: shared infrastructure lifecycle
  kubernetes_test.go      -- Kubernetes test entry points (per-component)
  framework/
    runner/               -- 6-phase lifecycle engine, Pulumi/Terraform execution
    provider/             -- Harness interface definition
    discovery/            -- Filesystem scanner for components and scenarios
    reporter/             -- JSON + Markdown report generation
```

The framework is engine-agnostic. The runner supports both Pulumi and Terraform
execution paths. Each component test runs through the same lifecycle regardless
of which engine is used.
