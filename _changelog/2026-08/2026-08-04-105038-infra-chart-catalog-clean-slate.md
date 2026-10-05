# Infra-Chart Catalog Clean Slate

**Date**: August 4, 2026
**Type**: Breaking Change
**Provider**: Multi-Provider
**Chart(s)**: entire `charts/` tree

## Summary

The entire infra-chart catalog — 64 charts across 10 providers — was removed
in one deliberate stroke to make room for a demand-curated catalog designed
from first principles. The replacement catalog admits a chart only when its
name and pitch alone invoke "I want this" in a practitioner: complete,
real-world platform architectures composed exclusively from kinds whose
schemas and modules meet the catalog's full depth bar. The `charts/` README
now carries the catalog doctrine the new collection is built under.

## Problem Statement / Motivation

The catalog had accumulated charts of very different vintages: some composed
kind schemas that no longer exist (the legacy Civo, DigitalOcean, and
Scaleway environment charts templated addon specs from retired API shapes),
some targeted providers whose kind catalogs predate the current depth
bar, and the collection as a whole was assembled provider-by-provider rather
than designed as one catalog with one desirability standard. A user browsing
the catalog judged the platform by its weakest entry.

### Pain Points

- Legacy environment charts carried invalid specs for rebuilt Kubernetes
  kinds — latent failures waiting at chart validation.
- Coverage-driven entries (several small-cloud charts) diluted the catalog's
  signal: not every chart answered a real architecture someone would seek out.
- The chart forge rule described a Kubernetes cluster-binding contract
  (a spec-level `target_cluster` selector) that no current Kubernetes kind
  carries, misleading future chart authors at the most load-bearing seam.

## Solution / What's New

1. **Clean slate**: every chart directory under `charts/` was deleted
   (`alicloud`, `aws`, `azure`, `civo`, `digital-ocean`, `gcp`,
   `hetznercloud`, `oci`, `openstack`, `scaleway`). The `charts/` README and
   Makefile remain — both are discovery-driven and correct on an empty tree.
2. **Catalog doctrine in `charts/README.md`**: what earns a slot
   (the desirability bar), where a chart lives (a chart lives under the
   provider that hosts its centerpiece; `charts/kubernetes/` holds charts
   that deploy onto an existing cluster), the platforms-not-applications
   boundary, and the Kubernetes connection-binding contract.
3. **Forge-rule correction**
   (`_rules/charts/forge-planton-infra-chart.mdc`): the stale
   selector-bound-kinds paragraph was replaced with the real contract —
   Kubernetes kinds bind to their cluster through provider connections; in
   charts this is the `planton.dev/connection-name` / `planton.dev/connection`
   annotation pair driven from one values param, paired with
   `metadata.relationships` edges of type `runs_on` for structural ordering.
4. **CI reconciliation** (`.github/workflows/lint.charts.yaml`): the offline
   validation step now sweeps the whole `charts/` tree instead of one
   provider directory, and passes with an honest log line while the tree
   holds no charts (`chart validate --all` deliberately errors on zero
   discovered charts — correct for users, wrong for CI on an intentionally
   empty tree).

## Impact

- **Platform catalogs**: control planes seed charts from the versioned
  release bundle; the chart seeder never gates instance readiness, so a
  release cut before new charts land would only show an empty Charts tab.
  Releases are deliberate, so in practice consumers see the old catalog
  until the new one ships.
- **Chart authors**: the forge rule and the `charts/` README now describe
  the exact composition contracts new charts are built against.

## Related Work

The new catalog's first entries are authored chart-by-chart under the forge
rule's full quality bar — densely commented templates, kind-docs-grade
READMEs, typed params, verified icons, and the offline validation gate.

---

**Status**: ✅ Production Ready
