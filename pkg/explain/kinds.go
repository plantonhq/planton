package explain

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/protodocs"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// DefaultEngine explains this repo's own schema universe: the infra-component
// kinds, with the shared option family and the embedded proto docs. Hosts
// composing additional universes (platform APIs, extra option families,
// kind-valued dispatchers) construct their own Engine and reuse the pieces.
func DefaultEngine() *Engine {
	return &Engine{
		Interpreters: []OptionInterpreter{SharedOptions},
		Docs:         protodocs.Lookup,
	}
}

// KindResource builds the explainable Resource for a catalog kind
// from the descriptors compiled into the binary.
func KindResource(kind catalogkind.CatalogKind) (Resource, error) {
	instance, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		return Resource{}, err
	}
	return Resource{
		Name:       kind.String(),
		ApiVersion: catalogkindreflect.GroupVersion(kind),
		Message:    instance.ProtoReflect().Descriptor(),
	}, nil
}

// ResolveKindName resolves a user-typed kind name (AwsVpc, aws-vpc, aws_vpc)
// to its explainable Resource.
func ResolveKindName(name string) (Resource, error) {
	kind := catalogkindreflect.KindFromString(name)
	if kind == catalogkind.CatalogKind_unspecified {
		return Resource{}, errors.Errorf("unknown catalog kind %q -- run `planton explain --list` to see all kinds", name)
	}
	return KindResource(kind)
}

// KindNames returns every explainable kind name, sorted.
func KindNames() []string {
	kinds := catalogkindreflect.KindsList()
	names := make([]string, 0, len(kinds))
	for _, k := range kinds {
		names = append(names, k.String())
	}
	sort.Strings(names)
	return names
}
