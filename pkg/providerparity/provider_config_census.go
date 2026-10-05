//go:build !codegen
// +build !codegen

// Provider-config census: the contract side of PROVIDER-BLOCK parity. Where
// the spec census walks each kind's spec, this walks the provider's one
// config proto (catalog/<provider>/provider.proto) and enumerates its leaf
// fields -- the exact surface an IaC input's provider_config can express.
// The provider-config accounting joins it against the distilled schema's
// provider block exactly like depth accounting joins specs against resource
// blocks.

package providerparity

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/iacinput/providerdetect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// configPathRoot roots every provider-config census path (the "spec" analogue
// for the provider block): census paths and manifest references read
// "config.assume_role_chain.role_arn".
const configPathRoot = "config"

// ProviderConfigCensus enumerates the leaf fields of one provider's config
// proto, sorted, rooted at "config". The provider->proto pairing comes from
// providerdetect.ProviderConfigProto -- the same map the CLI's -p validation
// trusts, so the census can never disagree with the runtime about which type
// a provider's config is.
func ProviderConfigCensus(provider catalogkind.CatalogProvider) ([]string, error) {
	msg, err := providerdetect.ProviderConfigProto(provider)
	if err != nil {
		return nil, errors.Wrapf(err, "no provider config proto for %s", provider.String())
	}
	md := msg.ProtoReflect().Descriptor()
	if md.Fields().Len() == 0 {
		return nil, errors.Errorf("provider config proto for %s has no fields", provider.String())
	}
	return collectConfigPaths(md), nil
}

// collectConfigPaths reuses the spec census's descriptor walk (one walk, two
// censuses -- a reader who knows one knows both) with the config root.
func collectConfigPaths(md protoreflect.MessageDescriptor) []string {
	return CollectSpecPaths(md, configPathRoot)
}
