package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesagentsubstratev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsubstrate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Agent Substrate control plane of the release
// spec.version names (vars.Release when empty), straight from what upstream
// publishes (release.go): the tag's own manifests, every image pinned by the
// digest upstream's registry serves, composed the way upstream's installer
// composes them, with the typed settings applied onto upstream's own objects
// (compose.go).
//
// ORDER: everything that can refuse runs before anything registers
// (preflight.go, fetchRelease, resolveImages): a namespace other than
// upstream's, a workload or namespace another installation holds, a release
// move upstream cannot make, and a release whose files or images upstream
// does not publish or whose files this composition cannot read. Then
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
	images, err := releaseImages(files)
	if err != nil {
		return err
	}
	pinned, err := resolveImages(s.version, images, registryDigest)
	if err != nil {
		return err
	}
	composed, err := compose(s, files, pinned)
	if err != nil {
		return err
	}

	// The definitions are the release's own, read at its tag, so a release
	// that changes them installs its changes. The upgrade rule (upgrade.go)
	// judges every move between releases, a rollback included, and upstream's
	// rollback restores the older definitions, so the primitive's
	// never-downgrade refusal defers to it.
	createdCrds, err := keptcrds.Apply(ctx, keptcrds.Args{
		Source: helmcrds.Source{
			BundleURLs:   CRDBundleURLTemplates(),
			BundleName:   vars.BundleName,
			Version:      s.version,
			VersionField: "spec.version",
		},
		Policy:          helmcrds.Policy{ExpectCRDs: true, DowngradeJudgedByKind: true},
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
	// Scoped to the ClusterTrustBundles alone: a cluster-scoped bundle outlives
	// a destroy, and the next install adopts it.
	bundleProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfigUpsert(ctx, iacInput.ProviderConfig, "kubernetes-trust-bundle-upsert")
	if err != nil {
		return errors.Wrap(err, "failed to set up the trust-bundle kubernetes provider")
	}
	secrets, err := applySecrets(ctx, s, locals.IdentityLabels, bundleProvider, pulumi.Provider(provider), pulumi.DependsOn(dependsOn))
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

// fetchRelease downloads every file the composition reads from upstream's
// tree at the release tag. Every refusal comes before anything is applied.
func fetchRelease(release string) (releaseFileSet, error) {
	return fetchReleaseFrom(upstreamFilesBase, release)
}

// fetchReleaseFrom downloads the release's files from base (upstream's raw
// files; a test server in the tests). A file the tag does not carry is
// either a version upstream never released or a release that moved a file
// this composition reads.
func fetchReleaseFrom(base, release string) (releaseFileSet, error) {
	set := releaseFileSet{files: map[string]string{}}
	for _, path := range releaseFiles {
		text, err := fetchText(base+release+"/"+path, absent{
			meaning: fmt.Sprintf("Agent Substrate %s has no %s: it is not a release upstream published, or upstream moved a file this module's composition reads", release, path),
			next:    fmt.Sprintf("set spec.version to a published release (https://github.com/agent-substrate/substrate/releases), or leave it empty for %s; a published release that moved this file needs a catalog update", vars.Release),
		})
		if err != nil {
			return releaseFileSet{}, err
		}
		set.files[path] = text
	}
	return set, nil
}
