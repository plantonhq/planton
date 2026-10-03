package protodocs

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	// Registers every catalog kind's proto files in the global
	// registry so the coverage check below sees the full compiled-in surface.
	_ "github.com/plantonhq/planton/pkg/catalogkindreflect"
)

// TestIndexCoversCompiledInProtos is the freshness gate for the committed
// index: every proto file of this repo's module compiled into this binary
// must be present in the embedded artifact. Adding a kind (or any proto)
// without running `make generate-proto-docs` fails here instead of silently
// rendering an undocumented schema.
func TestIndexCoversCompiledInProtos(t *testing.T) {
	indexed := make(map[string]bool, len(Files()))
	for _, f := range Files() {
		indexed[f] = true
	}

	// The module's proto roots; third-party descriptors in the registry
	// (google/protobuf, buf/validate) are deliberately outside the index.
	moduleRoots := []string{"catalog/", "shared/", "qa/", "iac/"}

	var missing []string
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		path := fd.Path()
		for _, root := range moduleRoots {
			if strings.HasPrefix(path, root) && !indexed[path] {
				missing = append(missing, path)
				break
			}
		}
		return true
	})
	if len(missing) > 0 {
		t.Fatalf("embedded docs index is stale -- run `make generate-proto-docs`; missing %d files, first: %s",
			len(missing), missing[0])
	}
}

// TestLookupKeysMatchRuntimeReflection proves the distiller's name resolution
// produces exactly the fully-qualified names protoreflect reports at runtime,
// on a densely documented spec: message, plain field, and nested enum value
// (enum values scope to the enum's PARENT per proto scoping rules).
func TestLookupKeysMatchRuntimeReflection(t *testing.T) {
	const specName = "dev.planton.kubernetes.kubernetesnamespace.v1alpha1.KubernetesNamespaceSpec"
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(specName))
	if err != nil {
		t.Fatalf("spec message not registered: %v", err)
	}
	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		t.Fatalf("%s is not a message descriptor", specName)
	}

	if doc := Lookup(md.FullName()); !strings.Contains(doc, "Namespace-as-a-Service") {
		t.Errorf("message doc mismatch for %s: %q", specName, doc)
	}
	// The catalog's house style opens docs with a bold name. The gutter
	// stripper must not eat the first bold marker ("**X**" is markdown,
	// not a block-comment gutter).
	if doc := Lookup(md.FullName()); !strings.HasPrefix(doc, "**KubernetesNamespaceSpec**") {
		t.Errorf("bold marker mangled by comment cleaning: %q", doc[:40])
	}

	nameField := md.Fields().ByName("name")
	if nameField == nil {
		t.Fatal("field `name` not found on the spec")
	}
	if doc := Lookup(nameField.FullName()); !strings.Contains(doc, "DNS label") {
		t.Errorf("field doc mismatch for %s: %q", nameField.FullName(), doc)
	}

	enum := md.Enums().ByName("KubernetesNamespacePodSecurityStandard")
	if enum == nil {
		t.Fatal("nested enum not found on the spec")
	}
	baseline := enum.Values().ByName("baseline")
	if baseline == nil {
		t.Fatal("enum value `baseline` not found")
	}
	if doc := Lookup(baseline.FullName()); !strings.Contains(doc, "Minimally restrictive") {
		t.Errorf("enum value doc mismatch for %s: %q", baseline.FullName(), doc)
	}
}

// TestDocDensity guards against a silently hollow index (e.g. a distiller
// regression that drops comment classes): the documented-element count must
// stay in the tens of thousands for a 400+ kind catalog.
func TestDocDensity(t *testing.T) {
	load()
	if got := len(loaded.Docs); got < 10000 {
		t.Fatalf("suspiciously sparse docs index: %d documented elements", got)
	}
}

// TestEmbedSizeBudget keeps the artifact a monitored number: the CLI binary
// is ~110MB, and the docs embed must stay a rounding error on it. Growth past
// this generous budget deserves a deliberate decision (shard per provider or
// move to a sidecar), never an accidental drift.
func TestEmbedSizeBudget(t *testing.T) {
	const budget = 5 << 20 // 5MB compressed
	if len(indexGz) > budget {
		t.Fatalf("embedded docs index is %d bytes, over the %d budget -- decide sharding/sidecar deliberately", len(indexGz), budget)
	}
}
