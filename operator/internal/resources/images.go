package resources

import (
	"strings"

	"github.com/plantonhq/planton/operator/internal/plantonregistry"
)

// Planton's own images live at <registry>/<slug>. The registry root is one
// setting on the platform resource (spec.imageRegistry), so an install moves
// every Planton image to another registry -- the Google Artifact Registry
// mirror, a private mirror of its own -- by changing one value. Every release
// is published to ghcr.io/plantonhq/planton and copied, byte for byte, to the
// mirror at asia-south1-docker.pkg.dev/plantonhq/planton: the slug after the
// root is the same on both, which is what lets one value move them all.

// DefaultImageRegistry is the root an install pulls from when it names none.
const DefaultImageRegistry = plantonregistry.DefaultImageRegistry

// The slug of each image the operator deploys, the path after the root.
const (
	ControlPlaneImageSlug = "control-plane"
	ConsoleImageSlug      = "client-apps/web"
	RunnerImageSlug       = "runner"
)

// ImageRepository resolves a component's image repository: its own override
// when the platform sets one (a custom build, a one-off mirror), else the
// platform's registry root, else the default root -- followed by the slug.
func ImageRepository(override, registry, slug string) string {
	if override != "" {
		return override
	}
	if registry == "" {
		registry = DefaultImageRegistry
	}
	return strings.TrimSuffix(registry, "/") + "/" + slug
}
