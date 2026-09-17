//go:build !codegen
// +build !codegen

package providerparity

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// TestScanModule_HermeticFixture proves the scan against a module that would
// defeat a main.tf-only reader: resources split across sibling files, a
// duplicate declaration de-duplicated, pins in provider.tf, a resource
// attached through the secondary provider by meta-argument, and non-.tf
// noise ignored.
func TestScanModule_HermeticFixture(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"provider.tf": `terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 6.0"
    }
  }
}

provider "google" {
}
`,
		"main.tf": `resource "google_storage_bucket" "this" {
  name = "b"
}
`,
		"iam.tf": `resource "google_storage_bucket_iam_member" "reader" {
  member = "user:a@b.c"
}

resource "google_storage_bucket" "again" {
  name = "duplicate-type-must-not-double-count"
}
`,
		"beta.tf": `resource "google_firebase_project" "this" {
  provider = google-beta
  project  = "p"
}
`,
		"README.md": `resource "not_terraform" "docs" -- prose, must be ignored`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	census, err := ScanModule(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	wantResources := []string{"google_firebase_project", "google_storage_bucket", "google_storage_bucket_iam_member"}
	if !reflect.DeepEqual(census.Resources, wantResources) {
		t.Errorf("resources = %v, want %v", census.Resources, wantResources)
	}
	wantPins := map[string]string{"google": "~> 6.0", "google-beta": "~> 6.0"}
	if !reflect.DeepEqual(census.Pins, wantPins) {
		t.Errorf("pins = %v, want %v", census.Pins, wantPins)
	}
	// Only the explicit meta-argument is recorded: the buckets ride the
	// implied provider and have no entry.
	wantAttachments := map[string]string{"google_firebase_project": "google-beta"}
	if !reflect.DeepEqual(census.ProviderAttachments, wantAttachments) {
		t.Errorf("attachments = %v, want %v", census.ProviderAttachments, wantAttachments)
	}
}

// TestScanModule_SplitAttachmentIsAnError: one resource type attached
// through two providers has no single truth to account against.
func TestScanModule_SplitAttachmentIsAnError(t *testing.T) {
	dir := t.TempDir()
	content := `resource "google_firebase_project" "a" {
  provider = google-beta
}

resource "google_firebase_project" "b" {
  provider = google
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScanModule(dir); err == nil {
		t.Fatal("expected a split-attachment error, got nil")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestModuleCensusGcp is the live-catalog smoke test: every implemented GCP
// kind has a scannable module that declares at least one resource and at
// least one provider pin. GCP is anatomy-clean, so a missing module here is
// a hard failure, not a finding to baseline.
func TestModuleCensusGcp(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	census, err := ModuleCensusForProvider(root, cloudresourcekind.CloudResourceProvider_gcp)
	if err != nil {
		t.Fatalf("census: %v", err)
	}
	if len(census) == 0 {
		t.Fatal("GCP module census is empty -- the registry walk is broken")
	}
	for _, m := range census {
		if m.MissingModule {
			t.Errorf("%s: no Terraform module directory -- GCP holds zero anatomy debt", m.Kind)
			continue
		}
		if len(m.Resources) == 0 {
			t.Errorf("%s: module declares no resources", m.Kind)
		}
		if len(m.Pins) == 0 {
			t.Errorf("%s: module declares no provider pins", m.Kind)
		}
	}
}

// TestModuleCensusIsProviderAgnostic is the guard the sibling cloud catalogs
// rely on: the census must RUN over every major provider's catalog -- kinds
// with recorded anatomy debt (a missing iac/tf) surface as MissingModule
// census rows, never as an error that hides the rest of the catalog.
func TestModuleCensusIsProviderAgnostic(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	for _, provider := range []cloudresourcekind.CloudResourceProvider{
		cloudresourcekind.CloudResourceProvider_azure,
		cloudresourcekind.CloudResourceProvider_aws,
	} {
		census, err := ModuleCensusForProvider(root, provider)
		if err != nil {
			t.Errorf("%s: census must run over anatomy-baselined trees, got: %v", provider, err)
			continue
		}
		if len(census) == 0 {
			t.Errorf("%s: census is empty -- the registry walk is broken", provider)
			continue
		}
		missing := 0
		for _, m := range census {
			if m.MissingModule {
				missing++
			}
		}
		spec := SpecCensus(provider)
		if len(spec) == 0 {
			t.Errorf("%s: spec census is empty -- the descriptor walk is broken", provider)
		}
		t.Logf("%s: %d kinds censused (%d spec-censused), %d missing modules (anatomy-baselined debt)",
			provider, len(census), len(spec), missing)
	}
}
