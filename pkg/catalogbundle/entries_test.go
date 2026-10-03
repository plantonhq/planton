package catalogbundle

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
)

// The entries cargo end to end over the REAL tree and the REAL registry:
// every user-facing kind projects exactly one entry, the torture provider is
// withheld, and each entry's display fields agree with the kind's own
// catalog page -- read independently here, so the projection can never
// drift from its source without this failing.
func TestCatalogEntriesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	descriptorsPath := filepath.Join(dir, "descriptors.binpb")
	writeLinkedDescriptorSet(t, descriptorsPath)

	bundlePath := filepath.Join(dir, "catalog-bundle.zip")
	if _, err := Build(BuildInput{
		DescriptorSetPath: descriptorsPath,
		CatalogDir:        catalogDir(t),
		OutputPath:        bundlePath,
	}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	bundle, err := Load(bundlePath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	catalogEntries := bundle.CatalogEntries()
	byKind := map[string]CatalogEntry{}
	for _, entry := range catalogEntries {
		byKind[entry.Kind] = entry
	}

	userFacing := 0
	for _, kind := range catalogkindreflect.KindsList() {
		if catalogkindreflect.GetProvider(kind).String() == testProviderName {
			kindName := catalogkindreflect.ExtractKindNameByKind(kind)
			if _, present := byKind[kindName]; present {
				t.Errorf("the %s test kind must never get a catalog entry", kindName)
			}
			continue
		}
		userFacing++
	}
	if len(catalogEntries) != userFacing {
		t.Fatalf("expected one entry per user-facing kind (%d), got %d", userFacing, len(catalogEntries))
	}

	// One real kind pinned in depth, its truth read from the tree here
	// rather than hardcoded so catalog edits never break the bundle suite.
	entry, ok := byKind["AwsKmsKey"]
	if !ok {
		t.Fatal("AwsKmsKey has no catalog entry")
	}
	wantTitle, wantDescription := readCatalogPage(
		filepath.Join(catalogDir(t), "aws", "awskmskey", "catalog.md"), "AwsKmsKey")
	if entry.Title != wantTitle || entry.Description != wantDescription {
		t.Errorf("AwsKmsKey display fields diverge from its catalog page: got (%q, %q), want (%q, %q)",
			entry.Title, entry.Description, wantTitle, wantDescription)
	}
	if entry.Slug != "aws-kms-key" {
		t.Errorf("AwsKmsKey slug = %q, want aws-kms-key", entry.Slug)
	}
	if entry.LogoUrl != logoBaseURL+"/aws/awskmskey/logo.svg" {
		t.Errorf("AwsKmsKey logo url = %q", entry.LogoUrl)
	}
	if entry.IacModules.TerraformModuleDir != "catalog/aws/awskmskey/iac/tf" ||
		entry.IacModules.PulumiModuleDir != "catalog/aws/awskmskey/iac/pulumi" {
		t.Errorf("AwsKmsKey module dirs = %+v", entry.IacModules)
	}
	if !strings.Contains(entry.WebLinks.SourceCode.Spec, "catalog/aws/awskmskey/") ||
		!strings.HasSuffix(entry.WebLinks.SourceCode.Spec, "/spec.proto") {
		t.Errorf("AwsKmsKey spec link = %q", entry.WebLinks.SourceCode.Spec)
	}

	// Every entry must survive the conformance bar the release lane applies.
	if err := CheckConformance(bundle); err != nil {
		t.Fatalf("conformance failed: %v", err)
	}
}

// The slug rule is ONE rule for every kind -- the provider stays an atomic
// word, the rest kebabs. These pins hold the derivation still.
func TestEntrySlugDerivation(t *testing.T) {
	cases := []struct {
		kindName, providerDir, want string
	}{
		{"AwsS3Bucket", "aws", "aws-s3-bucket"},
		{"AwsKmsKey", "aws", "aws-kms-key"},
		{"AwsMemorydbCluster", "aws", "aws-memorydb-cluster"},
		{"KubernetesDeployment", "kubernetes", "kubernetes-deployment"},
		{"DigitalOceanDroplet", "digitalocean", "digitalocean-droplet"},
		{"CloudflareDnsZone", "cloudflare", "cloudflare-dns-zone"},
		// A kind not prefixed by its provider kebabs whole.
		{"GkeCluster", "gcp", "gke-cluster"},
	}
	for _, c := range cases {
		if got := entrySlug(c.kindName, c.providerDir); got != c.want {
			t.Errorf("entrySlug(%s, %s) = %q, want %q", c.kindName, c.providerDir, got, c.want)
		}
	}
}

// Title comes from the catalog page's H1; the description is the intro
// paragraph's first sentence; a kind without a page falls back to its
// kind name.
func TestReadCatalogPage(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "catalog.md")
	content := "# AWS Widget Fleet\n\nDeploys a widget fleet with sensible defaults. Everything after the first sentence is detail.\n"
	if err := os.WriteFile(page, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	title, description := readCatalogPage(page, "AwsWidgetFleet")
	if title != "AWS Widget Fleet" {
		t.Errorf("title = %q", title)
	}
	if description != "Deploys a widget fleet with sensible defaults." {
		t.Errorf("description = %q", description)
	}

	title, description = readCatalogPage(filepath.Join(dir, "absent.md"), "AwsWidgetFleet")
	if title != "AwsWidgetFleet" || description != "Deploy AwsWidgetFleet" {
		t.Errorf("fallbacks = (%q, %q)", title, description)
	}
}

// A tree that cannot satisfy the registry fails the BUILD, naming every
// missing kind -- deploy coordinates are proven at release build time.
func TestProjectEntriesRefusesForeignTree(t *testing.T) {
	if _, err := projectEntries(t.TempDir(), nil); err == nil {
		t.Fatal("a catalog tree missing the registry's kinds must fail entry projection")
	} else if !strings.Contains(err.Error(), "no kind directory") {
		t.Fatalf("refusal must name the missing kinds, got: %v", err)
	}
}

// A malformed or mis-keyed entry refuses the LOAD whole, exactly like a
// checksum failure -- consumers never see a partially valid catalog.
func TestLoadRefusesBrokenEntries(t *testing.T) {
	descriptors := linkedDescriptorSetBytes(t)

	malformed := map[string][]byte{
		"descriptors.binpb":       descriptors,
		"entries/aws/awsvpc.yaml": []byte("kind: AwsVpc\nnot_a_field: true\n"),
	}
	if _, err := Load(writeRawBundle(t, malformed)); err == nil {
		t.Fatal("an entry with unknown fields must refuse the load")
	}

	missingIdentity := map[string][]byte{
		"descriptors.binpb":       descriptors,
		"entries/aws/awsvpc.yaml": []byte("kind: AwsVpc\ntitle: AWS VPC\n"),
	}
	if _, err := Load(writeRawBundle(t, missingIdentity)); err == nil {
		t.Fatal("an entry without its full identity must refuse the load")
	}

	misKeyed := map[string][]byte{
		"descriptors.binpb": descriptors,
		"entries/aws/awsvpc.yaml": mustMarshalEntry(t, CatalogEntry{
			Kind: "AwsSubnet", Title: "AWS Subnet", Description: "d", Slug: "aws-subnet",
		}),
	}
	if _, err := Load(writeRawBundle(t, misKeyed)); err == nil {
		t.Fatal("an entry keyed by a different kind than it declares must refuse the load")
	}
}

// A bundle built before entries existed still LOADS (its checksums hold) but
// must fail conformance loudly, naming the kinds that have no entry -- the
// agreement gate is what makes a stale artifact unusable, not a version dance.
func TestConformanceRefusesEntrylessBundle(t *testing.T) {
	bundle, err := Load(writeRawBundle(t, map[string][]byte{
		"descriptors.binpb": linkedDescriptorSetBytes(t),
	}))
	if err != nil {
		t.Fatalf("an entryless bundle must still load (structure is intact): %v", err)
	}
	err = CheckConformance(bundle)
	if err == nil {
		t.Fatal("an entryless bundle must fail conformance")
	}
	if !strings.Contains(err.Error(), "no catalog entry") {
		t.Fatalf("the refusal must name the missing entries, got: %v", err)
	}
}

// Stray entries (kinds the registry does not serve) and duplicate slugs are
// conformance findings with the offenders named.
func TestConformanceRefusesStrayAndDuplicateEntries(t *testing.T) {
	stray := mustMarshalEntry(t, CatalogEntry{
		Kind: "AwsImaginaryThing", Title: "Imaginary", Description: "d", Slug: "shared-slug",
		IacModules: CatalogEntryIacModules{TerraformModuleDir: "catalog/aws/awsimaginarything/iac/tf"},
	})
	strayTwin := mustMarshalEntry(t, CatalogEntry{
		Kind: "AwsImaginaryTwin", Title: "Imaginary Twin", Description: "d", Slug: "shared-slug",
		IacModules: CatalogEntryIacModules{TerraformModuleDir: "catalog/aws/awsimaginarytwin/iac/tf"},
	})
	bundle, err := Load(writeRawBundle(t, map[string][]byte{
		"descriptors.binpb":                  linkedDescriptorSetBytes(t),
		"entries/aws/awsimaginarything.yaml": stray,
		"entries/aws/awsimaginarytwin.yaml":  strayTwin,
	}))
	if err != nil {
		t.Fatal(err)
	}
	err = CheckConformance(bundle)
	if err == nil {
		t.Fatal("stray and duplicate-slug entries must fail conformance")
	}
	if !strings.Contains(err.Error(), "does not serve this kind") {
		t.Errorf("the refusal must name the stray entry, got: %v", err)
	}
	if !strings.Contains(err.Error(), "share the slug") {
		t.Errorf("the refusal must name the slug collision, got: %v", err)
	}
}

// A tampered entry document is caught by the same self-verification that
// guards every other cargo class.
func TestLoadRefusesTamperedEntry(t *testing.T) {
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

	corruptedPath := filepath.Join(dir, "corrupted.zip")
	rezipWithAlteredEntry(t, bundlePath, corruptedPath, "entries/aws/awskmskey.yaml")
	if _, err := Load(corruptedPath); err == nil {
		t.Fatal("a bundle whose entry does not match its recorded checksum must be refused")
	}
}

// writeRawBundle assembles a zip whose manifest checksums are CONSISTENT
// with the given contents -- structurally sound by construction, so tests
// exercise the semantic gates rather than the checksum gate.
func writeRawBundle(t *testing.T, contents map[string][]byte) string {
	t.Helper()
	manifest := &Manifest{FormatVersion: FormatVersion, Checksums: map[string]string{}}
	for name, content := range contents {
		sum := sha256.Sum256(content)
		manifest.Checksums[name] = hex.EncodeToString(sum[:])
	}
	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "bundle.zip")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	writeEntry := func(name string, content []byte) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	writeEntry("manifest.yaml", manifestYAML)
	for name, content := range contents {
		writeEntry(name, content)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func linkedDescriptorSetBytes(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "descriptors.binpb")
	writeLinkedDescriptorSet(t, path)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustMarshalEntry(t *testing.T, entry CatalogEntry) []byte {
	t.Helper()
	raw, err := yaml.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
