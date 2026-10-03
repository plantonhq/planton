package pulumistack

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// captureOutputs reads the just-updated stack's outputs and fills sink with
// the raw map, the flattened map, the kind's secret outputs, and its typed
// Outputs proto (honoring module-shipped transform overrides via the
// module directory). The read shows secrets, because downstream reference
// resolution needs the real values; which outputs are secrets comes from the
// kind's schema, never from what the engine happened to mask.
func captureOutputs(
	stackFqdn string,
	moduleRepoPath string,
	kindName string,
	extraEnv []string,
	sink *outputs.CaptureResult,
) error {
	shown, err := readOutputs(stackFqdn, moduleRepoPath, extraEnv)
	if err != nil {
		return errors.Wrap(err, "failed to read outputs with --show-secrets")
	}

	sink.Raw = shown
	sink.Flat = outputs.Flatten(shown)

	kind := catalogkindreflect.KindFromString(kindName)
	if kind == catalogkind.CatalogKind_unspecified {
		return errors.Errorf("cannot resolve catalog kind from %q for output transformation", kindName)
	}
	// A kind whose schema cannot be read leaves Secrets empty, and every
	// output then renders masked.
	sink.Secrets, _ = outputs.SecretOutputs(kind)

	typed, flat, err := outputs.TransformRaw(kind, shown, &outputs.TransformOptions{ModuleDir: moduleRepoPath})
	if err != nil {
		return errors.Wrapf(err, "output transformation failed for kind %s", kindName)
	}
	sink.Typed = typed
	if flat != nil {
		sink.Flat = flat
	}

	return nil
}

// readOutputs runs `pulumi stack output --json --show-secrets` and
// decodes the plain name->value map.
func readOutputs(stackFqdn, moduleRepoPath string, extraEnv []string) (map[string]interface{}, error) {
	args := []string{"stack", "output", "--stack", stackFqdn, "--json", "--non-interactive", "--show-secrets"}

	cmd := exec.Command("pulumi", args...)
	cmd.Dir = moduleRepoPath
	cmd.Env = append(os.Environ(), extraEnv...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "failed to execute pulumi %v", args)
	}

	values := map[string]interface{}{}
	if len(bytes.TrimSpace(stdout.Bytes())) > 0 {
		if err := json.Unmarshal(stdout.Bytes(), &values); err != nil {
			return nil, errors.Wrap(err, "output document is not a JSON object")
		}
	}
	return values, nil
}
