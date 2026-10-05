package runner

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/iacinput/providerdetect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/encoding/protojson"
)

// ProviderConfigFixtureName is the OPT-IN per-kind provider-config
// fixture, at the kind's e2e/ root beside profile.yaml. Absent file
// means exactly the harness's historical contract: a nil provider config,
// credentials from the ambient chain, an empty provider block.
//
// The fixture exists so provider-BLOCK surface (default tags, retry tuning,
// ...) can reach PROVEN through a real dual-engine run. It is non-secret by
// design: credentials stay on the harness's ambient chain, and the fixture
// carries only the provider-block arguments the lane proves (default_tags is
// the natural proof -- observable on every created resource, zero secrets in
// the repo).
const ProviderConfigFixtureName = "provider-config.yaml"

// LoadProviderConfigFixture returns the kind's provider-config fixture,
// or nil when the kind ships none. moduleDir anchors the lookup
// (catalog/<p>/<kind>/iac/<engine> -> the kind's e2e/ dir) because the
// manifest path may point at a token-expanded temp copy. A present-but-
// invalid fixture is a hard error: it is an explicit authoring act, exactly
// like the parity manifests.
func LoadProviderConfigFixture(moduleDir, manifestPath string) (*iacinputproviderconfig.ProviderConfig, error) {
	fixturePath := filepath.Join(moduleDir, "..", "..", "e2e", ProviderConfigFixtureName)
	if _, err := os.Stat(fixturePath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "checking provider-config fixture %s", fixturePath)
	}

	detected, err := detectProvider(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "detecting provider for fixture %s", fixturePath)
	}
	if err := providerdetect.ValidateProviderConfig(fixturePath, detected); err != nil {
		return nil, errors.Wrapf(err, "provider-config fixture %s is invalid", fixturePath)
	}
	return &iacinputproviderconfig.ProviderConfig{
		Path:     fixturePath,
		Provider: detected,
	}, nil
}

// detectProvider reads the manifest and names the cloud provider its kind
// belongs to, the way the CLI does before it reads a provider configuration.
func detectProvider(manifestPath string) (catalogkind.CatalogProvider, error) {
	manifestObject, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return 0, errors.Wrapf(err, "loading manifest %s to detect its provider", manifestPath)
	}
	manifestJson, err := protojson.Marshal(manifestObject)
	if err != nil {
		return 0, errors.Wrap(err, "marshaling manifest for provider detection")
	}
	detection, err := providerdetect.DetectFromManifest(manifestJson)
	if err != nil {
		return 0, err
	}
	return detection.Provider, nil
}
