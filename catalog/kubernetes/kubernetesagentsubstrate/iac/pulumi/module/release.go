package module

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// What the module installs, and where it reads it.
//
// Agent Substrate publishes a release as a git tag of
// github.com/agent-substrate/substrate and its images at
// ghcr.io/agent-substrate/substrate, tagged with the release: no chart and no
// install manifest. The module reads what upstream publishes, at deploy time
// and at the tag spec.version names:
//   - the install's own files (manifests/, the database setup script) and the
//     CRDs (manifests/ate-install/generated/), raw from the tag;
//   - each image upstream's manifests name by placeholder, rewritten to
//     <registry>/<image>:<tag>@<digest>, the digest being the one upstream's
//     registry serves for the tag, exactly as upstream's own installer pins
//     them (cmd/ate-setup/internal/images/prebuilt.go, Pin).
//
// A NEW UPSTREAM RELEASE NEEDS NOTHING FROM PLANTON: set spec.version to it.
// The module changes only when upstream changes what this composition reads:
// a file below moves, an object, container or patch target the module edits
// is renamed, or a flag the module sets is removed. Every case but the last is
// refused before anything is applied (a missing file, image or patch target);
// a removed flag fails the component's rollout. vars.Release, what an empty
// spec.version installs, is the only per-release pin, and moving it is
// optional.
//
// Terraform twin: locals.tf (the same base URL, files and placeholder rules)
// and images.tf (the same digest lookup); TestTerraformTwinPinsTheSameRelease
// holds them equal.

// UpstreamRegistry is where upstream publishes every image of a release.
const UpstreamRegistry = "ghcr.io/agent-substrate/substrate"

// upstreamFilesBase serves a tag's files raw.
const upstreamFilesBase = "https://raw.githubusercontent.com/agent-substrate/substrate/"

// UpstreamFileURL is one file of upstream's tree at a release tag.
func UpstreamFileURL(release, path string) string {
	return upstreamFilesBase + release + "/" + path
}

// crdFiles are upstream's definitions, one file per kind, read at the tag
// spec.version names.
var crdFiles = []string{
	"manifests/ate-install/generated/ate.dev_csidriverconfigs.yaml",
	"manifests/ate-install/generated/ate.dev_sandboxconfigs.yaml",
	"manifests/ate-install/generated/ate.dev_workerpools.yaml",
}

// CRDBundleURLTemplates are the definitions as kept-CRD bundle URLs, with the
// version left for the primitive to fill in.
func CRDBundleURLTemplates() []string {
	urls := make([]string, len(crdFiles))
	for i, path := range crdFiles {
		urls[i] = upstreamFilesBase + "{{version}}/" + path
	}
	return urls
}

// The image placeholders upstream's manifests carry, which upstream's
// installer resolves at install time (images/prebuilt.go rewrite and
// steps/overlay.go patchEnvoyDataplaneImage):
//   - ko://go.ate.dev/substrate/cmd/<path> for every image upstream builds
//     from a Go package, the image named by the path's last element
//     (images.ImageName);
//   - ${ENVOY_DATAPLANE_IMAGE} for the one image upstream builds from a
//     Dockerfile.
//
// Only image: values are read: a file may carry other ${...} text (the
// bundled database's start script), which is upstream's and stays as it is.
const (
	koImagePrefix    = "ko://go.ate.dev/substrate/cmd/"
	envoyPlaceholder = "${ENVOY_DATAPLANE_IMAGE}"
	envoyImage       = "envoy-dataplane"
)

// workerImages are the release's worker images, which no manifest the module
// applies names: a worker pool runs them, and the outputs hand them over.
var workerImages = []string{"ateworker-gvisor", "ateworker-microvm"}

// imageLine is one image: field of a manifest, whatever its indentation or
// list marker: group 1 is everything before the value, group 2 the value
// (quoted or not), group 3 a trailing comment.
var imageLine = regexp.MustCompile(`(?m)^([ \t]*(?:-[ \t]+)?image:[ \t]*)("[^"\n]*"|'[^'\n]*'|[^\s#]+)([ \t]+#.*)?[ \t]*$`)

// placeholderImage is the image a placeholder names, or "" when value is a
// literal image (a third-party image upstream pins itself). An unknown
// placeholder is refused: upstream resolves it by a rule this module does not
// know, so installing it would pull nothing.
func placeholderImage(path, value string) (string, error) {
	value = strings.Trim(value, `"'`)
	switch {
	case value == envoyPlaceholder:
		return envoyImage, nil
	case strings.HasPrefix(value, koImagePrefix):
		pkg := strings.TrimPrefix(value, koImagePrefix)
		return pkg[strings.LastIndex(pkg, "/")+1:], nil
	case strings.HasPrefix(value, "ko://") || strings.HasPrefix(value, "${"):
		return "", fmt.Errorf("observed: %s names the image %q, a placeholder this module does not know how to resolve\nmeaning: upstream changed how its manifests name an image, so the release cannot be installed as this module composes it\nnext step: set spec.version to a release this module installs (%s installs), or update the module to the new placeholder (pin needs a catalog fix)", path, value, vars.Release)
	}
	return "", nil
}

// releaseImages are every image the release's files name by placeholder,
// plus its worker images, sorted.
func releaseImages(f releaseFileSet) ([]string, error) {
	seen := map[string]bool{}
	for _, image := range workerImages {
		seen[image] = true
	}
	for path, text := range f.files {
		for _, match := range imageLine.FindAllStringSubmatch(text, -1) {
			image, err := placeholderImage(path, match[2])
			if err != nil {
				return nil, err
			}
			if image != "" {
				seen[image] = true
			}
		}
	}
	images := make([]string, 0, len(seen))
	for image := range seen {
		images = append(images, image)
	}
	sort.Strings(images)
	return images, nil
}

// digestLookup answers the digest the registry serves for a tagged
// reference.
type digestLookup func(reference string) (string, error)

// registryDigest asks upstream's registry, anonymously, for the digest of a
// tagged image: the multi-platform index, as a container runtime resolves it.
func registryDigest(reference string) (string, error) {
	ref, err := name.ParseReference(reference)
	if err != nil {
		return "", err
	}
	descriptor, err := remote.Head(ref, remote.WithAuth(authn.Anonymous))
	if err != nil {
		return "", err
	}
	return descriptor.Digest.String(), nil
}

// resolveImages pins each image of the release: name -> its reference at the
// release's tag and the digest upstream's registry serves for it.
func resolveImages(release string, images []string, lookup digestLookup) (map[string]string, error) {
	pinned := map[string]string{}
	for _, image := range images {
		tagged := UpstreamRegistry + "/" + image + ":" + release
		digest, err := lookup(tagged)
		if err != nil {
			return nil, imageLookupRefusal(release, tagged, err)
		}
		pinned[image] = tagged + "@" + digest
	}
	return pinned, nil
}

// imageLookupRefusal explains an image the registry would not resolve, by
// what its answer means: an image the release does not publish, or a registry
// this deploy cannot reach.
func imageLookupRefusal(release, tagged string, err error) error {
	var terr *transport.Error
	if errors.As(err, &terr) && (terr.StatusCode == 404 || terr.StatusCode == 401 || terr.StatusCode == 403) {
		// ghcr.io answers an image that does not exist as unauthorized to an
		// anonymous reader, so 401 and 403 mean "not published" here too.
		return fmt.Errorf("observed: %s is not published (the registry answered HTTP %d)\nmeaning: %s is not an Agent Substrate release, or it does not publish this image\nnext step: set spec.version to a published release (https://github.com/agent-substrate/substrate/releases), or leave it empty for %s", tagged, terr.StatusCode, release, vars.Release)
	}
	return fmt.Errorf("observed: the digest of %s could not be read (%s)\nmeaning: the registry does not resolve, or egress to it is blocked from where this deploy runs; the module pins every image by the digest upstream's registry serves, as upstream's installer does\nnext step: confirm DNS and egress to ghcr.io (`curl -I https://ghcr.io/v2/`), then re-run", tagged, err.Error())
}

// rewriteImages replaces every image placeholder in the release's files with
// its pinned reference, leaving every other line as upstream wrote it.
func rewriteImages(f releaseFileSet, pinned map[string]string) (releaseFileSet, error) {
	out := releaseFileSet{files: make(map[string]string, len(f.files))}
	for path, text := range f.files {
		var failure error
		out.files[path] = imageLine.ReplaceAllStringFunc(text, func(line string) string {
			match := imageLine.FindStringSubmatch(line)
			image, err := placeholderImage(path, match[2])
			if err != nil {
				failure = err
				return line
			}
			if image == "" {
				return line
			}
			return match[1] + pinned[image] + match[3]
		})
		if failure != nil {
			return releaseFileSet{}, failure
		}
	}
	return out, nil
}
