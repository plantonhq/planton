package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kubernetesagentsubstratev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsubstrate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Agent Substrate control plane from Planton's build
// of the release spec.version names (vars.Release when empty): upstream's own
// manifests with every image digest-pinned, composed the way upstream's
// installer composes them, with the typed settings applied onto upstream's
// own objects (compose.go).
//
// ORDER: everything that can refuse runs before anything registers
// (preflight.go, fetchRelease): a namespace other than upstream's, a workload
// or namespace another installation holds, a release move upstream cannot
// make, and a release Planton has not published or one of another install
// shape. Then
// the namespaces, the kept definitions (ownership-checked, never downgraded),
// the generated and derived Secrets, and the four document phases, each
// awaited before the next: prerequisites (the pod-certificate controller
// every certificate comes from), database, system, runtime.
//
// DESTROY: in reverse. The default SandboxConfig goes first, while the
// controller that clears its finalizer still runs; the namespaces last. The
// definitions are kept unless crds.keep_on_uninstall is false, so worker
// pools and their actors are not deleted with the control plane; the bundled
// database's volume is kept by its StatefulSet, and the upgrade refusal reads
// its stamp.
func Resources(ctx *pulumi.Context, iacInput *kubernetesagentsubstratev1alpha1.KubernetesAgentSubstrateIacInput) error {
	locals := initializeLocals(iacInput)
	s := resolveSettings(locals.Spec)
	s.identityLabels = locals.IdentityLabels
	if err := preflight(iacInput.ProviderConfig, s); err != nil {
		return err
	}

	files, err := fetchRelease(s.version)
	if err != nil {
		return err
	}
	composed, err := compose(s, files)
	if err != nil {
		return err
	}

	// The definitions belong to the install shape, so they stay on the
	// module's release whatever release spec.version picks, and moving
	// between releases of the shape never reads as a definition downgrade.
	createdCrds, err := keptcrds.Apply(ctx, keptcrds.Args{
		Source: helmcrds.Source{
			BundleURL:  CRDBundleURLTemplate(),
			BundleName: vars.BundleName,
			Version:    vars.Release,
		},
		Policy:          helmcrds.Policy{ExpectCRDs: true},
		ReleaseName:     locals.ResourceName,
		Namespace:       vars.Namespace,
		Install:         s.crdsInstall,
		KeepOnUninstall: s.crdsKeep,
		ProviderConfig:  iacInput.ProviderConfig,
		ProviderName:    "kubernetes-crd-upsert",
	})
	if err != nil {
		return errors.Wrap(err, "failed to apply the Agent Substrate resource definitions")
	}

	provider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	dependsOn := createdCrds
	group := func(suffix string, objects []document) error {
		if len(objects) == 0 {
			return nil
		}
		applied, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-"+suffix,
			&pulumiyaml.ConfigGroupArgs{Objs: objects},
			pulumi.Provider(provider), pulumi.DependsOn(dependsOn))
		if err != nil {
			return errors.Wrapf(err, "failed to apply the Agent Substrate %s", suffix)
		}
		dependsOn = []pulumi.Resource{applied}
		return nil
	}

	if err := group("namespaces", composed.namespaces); err != nil {
		return err
	}
	secrets, err := applySecrets(ctx, s, locals.IdentityLabels, pulumi.Provider(provider), pulumi.DependsOn(dependsOn))
	if err != nil {
		return err
	}
	dependsOn = append(dependsOn, secrets...)
	for _, phase := range []struct {
		suffix  string
		objects []document
	}{
		{"prerequisites", composed.prerequisites},
		{"database", composed.database},
		{"system", composed.system},
		{"runtime", composed.runtime},
	} {
		if err := group(phase.suffix, phase.objects); err != nil {
			return err
		}
	}

	exportOutputs(ctx, s, composed.images)
	return nil
}

// fetchRelease downloads the release the spec names from where Planton
// serves releases.
func fetchRelease(release string) (releaseFileSet, error) {
	return fetchReleaseFrom(releaseBaseURL, release)
}

// fetchReleaseFrom downloads a release from base: its seal first, then its
// image list, whose install shape must be the one this module composes, then
// every other file the composition reads. Each file is checked against the
// seal, so the composition reads exactly the bytes Planton published. Every
// refusal comes before anything is applied.
func fetchReleaseFrom(base, release string) (releaseFileSet, error) {
	sums, err := fetchText(base+release+"/"+fileChecksums, absent{
		meaning: fmt.Sprintf("Planton has not published Agent Substrate %s: every published release carries SHA256SUMS, uploaded last", release),
		next:    fmt.Sprintf("set spec.version to a release Planton publishes (its %s answers 200), or leave it empty for %s", ReleaseFileURL("<version>", fileChecksums), vars.Release),
	})
	if err != nil {
		return releaseFileSet{}, err
	}
	seal := parseChecksums(sums)
	set := releaseFileSet{files: map[string]string{}}
	fetch := func(path string) error {
		url := base + release + "/" + path
		text, err := fetchText(url, absent{
			meaning: fmt.Sprintf("Agent Substrate %s is published without %s, which every release of its install shape carries", release, path),
			next:    "report it to Planton; a published release never changes, so re-running will not help",
		})
		if err != nil {
			return err
		}
		if sum := sha256.Sum256([]byte(text)); hex.EncodeToString(sum[:]) != seal[path] {
			return fmt.Errorf("observed: %s does not match the release's SHA256SUMS\nmeaning: the download was altered on its way, or the file changed after the release was sealed\nnext step: re-run; if it repeats, report it to Planton", url)
		}
		set.files[path] = text
		return nil
	}
	if err := fetch(fileImages); err != nil {
		return releaseFileSet{}, err
	}
	if err := refuseForeignShape(release, set.files[fileImages]); err != nil {
		return releaseFileSet{}, err
	}
	for _, path := range releaseFiles {
		if path == fileImages {
			continue
		}
		if err := fetch(path); err != nil {
			return releaseFileSet{}, err
		}
	}
	return set, nil
}

// parseChecksums reads SHA256SUMS (`sha256sum` format) into path -> digest.
func parseChecksums(text string) map[string]string {
	sums := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		if digest, path, ok := strings.Cut(line, "  "); ok {
			sums[path] = digest
		}
	}
	return sums
}

// refuseForeignShape refuses a release whose install is not the one this
// module composes. A release published before shapes were recorded is its
// own.
func refuseForeignShape(release, imagesJSON string) error {
	var list struct {
		Shape string `json:"shape"`
	}
	if err := json.Unmarshal([]byte(imagesJSON), &list); err != nil {
		return errors.Wrapf(err, "parsing Agent Substrate %s's images.json", release)
	}
	shape := list.Shape
	if shape == "" {
		shape = release
	}
	if shape == vars.Release {
		return nil
	}
	return fmt.Errorf("observed: Agent Substrate %s has the install shape of %s, and this KubernetesAgentSubstrate composes %s's\nmeaning: %s installs differently from what this Planton catalog release knows how to compose, so nothing was applied\nnext step: set spec.version to %s or a release of its shape, or deploy this resource from a Planton catalog release that composes %s's shape", release, shape, vars.Release, release, vars.Release, shape)
}
