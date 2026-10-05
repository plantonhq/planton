package conversion

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"

	// The corpus converts between BOTH torture-kind versions; the old
	// version's package is linked explicitly so its descriptors resolve (only
	// the served version reaches the binary through the kind registry). A
	// future kind's first graduation must add its old version's package here
	// the same way, or the descriptor lookup below fails loudly.
	_ "github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/v1alpha1"
)

// The golden corpus: every conversion spec in the catalog must carry fixture
// pairs (conversions/testdata/<case>/{input,expected}.yaml), and every
// fixture must satisfy two laws, both compared in the CANONICAL DOMAIN (each
// side canonicalized at its version — see Canonicalize):
//
//  1. canon(upgrade(canon(input))) == canon(expected).
//  2. canon(downgrade(upgrade(canon(input)))) == canon(input) minus the
//     upgrade's DECLARED losses — the round-trip law: loss is legal only
//     when the spec says so.
//
// The canonical domain is what makes storage-spelled fixtures first-class:
// a case authored in the persist print's shape (proto-name keys, 64-bit
// integers as strings) proves the exact documents production lanes feed the
// engine after canonicalizing at entry. Representation differences protobuf
// declares meaningless (key spelling, 64-bit string form, non-presence
// zeros) dissolve in the comparison; REAL differences still fail exactly.
//
// Both the Go engine here and the platform's Java engine gate on this same
// corpus; an engine disagreement is a CI failure, not a runtime surprise.

func catalogDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "catalog")
}

func loadDoc(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return doc
}

// messageDescriptor resolves the kind's message descriptor at the given
// version. The kind registry serves only the served version's message; other
// versions' packages are linked by this test's explicit imports and located
// by swapping the version segment of the served message's full name — the
// package-path convention the registry gates enforce.
func messageDescriptor(t *testing.T, kindName, version string) protoreflect.MessageDescriptor {
	t.Helper()
	kind := catalogkindreflect.KindFromString(kindName)
	served, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("kind %q is not in the registry: %v", kindName, err)
	}
	servedVersion, err := catalogkindreflect.KindVersion(kind)
	if err != nil {
		t.Fatalf("kind %q has no version: %v", kindName, err)
	}
	full := string(served.ProtoReflect().Descriptor().FullName())
	swapped := strings.Replace(full, "."+servedVersion+".", "."+version+".", 1)
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(swapped))
	if err != nil {
		t.Fatalf("no descriptor %s -- link the %s package into this test's imports: %v",
			swapped, version, err)
	}
	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		t.Fatalf("%s is not a message descriptor", swapped)
	}
	return md
}

func canon(t *testing.T, md protoreflect.MessageDescriptor, doc map[string]any, label string) map[string]any {
	t.Helper()
	canonical, err := Canonicalize(md, doc)
	if err != nil {
		t.Fatalf("canonicalizing %s: %v", label, err)
	}
	return canonical
}

func TestGoldenCorpus(t *testing.T) {
	base := catalogDir(t)
	fsys := os.DirFS(base)
	specFiles, err := SpecFiles(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(specFiles) == 0 {
		t.Fatal("no conversion specs found -- the corpus gate is broken")
	}

	for _, specFile := range specFiles {
		spec, err := LoadSpec(filepath.Join(base, specFile))
		if err != nil {
			t.Errorf("%s: %v", specFile, err)
			continue
		}

		fixturesDir := filepath.Join(base, filepath.Dir(specFile), "testdata")
		entries, err := os.ReadDir(fixturesDir)
		if err != nil || len(entries) == 0 {
			t.Errorf("%s has no golden fixtures -- every conversion spec ships its corpus (conversions/testdata/<case>/{input,expected}.yaml)", specFile)
			continue
		}

		fromMd := messageDescriptor(t, spec.Kind, spec.From)
		toMd := messageDescriptor(t, spec.Kind, spec.To)

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			caseName := entry.Name()
			t.Run(spec.Kind+"/"+spec.From+"_to_"+spec.To+"/"+caseName, func(t *testing.T) {
				input := loadDoc(t, filepath.Join(fixturesDir, caseName, "input.yaml"))
				expected := loadDoc(t, filepath.Join(fixturesDir, caseName, "expected.yaml"))

				// The engine converts the canonical form -- exactly what every
				// production lane feeds it after canonicalizing at entry.
				canonIn := canon(t, fromMd, input, "input")

				upgraded, losses, err := Apply(spec, Upgrade, canonIn)
				if err != nil {
					t.Fatalf("upgrade failed: %v", err)
				}
				if !reflect.DeepEqual(canon(t, toMd, upgraded, "upgraded"),
					canon(t, toMd, expected, "expected")) {
					t.Errorf("upgrade(input) != expected (canonical domain)\ngot:      %#v\nexpected: %#v",
						upgraded, expected)
				}

				// Round-trip law: the downgrade restores the canonical input
				// except for losses the UPGRADE declared.
				roundTripped, _, err := Apply(spec, Downgrade, upgraded)
				if err != nil {
					t.Fatalf("downgrade failed: %v", err)
				}
				pruned := deepCopy(canonIn).(map[string]any)
				for _, loss := range losses {
					deletePath(pruned, loss.Path)
				}
				if !reflect.DeepEqual(canon(t, fromMd, roundTripped, "round-tripped"),
					canon(t, fromMd, pruned, "pruned input")) {
					t.Errorf("round-trip broke: downgrade(upgrade(input)) != input minus declared losses (canonical domain)\ngot:      %#v\nexpected: %#v",
						roundTripped, pruned)
				}
			})
		}
	}
}
