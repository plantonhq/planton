// Package plantonregistry is the operator's one home for where Planton's own
// images and charts are published. Every other file derives from these two
// roots, and hack/guards/ensure_image_paths.sh refuses a registry root written
// anywhere else, so moving the defaults is a change to this file alone.
//
// Every release is published to ghcr.io and copied, byte for byte, to the
// Google Artifact Registry mirror at asia-south1-docker.pkg.dev/plantonhq; the
// path after the root is the same on both.
package plantonregistry

// DefaultImageRegistry is the root Planton's images are pulled from when a
// platform names none (spec.imageRegistry).
const DefaultImageRegistry = "ghcr.io/plantonhq/planton"

// DefaultChartRepository is the OCI path Planton's Helm charts are published
// under.
const DefaultChartRepository = "oci://ghcr.io/plantonhq/charts"
