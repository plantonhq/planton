package profile

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
	providerv1 "github.com/plantonhq/planton/qa/providere2eprofile/v1"
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

// LoadKindProfile reads and parses a kind's E2E profile from disk.
func LoadKindProfile(repoRoot, provider, kindDir string) (*kindv1.CatalogKindE2EProfile, error) {
	path, err := KindProfilePath(repoRoot, provider, kindDir)
	if err != nil {
		return nil, err
	}
	p := &kindv1.CatalogKindE2EProfile{}
	if err := loadYAMLProto(path, p); err != nil {
		return nil, errors.Wrapf(err, "loading catalog kind E2E profile for %s/%s", provider, kindDir)
	}
	return p, nil
}

// loadYAMLProto reads a YAML file and unmarshals it into a proto message.
// Uses the canonical YAML 1.2 conversion, then protojson to parse into proto.
func loadYAMLProto(path string, msg proto.Message) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "reading %s", path)
	}

	jsonBytes, err := protobufyaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrapf(err, "converting YAML to JSON from %s", path)
	}

	if err := unmarshalOpts.Unmarshal(jsonBytes, msg); err != nil {
		return errors.Wrapf(err, "unmarshaling proto from %s", path)
	}

	return nil
}
