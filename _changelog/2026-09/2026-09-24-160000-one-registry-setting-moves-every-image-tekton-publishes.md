# One Registry Setting Moves Every Image Tekton Publishes

**Date**: September 24, 2026
**Type**: Feature
**Components**: Catalog (KubernetesTektonOperator, KubernetesTekton), Guards (cross-engine parity)

## Summary

`KubernetesTektonOperator` gains `image_registry`, the catalog's air-gap/mirror field under the name fifteen other kinds already use. Set it to a mirror or pull-through cache of ghcr.io and every image Tekton publishes pulls from there at its path, tag and digest: the operator and its webhook, and every component the operator installs (Pipelines, Triggers, Dashboard, Chains, Results, the pruner), including the entrypoint, nop, workingdirinit and sidecarlogresults images every build pod starts with. Before, only the operator's own two images could be pointed anywhere, and only by pinning one exact image that the next catalog upgrade would leave behind.

## What Changed

- `catalog/kubernetes/kubernetestektonoperator/v1alpha1/spec.proto`: `image_registry = 8`, with the platform kind's format rule (no scheme, no trailing slash). `KubernetesTekton`'s spec says where its images come from. Stubs, both reference pages and the proto-docs index regenerated.
- Both modules hand the operator's lifecycle container one `IMAGE_<COMPONENT>_<key>` variable per image, moved to the registry, and move the manifest's own images and variables. The table of images is one text in both engines (`iac/pulumi/module/images.go`, `iac/tf/images.tf`), read from the component manifests bundled in the pinned operator image, never from the operator repository's `components.yaml` (at `v0.80.0` that file names Results `v0.19.0` while the image installs `v0.18.0`).
- The operator's own `TEKTON_REGISTRY_OVERRIDE` is not used: it rewrites every host, and would send the images Tekton does not publish (the `cgr.dev` and `mcr.microsoft.com` shell images, Results' Docker Hub Postgres) to a mirror that does not carry them. Those keep their registry.
- The table names the operator release it was read from. A unit test fails when `OperatorRelease` moves without it, and both engines refuse at plan time with the sentence that says how to rebuild it, so a new operator can never be handed last release's component images.
- The Terraform module's floor moves to 1.3 (`startswith`, lifecycle preconditions).
- `hack/guards/ensure_cross_engine_script_parity.sh`: the header says what it now also carries, a data table both engines read; the check is unchanged.
- The kind's catalog page, guide, README and preset note lead with `image_registry` over the image overrides, and the guide names the `kubectl` read that proves the mirror is in use.
- `_issues/`: the Tekton registry-root request is closed; its third ask, outputs naming the images an install pulls, is its own issue.

## Verification

- `go test` of the kind's spec suite and module: pass. The module tests pin an empty setting rendering the manifest untouched, every Tekton image moving at its digest, images on other hosts staying, an explicit image winning, and the table belonging to the pinned release; changing the table's release to `v0.79.0` turned three of them red with the refusal sentence.
- Terraform, on a scratch copy against the real `v0.80.0` release manifest: an empty setting renders images and env identical to the manifest; set, the lifecycle container carries 21 image variables all on the mirror (the manifest's 2 plus the table's 19) with every other variable unchanged; `plan` refuses a mismatched table with the same sentence and plans all 41 documents with the matching one; `terraform validate` and `terraform fmt -check` clean.
- The parity guard: 6 pairs byte-identical, and red on a planted one-byte drift in the Terraform copy.
- `ensure_image_paths`, `ensure_modules_are_self_contained`, `ensure_license_footers`, `buf lint` and `buf format` on the two protos: clean.
- The table's 19 entries match what a live `v0.80.0` install runs (read from a GKE cluster's Deployments and the controller's image args) for every component that cluster runs, and the bundled manifests for Triggers, Dashboard and the pruner, which it does not.
