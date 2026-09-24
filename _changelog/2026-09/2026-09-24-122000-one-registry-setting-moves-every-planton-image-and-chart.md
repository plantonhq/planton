# One Registry Setting Moves Every Planton Image and Chart

**Date**: September 24, 2026
**Type**: Feature
**Components**: Operator, Helm (planton-operator), Catalog (KubernetesPlantonPlatform, KubernetesPlantonOperator, KubernetesPlantonRunner), Guards

## Summary

Every Planton release is published to ghcr.io and copied, byte for byte, to Google Artifact Registry at `asia-south1-docker.pkg.dev/plantonhq` under the same paths. An install now chooses between them, or names a mirror of its own, with one setting per layer instead of one override per image.

- **The operator:** `spec.imageRegistry` on the platform resource is the root the control plane, console, and runner images are pulled from, as `<imageRegistry>/<image>`. Empty keeps `ghcr.io/plantonhq/planton`. A component's own `image.repository` still wins. The operator's version check now reads the minimum-operator label from the repository the platform actually pulls, so an install on a mirror never reaches for ghcr.io.
- **The catalog:** `KubernetesPlantonPlatform` gains `image_registry`, rendered only when set, and `runner.image`, the runner override the operator already supported. `KubernetesPlantonOperator` and `KubernetesPlantonRunner` gain `chart_repository`, defaulting to `oci://ghcr.io/plantonhq/charts`, in place of the root each module hard-coded.
- **The guard:** `hack/guards/ensure_image_paths.sh` refuses a registry root written into code anywhere but its home: `operator/internal/plantonregistry` for the operator, and each catalog module's own `vars.go` and `locals.tf`, which mirror the proto default.

## What Changed

- `operator/internal/plantonregistry/plantonregistry.go`: the two roots, images and charts, in one leaf package the resources, version check, and chart lifecycle suite all read.
- `operator/internal/resources/images.go`: the image slugs and `ImageRepository(override, registry, slug)`; the three repository constants derive from it.
- `operator/internal/component/{control_plane,console,runner}.go`: each resolves its repository through `ImageRepository`.
- `operator/api/v1/planton_platform_types.go`: `spec.imageRegistry`, with a rule refusing a scheme or a trailing slash; the CRD regenerated in `config/` and `helm/planton-operator`.
- `operator/internal/platformversion/requirement.go` and the controller: `RequiredOperator` takes the resolved control-plane repository, and answers are remembered per repository and version.
- `catalog/kubernetes/kubernetesplantonplatform`: `image_registry` (field 20) and `runner.image` (field 6) on both engines.
- `catalog/kubernetes/kubernetesplantonoperator` and `kubernetesplantonrunner`: `chart_repository` (fields 17 and 12), with an `oci://` rule, on both engines.
- Generated: the three kinds' Go stubs and `reference.md`, and `pkg/protodocs/index.json.gz`.

## Compatibility

Nothing moves for an install that sets none of the new fields: every default is the address it pulled before. A manifest that sets `image_registry` needs a planton-operator chart that knows the field (0.22.0 or newer); an older definition refuses the declaration.

## Verification

- `go test` for `internal/resources`, `internal/component`, `internal/platformversion`, and `internal/controller` (envtest) in `operator/`; `go vet` on each and on `cmd`.
- `go test` for the platform module and the three kinds' spec suites; `go vet` on the three Pulumi modules; `terraform validate` on the three Terraform modules.
- `hack/guards/ensure_image_paths.sh` passes, and fails naming the file when a root is planted in operator code; the module self-containment, cross-engine parity, and provider-pin guards pass.
