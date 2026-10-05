package eject

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/staging"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumisource"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofuzip"
)

// resolvedSource is where an eject copies from.
type resolvedSource struct {
	// dir is the official module directory to copy.
	dir string

	// version records the release tag (or staging checkout) the module
	// content corresponds to — the provenance the contract notes carry.
	version string

	// repoRoot is set when dir lives inside a repo checkout (the staging
	// clone). Release artifacts embed LICENSE and NOTICE themselves; a repo
	// checkout keeps them at its root, so the eject backfills from here.
	repoRoot string
}

// resolveSourceFn is the source-resolution seam, replaceable in tests so
// the eject pipeline can be exercised hermetically (the real resolution
// downloads artifacts or clones the staging repo).
var resolveSourceFn = resolveSource

// resolveSource locates the official module source, mirroring the deploy
// path's resolution order: release artifact first (fast, cached), git
// staging clone as the fallback. A released CLI whose tag predates a given
// artifact self-corrects through the fallback (a git checkout of that tag).
func resolveSource(kindName string, prov provisioner.ProvisionerType) (*resolvedSource, error) {
	targetVersion := ""
	if version.Version != "" && version.Version != version.DefaultVersion {
		targetVersion = version.Version
	}

	if targetVersion != "" {
		source, err := resolveFromArtifact(kindName, prov, targetVersion)
		if err != nil {
			cliprint.PrintWarning(fmt.Sprintf("Release artifact unavailable, falling back to the git staging area: %v", err))
		} else {
			return source, nil
		}
	}

	return resolveFromStaging(kindName, prov, targetVersion)
}

func resolveFromArtifact(kindName string, prov provisioner.ProvisionerType, releaseVersion string) (*resolvedSource, error) {
	var dir string
	var err error
	if prov == provisioner.ProvisionerTypePulumi {
		dir, err = pulumisource.EnsureSource(kindName, releaseVersion)
	} else {
		dir, err = tofuzip.EnsureModule(kindName, releaseVersion)
	}
	if err != nil {
		return nil, err
	}
	return &resolvedSource{dir: dir, version: releaseVersion}, nil
}

func resolveFromStaging(kindName string, prov provisioner.ProvisionerType, targetVersion string) (*resolvedSource, error) {
	cliprint.PrintStep("Ensuring staging area is ready...")
	if err := staging.EnsureStaging(targetVersion); err != nil {
		return nil, errors.Wrap(err, "failed to ensure the git staging area")
	}

	repoRoot, err := staging.GetStagingRepoPath()
	if err != nil {
		return nil, err
	}

	var dir string
	if prov == provisioner.ProvisionerTypePulumi {
		dir, err = pulumimodule.GetLocalModulePath(repoRoot, kindName)
	} else {
		dir, err = tofumodule.GetLocalModulePath(repoRoot, kindName)
	}
	if err != nil {
		kind := catalogkindreflect.KindFromString(kindName)
		return nil, errors.Wrapf(err,
			"the official %s module for %s was not found in the staging checkout — expected at %s; run 'planton pull' to refresh the staging area",
			describeProvisioner(prov), kindName, moduleSubPath(kind, kindName, prov))
	}

	sourceVersion := targetVersion
	if sourceVersion == "" {
		if stagingVersion, err := staging.GetCurrentStagingVersion(); err == nil && stagingVersion != "" {
			sourceVersion = stagingVersion
		}
	}
	if sourceVersion == "" {
		sourceVersion = version.DefaultVersion
	}

	return &resolvedSource{dir: dir, version: sourceVersion, repoRoot: repoRoot}, nil
}
