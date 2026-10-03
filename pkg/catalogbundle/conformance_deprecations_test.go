package catalogbundle

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"sigs.k8s.io/yaml"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// The deprecation half of the conformance gate, pinned violation by
// violation. Every case starts from a REAL bundle built over the real
// catalog tree, then breaks exactly one deprecation invariant -- the same
// crafted-mutation idiom as the loader's corruption pins.

// A stale bundle -- built before a deprecation was authored -- passes every
// schema check yet would make runtimes silently announce nothing. The
// agreement check refuses it by name.
func TestConformanceRefusesStaleDeprecations(t *testing.T) {
	bundle := loadBundleWithMutatedDeprecations(t, func(meta *catalogkind.CatalogKindMeta) {
		meta.Deprecations = nil
	})
	err := CheckConformance(bundle)
	if err == nil {
		t.Fatal("a bundle whose kind registry lacks a compiled-registry deprecation must be refused")
	}
	if !strings.Contains(err.Error(), "version deprecations differ") {
		t.Fatalf("expected the deprecation-agreement finding, got: %v", err)
	}
}

// A deprecation naming a version the bundle ships no schema for is a lie the
// bundle tells about itself -- refused with the missing schema named.
func TestConformanceRefusesDeprecationOfSchemalessVersion(t *testing.T) {
	bundle := loadBundleWithMutatedDeprecations(t, func(meta *catalogkind.CatalogKindMeta) {
		meta.Deprecations[0].Version = "v9alpha9"
	})
	err := CheckConformance(bundle)
	if err == nil {
		t.Fatal("a deprecation naming a schema-less version must be refused")
	}
	if !strings.Contains(err.Error(), "the bundle carries no schema for it") {
		t.Fatalf("expected the schema-existence finding, got: %v", err)
	}
}

// A deprecation whose version has no authored conversion path to the served
// version strands its writers -- the dead-end refusal.
func TestConformanceRefusesDeprecationWithoutConversionPath(t *testing.T) {
	dir := t.TempDir()
	descriptorsPath := filepath.Join(dir, "descriptors.binpb")
	writeLinkedDescriptorSet(t, descriptorsPath)

	bundlePath := filepath.Join(dir, "catalog-bundle.zip")
	if _, err := Build(BuildInput{
		DescriptorSetPath: descriptorsPath,
		CatalogDir:        catalogDir(t),
		OutputPath:        bundlePath,
	}); err != nil {
		t.Fatal(err)
	}

	strippedPath := filepath.Join(dir, "stripped.zip")
	rezipWithoutEntry(t, bundlePath, strippedPath,
		"conversions/_test/testcatalogkindgeneric/v1alpha1_to_v1alpha2.yaml")

	bundle, err := Load(strippedPath)
	if err != nil {
		t.Fatalf("a bundle with a consistently-removed entry still loads (conformance owns the verdict): %v", err)
	}
	err = CheckConformance(bundle)
	if err == nil {
		t.Fatal("a deprecation with no conversion path to the served version must be refused")
	}
	if !strings.Contains(err.Error(), "no conversion path to the served version") {
		t.Fatalf("expected the dead-end finding, got: %v", err)
	}
}

// loadBundleWithMutatedDeprecations builds a bundle whose descriptor set
// carries a mutated copy of the torture kind's deprecations -- the compiled
// registry stays truthful, the bundle lies.
func loadBundleWithMutatedDeprecations(t *testing.T, mutate func(*catalogkind.CatalogKindMeta)) *Bundle {
	t.Helper()
	dir := t.TempDir()
	descriptorsPath := filepath.Join(dir, "descriptors.binpb")
	writeLinkedDescriptorSet(t, descriptorsPath)

	raw, err := os.ReadFile(descriptorsPath)
	if err != nil {
		t.Fatal(err)
	}
	var fds descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(raw, &fds); err != nil {
		t.Fatal(err)
	}
	mutated := false
	for _, file := range fds.File {
		if file.GetName() != "shared/catalogkind/catalog_kind.proto" {
			continue
		}
		for _, enum := range file.EnumType {
			if enum.GetName() != "CatalogKind" {
				continue
			}
			for _, value := range enum.Value {
				if value.GetName() != "TestCatalogKindGeneric" {
					continue
				}
				meta, ok := proto.GetExtension(value.Options, catalogkind.E_KindMeta).(*catalogkind.CatalogKindMeta)
				if !ok || meta == nil || len(meta.GetDeprecations()) == 0 {
					t.Fatal("the torture kind's deprecation fixture is missing from the linked registry")
				}
				mutate(meta)
				proto.SetExtension(value.Options, catalogkind.E_KindMeta, meta)
				mutated = true
			}
		}
	}
	if !mutated {
		t.Fatal("the torture kind's registry entry was not found -- the mutation walk is broken")
	}
	remarshaled, err := proto.Marshal(&fds)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(descriptorsPath, remarshaled, 0o644); err != nil {
		t.Fatal(err)
	}

	bundlePath := filepath.Join(dir, "catalog-bundle.zip")
	if _, err := Build(BuildInput{
		DescriptorSetPath: descriptorsPath,
		CatalogDir:        catalogDir(t),
		OutputPath:        bundlePath,
	}); err != nil {
		t.Fatal(err)
	}
	bundle, err := Load(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

// rezipWithoutEntry writes a copy of the bundle with one entry removed and
// the manifest's checksum record removed with it -- an internally consistent
// zip (Load passes), so the CONFORMANCE verdict is what the test exercises.
func rezipWithoutEntry(t *testing.T, src, dst, dropEntry string) {
	t.Helper()
	reader, err := zip.OpenReader(src)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	for _, file := range reader.File {
		if file.Name == dropEntry {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if file.Name == "manifest.yaml" {
			var manifest Manifest
			if err := yaml.UnmarshalStrict(content, &manifest); err != nil {
				t.Fatal(err)
			}
			delete(manifest.Checksums, dropEntry)
			content, err = yaml.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
		}
		w, err := zw.Create(file.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}
