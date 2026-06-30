package profile

import (
	"os"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	componentv1 "github.com/plantonhq/planton/apis/dev/planton/qa/componente2eprofile/v1"
	providerv1 "github.com/plantonhq/planton/apis/dev/planton/qa/providere2eprofile/v1"
)

var unmarshalOpts = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

// LoadProviderProfile reads and parses a provider's E2E profile from disk.
func LoadProviderProfile(repoRoot, provider string) (*providerv1.ProviderE2EProfile, error) {
	p := &providerv1.ProviderE2EProfile{}
	if err := loadYAMLProto(ProviderProfilePath(repoRoot, provider), p); err != nil {
		return nil, errors.Wrapf(err, "loading provider E2E profile for %s", provider)
	}
	return p, nil
}

// LoadComponentProfile reads and parses a component's E2E profile from disk.
func LoadComponentProfile(repoRoot, provider, component string) (*componentv1.ComponentE2EProfile, error) {
	p := &componentv1.ComponentE2EProfile{}
	if err := loadYAMLProto(ComponentProfilePath(repoRoot, provider, component), p); err != nil {
		return nil, errors.Wrapf(err, "loading component E2E profile for %s/%s", provider, component)
	}
	return p, nil
}

// loadYAMLProto reads a YAML file and unmarshals it into a proto message.
// Uses sigs.k8s.io/yaml to convert YAML→JSON, then protojson to parse into proto.
func loadYAMLProto(path string, msg proto.Message) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "reading %s", path)
	}

	jsonBytes, err := yaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrapf(err, "converting YAML to JSON from %s", path)
	}

	if err := unmarshalOpts.Unmarshal(jsonBytes, msg); err != nil {
		return errors.Wrapf(err, "unmarshaling proto from %s", path)
	}

	return nil
}
