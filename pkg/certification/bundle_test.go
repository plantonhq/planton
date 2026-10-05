package certification

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/plantonhq/planton/pkg/catalogbundle"
	"github.com/plantonhq/planton/shared/catalogkind"

	// This test assembles its descriptor set from the LINKED registry, and
	// only the served version reaches a binary through the kind registry --
	// the conversion-source version must be linked explicitly. (The real
	// bundle is buf-built from the proto tree and carries every version
	// regardless of linkage.)
	_ "github.com/plantonhq/planton/catalog/_test/testcatalogkindgeneric/v1alpha1"
)

// The catalog-as-data certification case: a built bundle must carry BOTH of
// the torture kind's versions -- the old one is the conversion source, and a
// runtime serving documents stored under it needs its schema from the same
// bundle. This is the groundwork multi-version serving stands on.
func TestCertify_BundleCarriesBothTortureVersions(t *testing.T) {
	bundle := buildRealBundle(t)

	for _, version := range []string{"v1alpha1", "v1alpha2"} {
		name := protoreflect.FullName(
			"dev.planton._test.testcatalogkindgeneric." + version + ".TestCatalogKindGeneric")
		if _, err := bundle.Files.FindDescriptorByName(name); err != nil {
			t.Errorf("the bundle must carry %s (the %s schema) -- stored documents at that version cannot be served without it", name, version)
		}
	}

	if _, ok := bundle.ConversionSpecs()["conversions/_test/testcatalogkindgeneric/v1alpha1_to_v1alpha2.yaml"]; !ok {
		t.Error("the bundle must carry the bridge between the versions it serves")
	}
}

// The deprecation-lifecycle certification case: the built bundle ANNOUNCES
// the torture kind's v1alpha1 as deprecated -- readable from the bundle's
// own kind registry, exactly where runtimes read it -- and still passes
// conformance, proving a deprecation can only ship with its schema aboard
// and its conversion path authored. Removal-after-deprecation is the later
// half of this lifecycle; this case is its standing seed.
func TestCertify_BundleAnnouncesTortureDeprecation(t *testing.T) {
	bundle := buildRealBundle(t)

	desc, err := bundle.Files.FindDescriptorByName("dev.planton.shared.catalogkind.CatalogKind")
	if err != nil {
		t.Fatalf("the bundle carries no kind registry enum: %v", err)
	}
	enum, ok := desc.(protoreflect.EnumDescriptor)
	if !ok {
		t.Fatal("the bundle's kind registry is not an enum")
	}
	value := enum.Values().ByName("TestCatalogKindGeneric")
	if value == nil {
		t.Fatal("the bundle's kind registry does not name the torture kind")
	}
	meta, ok := proto.GetExtension(value.Options(), catalogkind.E_KindMeta).(*catalogkind.CatalogKindMeta)
	if !ok || meta == nil {
		t.Fatal("the torture kind's registry entry carries no kind_meta")
	}
	deprecations := meta.GetDeprecations()
	if len(deprecations) != 1 {
		t.Fatalf("expected exactly one announced deprecation (the permanent fixture), got %d", len(deprecations))
	}
	if got := deprecations[0].GetVersion(); got != "v1alpha1" {
		t.Errorf("expected v1alpha1 announced deprecated, got %q", got)
	}
	if deprecations[0].GetNote() == "" {
		t.Error("the fixture's note must ride the bundle so note passthrough stays exercised end to end")
	}
}

// buildRealBundle assembles a descriptor set from the linked registry, builds
// a bundle over the real catalog tree, loads it, and proves conformance --
// the shared starting point every certification case trusts.
func buildRealBundle(t *testing.T) *catalogbundle.Bundle {
	t.Helper()
	dir := t.TempDir()

	var fds descriptorpb.FileDescriptorSet
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fd))
		return true
	})
	raw, err := proto.Marshal(&fds)
	if err != nil {
		t.Fatal(err)
	}
	descriptorsPath := filepath.Join(dir, "descriptors.binpb")
	if err := os.WriteFile(descriptorsPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	bundlePath := filepath.Join(dir, "catalog-bundle.zip")
	if _, err := catalogbundle.Build(catalogbundle.BuildInput{
		DescriptorSetPath: descriptorsPath,
		CatalogDir:        filepath.Join(TortureKindRoot(t), "..", ".."),
		OutputPath:        bundlePath,
	}); err != nil {
		t.Fatal(err)
	}
	bundle, err := catalogbundle.Load(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalogbundle.CheckConformance(bundle); err != nil {
		t.Fatal(err)
	}
	return bundle
}
