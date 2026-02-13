# Getting Started Page Fresh Rewrite

**Date**: February 13, 2026
**Type**: Enhancement
**Components**: Documentation

## Summary

Rewrote the Getting Started page from scratch to match the quality bar established by the Phase 1-5 documentation rewrites. The previous version was only patched during earlier phases — it had a missing `openmcf init` step, duplicate sections, and structural issues. The new page delivers a focused, source-verified 10-minute quickstart experience.

## Problem Statement / Motivation

The Getting Started page is the most critical page in any open-source project's documentation. It is the first hands-on experience a developer has with the framework, and the page that determines whether they continue or abandon adoption.

### Pain Points

- **Missing `openmcf init` step** — the page showed `openmcf apply` without first initializing the Pulumi stack, which would fail in practice
- **Confusing dual command presentation** — showed both `openmcf apply` and `openmcf pulumi up` without explaining the difference, leaving new users unsure which to use
- **"Common Commands" section** duplicated the CLI reference page, violating the deduplication architecture
- **"Troubleshooting" section** duplicated the dedicated troubleshooting page
- **Thin prerequisites** — no version guidance, no tool-purpose mapping
- **Filler content** — "This guide will help you install..." preamble, emoji in explanations
- **No cleanup step** — the page ended with PostgreSQL still running, no `openmcf destroy`
- **No explanation of key concepts** — provisioner labels, stack labels, and the KRM manifest model were used but not explained

## Solution / What's New

Complete fresh-start rewrite with a focused structure that covers the full deployment lifecycle.

### Page Structure

```
Getting Started
├── Intro (clear outcome: deploy PostgreSQL, verify, tear down in 10 minutes)
├── What You'll Need (prerequisites with purpose and install commands)
├── Install OpenMCF (Homebrew + verify)
├── Create a Local Cluster (Kind)
├── Write Your Manifest (KubernetesPostgres with inline KRM explanations)
├── Validate the Manifest (openmcf validate)
├── Prepare for Deployment (pulumi login --local + openmcf init)
├── Deploy (openmcf apply)
├── Verify (kubectl get pods/svc)
├── Clean Up (openmcf destroy + kind delete cluster)
├── What Just Happened (5-step pipeline explanation)
└── Next Steps (progressive learning path)
```

### Key Design Decisions

- **Unified `apply` command only** — new users see one way to deploy, not two. Direct engine commands are covered in CLI docs.
- **Stack labels embedded in manifest** — matches the official example in `examples/kubernetes-postgres.yaml` and teaches the idiomatic label-based configuration pattern.
- **`openmcf init` step included** — the correct flow verified against `pkg/iac/pulumi/pulumistack/init.go` and `run.go`.
- **No duplicate sections** — removed "Common Commands" (link to CLI reference) and "Troubleshooting" (link to troubleshooting page).
- **Full lifecycle** — validate, init, apply, verify, destroy. Reader cleans up after themselves.

## Implementation Details

### Manifest Verification

Every field in the example manifest was verified against the KubernetesPostgres protobuf definitions:

- `apiVersion: kubernetes.openmcf.org/v1` — matches `api.proto` constant validation
- `kind: KubernetesPostgres` — matches `api.proto` constant validation
- `spec.namespace` with `value` wrapper — matches `StringValueOrRef` type
- `spec.createNamespace` — valid boolean field in `spec.proto`
- `spec.container.replicas`, `resources`, `diskSize` — all match `KubernetesPostgresContainer` message
- camelCase field names — matches JSON serialization convention used across all OpenMCF manifests

### CLI Command Verification

All commands verified against source:

- `openmcf validate -f` — alias for `validate-manifest`, verified in `cmd/openmcf/root/`
- `openmcf init -f` — verified in `cmd/openmcf/root/init.go`, calls `pulumi stack init`
- `openmcf apply -f` — verified in `cmd/openmcf/root/apply.go`, routes via provisioner label
- `openmcf destroy -f` — verified in command tree
- `openmcf version` — verified in command tree

### Link Verification

All 13 internal documentation links verified as pointing to existing files:
- 4 concept pages (dual-iac-engines, cloud-resource-kinds, manifests, state-management)
- 3 tutorial pages (first-aws-resource, multi-provider, first-kubernetes-resource)
- 3 guide pages (aws/gcp/azure-provider-setup)
- Concepts index, catalog, troubleshooting

## Benefits

- **Correct deployment flow** — includes the `openmcf init` step that was previously missing
- **Focused structure** — no duplicate sections, every paragraph earns its place
- **Full lifecycle coverage** — validate through destroy, reader cleans up after themselves
- **Concept introduction** — briefly explains KRM, provisioner labels, and stack labels without overloading
- **Progressive learning path** — clear Next Steps linking to concepts, tutorials, guides, and catalog
- **Deduplication** — links to CLI reference and troubleshooting instead of duplicating content

## Impact

- **New users** get a correct, working quickstart that succeeds on the first try
- **Evaluators** see a professional, focused page that demonstrates the framework's KRM consistency
- **Tutorials** can reference Getting Started for prerequisites without concern about incorrect steps
- **Documentation consistency** — Getting Started now meets the same fresh-start quality bar as all other sections

## Related Work

- Phase 0: Existing docs audit (patched this page — identified issues)
- Phase 1: Concepts section rewrite (pages referenced from Getting Started)
- Phase 2: CLI docs expansion (deduplication architecture established)
- Phase 4: Tutorials section (first-kubernetes-resource builds on Getting Started)
- Phase 6: Final audit (flagged Getting Started as needing fresh rewrite)

---

**Status**: Production Ready
**Timeline**: Single session
