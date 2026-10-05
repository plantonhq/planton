package catalogkindreflect

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// canonicalKindName lowercases a kind name and strips separators so that
// names differing only in case or separator style compare equal. Registry
// uniqueness is enforced on this form: two kinds whose names collide
// canonically are ambiguous on at least one resolution surface even when
// their exact spellings differ.
func canonicalKindName(name string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(name))
}

// kindNameIndex maps each kind's effective name (kind_meta.name, falling
// back to the enum value name) to its kind. Built once; construction fails
// if two kinds claim the same name, because a manifest's kind must resolve
// to exactly one type — first-match-wins over an unordered walk resolved
// duplicates randomly.
var kindNameIndex = sync.OnceValues(buildKindNameIndex)

func buildKindNameIndex() (map[string]catalogkind.CatalogKind, error) {
	index := make(map[string]catalogkind.CatalogKind)
	canonicalOwner := make(map[string]catalogkind.CatalogKind)
	for _, kind := range KindsList() {
		effectiveName := kind.String()
		if kindMeta, err := KindMeta(kind); err == nil && kindMeta.Name != "" {
			effectiveName = kindMeta.Name
		}
		canonicalName := canonicalKindName(effectiveName)
		if owner, taken := canonicalOwner[canonicalName]; taken {
			return nil, errors.Errorf(
				"catalog kind registry is ambiguous: %s and %s both answer to the kind name %q; "+
					"kind names must be unique across the registry so every manifest resolves to exactly one kind — rename one of them",
				owner, kind, effectiveName)
		}
		canonicalOwner[canonicalName] = kind
		index[effectiveName] = kind
	}
	return index, nil
}

func KindByKindName(kindName string) (catalogkind.CatalogKind, error) {
	index, err := kindNameIndex()
	if err != nil {
		return catalogkind.CatalogKind_unspecified, err
	}
	if kind, ok := index[kindName]; ok {
		return kind, nil
	}
	return catalogkind.CatalogKind_unspecified,
		errors.Errorf("no matching CatalogKind found for kind: %s", kindName)
}
