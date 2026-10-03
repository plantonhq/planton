package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// The no-new-version-without-a-bridge gate.
//
// A kind directory holding MORE than one api-version directory has committed
// to serving a version pair: from that moment a conversion path between the
// served version and every other version must exist, and each bridge must be
// TOTAL (CheckTotality -- a field must never just vanish). Requiredness of
// added fields is proven by the corpus, which runs full schema validation on
// every converted fixture.

var versionDirRe = regexp.MustCompile(`^v\d+((alpha|beta)\d+)?$`)

func TestEveryVersionPairHasATotalBridge(t *testing.T) {
	base := catalogDir(t)
	fsys := os.DirFS(base)

	providers, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	checkedBridges := 0
	for _, provider := range providers {
		if !provider.IsDir() {
			continue
		}
		kinds, err := os.ReadDir(filepath.Join(base, provider.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, kindDir := range kinds {
			if !kindDir.IsDir() {
				continue
			}
			var versions []string
			entries, err := os.ReadDir(filepath.Join(base, provider.Name(), kindDir.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				if e.IsDir() && versionDirRe.MatchString(e.Name()) {
					versions = append(versions, e.Name())
				}
			}
			if len(versions) < 2 {
				continue
			}

			kind := catalogkindreflect.KindFromString(kindDir.Name())
			if kind == catalogkind.CatalogKind_unspecified {
				t.Errorf("%s/%s has version directories but is not a registered kind", provider.Name(), kindDir.Name())
				continue
			}
			specs, err := SpecsForKind(fsys, kind)
			if err != nil {
				t.Errorf("%s/%s: %v", provider.Name(), kindDir.Name(), err)
				continue
			}
			served, err := catalogkindreflect.KindVersion(kind)
			if err != nil {
				t.Errorf("%s/%s: %v", provider.Name(), kindDir.Name(), err)
				continue
			}

			for _, version := range versions {
				if version == served {
					continue
				}
				steps, err := Path(specs, version, served)
				if err != nil {
					t.Errorf("%s/%s serves %s but still carries %s: %v", provider.Name(), kindDir.Name(), served, version, err)
					continue
				}
				for _, step := range steps {
					fromMessage := specMessageName(provider.Name(), kindDir.Name(), step.Spec.From, step.Spec.Kind)
					toMessage := specMessageName(provider.Name(), kindDir.Name(), step.Spec.To, step.Spec.Kind)
					if err := CheckTotality(step.Spec, fromMessage, toMessage); err != nil {
						t.Error(err)
					}
					checkedBridges++
				}
			}
		}
	}
	if checkedBridges == 0 {
		t.Fatal("no version bridges were checked -- the torture kind's pair must exist; the gate walk is broken")
	}
}

func specMessageName(provider, kindDir, version, kindName string) string {
	return fmt.Sprintf("dev.planton.%s.%s.%s.%sSpec", provider, kindDir, version, kindName)
}
