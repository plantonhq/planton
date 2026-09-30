// Package eject copies the official IaC module behind a catalog component
// into a user-owned directory, ready to customize.
//
// An ejected module is a first-party redistribution of Apache-2.0 catalog
// code that lands in the user's own repository, so every eject guarantees
// three properties beyond the file copy:
//
//   - LICENSE and NOTICE are always present in the output (attribution
//     travels with the code, at every granularity).
//   - A pulumi copy compiles standalone: its self-imports are rewritten to a
//     module path the user owns and a go.mod is synthesized (see pulumi.go
//     for why keeping the original import path is impossible).
//   - Generated contract notes teach the module's input/outputs contract,
//     how to verify conformance, and how to run and register the copy.
package eject

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// NotesFileName is the generated contract-notes file written into every
// ejected module. Deliberately not README.md: the module's own README (when
// present) is preserved untouched.
const NotesFileName = "CONTRACT.md"

// Input configures a single eject operation.
type Input struct {
	// KindName is the cloud resource kind whose official module is ejected.
	// Tolerant forms accepted (AwsS3Bucket, awss3bucket, aws-s3-bucket).
	KindName string

	// Provisioner selects which module is ejected. Tofu and terraform share
	// one module (iac/tf); pulumi has its own (iac/pulumi).
	Provisioner provisioner.ProvisionerType

	// OutputDir receives the ejected module. It must not exist yet or must
	// be an empty directory — eject never overwrites user content.
	OutputDir string

	// GoModulePath (pulumi only) becomes the ejected module's Go module
	// path; the copy's self-imports are rewritten onto it.
	GoModulePath string

	// SkipGoModTidy leaves the synthesized go.mod without resolved
	// dependencies (no network access, no go toolchain requirement). The
	// contract notes always carry the command to run it later.
	SkipGoModTidy bool
}

// Result reports what an eject produced.
type Result struct {
	// KindName is the canonical kind name (e.g. AwsS3Bucket).
	KindName string

	// OutputDir is the absolute path holding the ejected module.
	OutputDir string

	// SourceVersion is the release tag (or staging checkout) the module was
	// ejected from — the provenance recorded in the contract notes.
	SourceVersion string

	// NotesPath is the absolute path of the generated contract-notes file.
	NotesPath string

	// GoModTidyRan reports whether the pulumi copy's dependencies were
	// resolved (go on PATH and tidy succeeded).
	GoModTidyRan bool
}

// Eject copies the official module for the given kind and provisioner into
// in.OutputDir and prepares it for customization.
func Eject(in Input) (*Result, error) {
	kind := crkreflect.KindFromString(in.KindName)
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return nil, errors.Errorf("unknown cloud resource kind %q — kind names follow the catalog (e.g. AwsS3Bucket)", in.KindName)
	}
	kindName := crkreflect.ExtractKindNameByKind(kind)

	switch in.Provisioner {
	case provisioner.ProvisionerTypePulumi, provisioner.ProvisionerTypeTofu, provisioner.ProvisionerTypeTerraform:
	default:
		return nil, errors.New("provisioner must be one of 'tofu', 'terraform', or 'pulumi'")
	}
	if err := provisioner.Require(kind, in.Provisioner); err != nil {
		return nil, err
	}

	outputDir, err := prepareOutputDir(in.OutputDir)
	if err != nil {
		return nil, err
	}

	source, err := resolveSourceFn(kindName, in.Provisioner)
	if err != nil {
		return nil, err
	}

	if err := fileutil.CopyDirFiltered(source.dir, outputDir, includeInEjectedCopy); err != nil {
		return nil, errors.Wrapf(err, "failed to copy the %s module into %s", kindName, outputDir)
	}

	result := &Result{
		KindName:      kindName,
		OutputDir:     outputDir,
		SourceVersion: source.version,
	}

	if in.Provisioner == provisioner.ProvisionerTypePulumi {
		if err := preparePulumiCopy(outputDir, kind, kindName, in, result); err != nil {
			return nil, err
		}
	}

	if err := ensureLicenseFiles(outputDir, source.repoRoot); err != nil {
		return nil, err
	}

	notesPath, err := writeContractNotes(outputDir, kindName, in, result)
	if err != nil {
		return nil, err
	}
	result.NotesPath = notesPath

	return result, nil
}

// prepareOutputDir normalizes the destination to an absolute path and
// enforces the never-overwrite rule: the directory must be new or empty.
func prepareOutputDir(outputDir string) (string, error) {
	if outputDir == "" {
		return "", errors.New("an output directory is required")
	}

	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve absolute path for %s", outputDir)
	}

	entries, err := os.ReadDir(absDir)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return "", errors.Wrapf(err, "failed to create output directory %s", absDir)
		}
	case err != nil:
		return "", errors.Wrapf(err, "cannot read the output directory %s", absDir)
	case len(entries) > 0:
		return "", errors.Errorf("the output directory %s is not empty — eject never overwrites existing content; point --output-dir at a new or empty directory", absDir)
	}

	return absDir, nil
}

// includeInEjectedCopy filters the module copy down to module content.
// Repo build scaffolding and run leftovers are not part of the module a user
// customizes; copying them would only confuse the ejected tree.
func includeInEjectedCopy(relPath string, d os.DirEntry) bool {
	name := d.Name()

	if d.IsDir() {
		// .terraform is tofu's own working directory; .git can appear when
		// copying from a repo checkout.
		return name != ".terraform" && name != ".git"
	}

	// Bazel build files are repo scaffolding, not module content.
	if name == "BUILD.bazel" {
		return false
	}
	// Plan files are run leftovers, never source.
	if strings.HasSuffix(name, ".tfplan") {
		return false
	}
	// Per-stack Pulumi files (Pulumi.<stack>.yaml) pin stacks that only
	// exist in the catalog's own test environments; Pulumi.yaml itself is
	// the project file and stays.
	if strings.HasPrefix(name, "Pulumi.") && strings.HasSuffix(name, ".yaml") && name != "Pulumi.yaml" {
		return false
	}

	return true
}

// moduleSubPath returns the module directory relative to a repo root for the
// given provisioner — the same derivation the deploy-path resolvers use.
func moduleSubPath(kind cloudresourcekind.CloudResourceKind, kindName string, prov provisioner.ProvisionerType) string {
	providerSegment := strings.ReplaceAll(crkreflect.GetProvider(kind).String(), "_", "")
	engineDir := "tf"
	if prov == provisioner.ProvisionerTypePulumi {
		engineDir = "pulumi"
	}
	return filepath.Join("catalog", providerSegment, strings.ToLower(kindName), "iac", engineDir)
}

// describeProvisioner names the module family in user-facing text.
func describeProvisioner(prov provisioner.ProvisionerType) string {
	if prov == provisioner.ProvisionerTypePulumi {
		return "Pulumi"
	}
	return "OpenTofu/Terraform"
}
