---
title: "Contributing"
description: "How to set up a development environment, build from source, run tests, and contribute to OpenMCF"
icon: "users"
order: 60
---

# Contributing to OpenMCF

OpenMCF is an open-source project and welcomes contributions from the community. This guide covers everything you need to set up a development environment, build the project, run tests, and submit changes.

## Prerequisites

Before you begin, install the following tools:

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.25+ | Core language for CLI, IaC modules, and APIs |
| **Bazel** | Latest (via `bazelw` wrapper) | Build system for proto generation and Gazelle |
| **Buf CLI** | Latest | Protocol Buffer linting, generation, and formatting |
| **Make** | System default | Build orchestration |
| **Python** | 3.x | Versioning scripts and helper tooling |
| **Terraform** | Latest | Validating Terraform/OpenTofu modules |

Optional but recommended:

| Tool | Purpose |
|------|---------|
| **Pulumi CLI** | Testing Pulumi modules locally |
| **OpenTofu CLI** | Testing OpenTofu modules locally |
| **Docker** | Building container images |

## Getting the Source

Fork the repository on GitHub, then clone your fork:

```bash
git clone https://github.com/<your-username>/openmcf.git
cd openmcf
```

Verify the build works:

```bash
make build
```

This runs the full build pipeline: proto generation, Bazel Gazelle, CLI compilation, and validation.

## Building

### Full Build

```bash
make build
```

Runs the complete build pipeline including proto generation, kind map generation, Bazel Gazelle updates, and CLI compilation.

### Proto Generation

```bash
make protos
```

Generates Go stubs from Protocol Buffer definitions using Buf. Run this after modifying any `.proto` file. This target:

1. Runs `buf generate` to produce Go, TypeScript, and Java stubs
2. Copies generated Go stubs into the `apis/` source tree
3. Runs `bazel run //:gazelle` to update Bazel BUILD files

### CLI Only

```bash
make build-cli
```

Cross-compiles the CLI binary for darwin/linux on amd64/arm64.

### Kind Map Generation

```bash
make generate-cloud-resource-kind-map
```

Regenerates the cloud resource kind map. Run this after adding a new deployment component to `cloud_resource_kind.proto`.

## Testing

### All Tests

```bash
make test
```

Runs all Go tests with race detection enabled:

```bash
go test -race -v -count=1 -p 4 ./...
```

### Component-Scoped Tests

For faster iteration when working on a single component, run tests directly in the component's directory:

```bash
# Run proto validation tests for a specific component
go test -v ./apis/org/openmcf/provider/aws/awss3bucket/v1/...

# Build check for a Pulumi module
go build ./apis/org/openmcf/provider/aws/awss3bucket/v1/iac/pulumi/...

# Vet check
go vet ./apis/org/openmcf/provider/aws/awss3bucket/v1/iac/pulumi/...

# Validate a Terraform module
cd apis/org/openmcf/provider/aws/awss3bucket/v1/iac/tf
terraform init && terraform validate
```

Running localized commands avoids rebuilding the entire project for single-component changes.

### Linting

```bash
# Go formatting
make fmt

# Go vet
make vet

# Proto linting
make buf-lint
```

Proto linting includes a custom Buf plugin (`buf/lint/optional-linter/`) that validates scalar fields with default annotations are marked as `optional`.

## Code Style

### Go

- Run `make fmt` before committing
- Run `make vet` to catch common issues
- Follow standard Go conventions

### Protocol Buffers

- Run `buf format` and `buf lint` before committing proto changes
- Do not use Java reserved words (`static`, `class`, `default`, `switch`, etc.) as enum values or field names — Java stub generation will fail
- Scalar fields with defaults must be marked `optional` and annotated with `(org.openmcf.shared.options.default)`

### Naming Conventions

- Component folders: `<provider><resource>` in lowercase (e.g., `awss3bucket`, `gcpgkecluster`)
- Component kinds: `<Provider><Resource>` in PascalCase (e.g., `AwsS3Bucket`, `GcpGkeCluster`)
- API versions: `<provider>.openmcf.org/v1`

## Submitting Changes

### Branch and Commit

1. Create a feature branch from `main`:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following the coding standards above

3. Run tests before committing:

```bash
make test
```

4. Commit with a clear, descriptive message:

```bash
git commit -m "feat(aws): add lifecycle rule support to AwsS3Bucket"
```

### Pull Request

1. Push your branch to your fork:

```bash
git push origin feature/your-feature-name
```

2. Open a Pull Request against the `main` branch

3. Include in your PR description:
   - What the change does and why
   - How to test it
   - Any breaking changes

4. Address review feedback and iterate

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>
```

| Type | When to use |
|------|-------------|
| `feat` | New feature or component |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or updating tests |
| `chore` | Build, tooling, or dependency changes |

Scope should reflect the affected area: `aws`, `gcp`, `kubernetes`, `cli`, `docs`, `build`.

## Communication

- **GitHub Issues** — Bug reports, feature requests
- **GitHub Discussions** — Questions, ideas, design proposals
- **Discord** — Real-time discussion with the community

When proposing larger changes, open a GitHub Issue first to discuss the approach before investing significant effort.

## What's Next

- **[Adding Components](./adding-components)** — Step-by-step guide for creating new deployment components
- **[Deployment Components](/docs/concepts/deployment-components)** — Understand the anatomy of a component
- **[Dual IaC Engines](/docs/concepts/dual-iac-engines)** — How Pulumi and OpenTofu modules work together
