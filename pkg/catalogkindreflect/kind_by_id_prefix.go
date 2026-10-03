package catalogkindreflect

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// kindIdPrefixIndex maps each kind's id prefix to its kind. Built once;
// construction fails if two kinds claim the same prefix, because resource
// ids embed the prefix as the only type discriminator — a shared prefix
// would decode ids as the wrong kind.
var kindIdPrefixIndex = sync.OnceValues(buildKindIdPrefixIndex)

func buildKindIdPrefixIndex() (map[string]catalogkind.CatalogKind, error) {
	index := make(map[string]catalogkind.CatalogKind)
	for _, kind := range KindsList() {
		kindMeta, err := KindMeta(kind)
		if err != nil || kindMeta.IdPrefix == "" {
			continue
		}
		if owner, taken := index[kindMeta.IdPrefix]; taken {
			return nil, errors.Errorf(
				"catalog kind registry is ambiguous: %s and %s share the id prefix %q; "+
					"id prefixes must be unique across the registry so every resource id resolves to exactly one kind — change one of them",
				owner, kind, kindMeta.IdPrefix)
		}
		index[kindMeta.IdPrefix] = kind
	}
	return index, nil
}

// KindByIdPrefix takes an id prefix as input and returns the corresponding CatalogKind.
func KindByIdPrefix(idPrefix string) (catalogkind.CatalogKind, error) {
	index, err := kindIdPrefixIndex()
	if err != nil {
		return catalogkind.CatalogKind_unspecified, err
	}
	if kind, ok := index[idPrefix]; ok {
		return kind, nil
	}
	return catalogkind.CatalogKind_unspecified,
		errors.Errorf("no matching CatalogKind found for id prefix: %s", idPrefix)
}
