package resources

import "testing"

// One registry root moves every Planton image; a component's own override
// still wins, and an install that names nothing keeps pulling from ghcr.io.
func TestImageRepository(t *testing.T) {
	const mirror = "asia-south1-docker.pkg.dev/plantonhq/planton"
	cases := []struct {
		name, override, registry, slug, want string
	}{
		{"default root", "", "", ControlPlaneImageSlug, "ghcr.io/plantonhq/planton/control-plane"},
		{"registry root", "", mirror, ConsoleImageSlug, mirror + "/client-apps/web"},
		{"trailing slash tolerated", "", mirror + "/", RunnerImageSlug, mirror + "/runner"},
		{"override wins over the root", "example.com/runner", mirror, RunnerImageSlug, "example.com/runner"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ImageRepository(tc.override, tc.registry, tc.slug); got != tc.want {
				t.Errorf("ImageRepository(%q, %q, %q) = %q, want %q", tc.override, tc.registry, tc.slug, got, tc.want)
			}
		})
	}
}

// The historical constants are the default root plus the slug, so a platform
// that never set spec.imageRegistry renders the same repositories as before.
func TestDefaultRepositoriesUnchanged(t *testing.T) {
	for got, want := range map[string]string{
		ImageRepository("", "", ControlPlaneImageSlug): "ghcr.io/plantonhq/planton/control-plane",
		ImageRepository("", "", ConsoleImageSlug):      "ghcr.io/plantonhq/planton/client-apps/web",
		ImageRepository("", "", RunnerImageSlug):       "ghcr.io/plantonhq/planton/runner",
	} {
		if got != want {
			t.Errorf("default repository = %q, want %q", got, want)
		}
	}
}
