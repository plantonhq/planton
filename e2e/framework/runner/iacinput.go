// Package runner implements the E2E test lifecycle engine.
package runner

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
)

// BuildIacInput loads a kind's hack manifest and constructs the iac-input
// YAML that Pulumi modules expect. providerConfig is nil for the harness's
// default posture (ambient credentials, empty provider block) and non-nil only
// when the kind ships the opt-in provider-config fixture (see
// LoadProviderConfigFixture).
//
// The file lands in a fresh per-call temp directory, NEVER in the shared
// module directory: a kind can be one lane's kind under test and a
// concurrent lane's dependency fixture at the same time, and a single
// module-dir path made those lanes overwrite each other's input between
// a lane's apply and its idempotency preview / destroy (live-caught: an
// AMP scenario's preview read the concurrent scraper lane's bare-fixture
// input and planned every satellite for deletion). The Terraform path is
// immune for the same reason this fix works - it already isolates per
// run via PrepareWorkDir.
//
// Returns the path to the written iac-input file.
func BuildIacInput(manifestPath string,
	providerConfig *iacinputproviderconfig.ProviderConfig) (string, error) {
	manifestObject, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load manifest from %s", manifestPath)
	}

	iacInputYaml, err := iacinput.BuildIacInputYaml(manifestObject, providerConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to build iac-input YAML")
	}

	outputDir, err := os.MkdirTemp("", "planton-e2e-iac-input-")
	if err != nil {
		return "", errors.Wrap(err, "failed to create iac-input temp dir")
	}
	outputPath := filepath.Join(outputDir, "iac-input.yaml")
	if err := os.WriteFile(outputPath, []byte(iacInputYaml), 0600); err != nil {
		return "", errors.Wrapf(err, "failed to write iac-input to %s", outputPath)
	}

	return outputPath, nil
}
