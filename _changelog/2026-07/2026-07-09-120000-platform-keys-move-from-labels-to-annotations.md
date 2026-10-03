# Platform Keys Move from Labels to Annotations

**Date**: July 9, 2026
**Type**: Breaking Change
**Components**: CLI (provisioner detection, backend config, kube context), Kubernetes Pulumi Modules, Catalog Documentation

## Summary

Every platform-behavior signal the CLI reads from a manifest now lives in
`metadata.annotations` instead of `metadata.labels`, with no label fallback.
This includes the provisioner selector (`planton.dev/provisioner`), Pulumi
stack location (`pulumi.planton.dev/*` including `stack.fqdn`),
Terraform/OpenTofu backend configuration (`terraform.planton.dev/backend.*` /
`tofu.planton.dev/backend.*`), and the Kubernetes CLI conveniences
(`kubernetes.planton.dev/context`,
`kubernetes.planton.dev/docker-config-json-file`).

`metadata.labels` is now exclusively user territory: organizational metadata
that IaC modules derive into cloud-provider tags. A platform key placed under
`labels` is ignored by every reader.

## Problem Statement

Planton IaC modules derive cloud-provider tags from `metadata.labels`. Platform
keys riding in labels therefore became tags on the user's real infra components
— polluting their cloud with internal orchestration detail and breaking
providers with strict tag charsets (free-text values have failed AWS tag
validation in production). The two metadata maps had overlapping, inconsistent
roles; this release gives each a single job.

## Changes

### Readers (annotations-only, label fallback removed)

- `pkg/reflection/metadatareflect`: `ExtractAnnotations()` is the shared
  extraction helper; `ExtractLabels()` remains for tag derivation only.
- `pkg/iac/provisioner`: provisioner detection reads
  `planton.dev/provisioner` from annotations.
- `pkg/iac/tofu/backendconfig` + `pkg/iac/pulumi/backendconfig`: backend
  location read from annotations; the legacy `backend.object` composite key is
  gone (use `backend.bucket` + `backend.key`).
- `pkg/kubernetes/kubecontext`: kube context read from annotations.
- `kubernetesdeployment` / `kubernetesstatefulset` Pulumi modules: the
  docker-config-json-file path is read from
  `metadata.annotations["kubernetes.planton.dev/docker-config-json-file"]`.

### Key constant packages renamed

- `provisionerlabels` → `provisionerannotationkeys`
- `tofulabels` → `tofuannotationkeys`
- `pulumilabels` → `pulumiannotationkeys`
- `kuberneteslabels` → `kubernetesannotationkeys`

Each package documents the labels-never-touch-the-cloud invariant.

### User-facing strings and docs

- CLI messages and errors now say "annotation" (e.g. "Detected Stack from
  Annotations", init command help, incomplete-backend-config error).
- Catalog pages (`apis/**/catalog-page.md`), site docs mirror, tutorial and
  concept pages (including `state-management.md`), `iac/hack/manifest.yaml`
  files, presets, `examples/*.yaml`, `hack/examples/manifest-backend-config/*`
  (also fixed a stale `aws.planton.ai/v1` apiVersion and the legacy
  `backend.object` form), and the openstack chart templates all show the
  annotation form.

## Migration

Move platform keys from `labels:` to `annotations:` in manifests:

```yaml
# Before
metadata:
  labels:
    planton.dev/provisioner: tofu
    tofu.planton.dev/backend.type: s3

# After
metadata:
  annotations:
    planton.dev/provisioner: tofu
    tofu.planton.dev/backend.type: s3
```

User tags stay under `labels:` unchanged.

## Verification

- `go build` across `pkg/`, `internal/`, `cmd/`
- `bazel build //pkg/iac/... //pkg/kubernetes/... //pkg/reflection/...`
- `bazel test` on `metadatareflect`, tofu/pulumi `backendconfig` (includes a
  regression test proving keys under `labels` are ignored)
- `make gazelle` after package renames
