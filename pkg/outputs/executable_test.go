package outputs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestRunTransformExecutable_HappyPath(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "with-executable")
	raw := map[string]interface{}{
		"custom_id": "abc-123",
		"name":      "my-resource",
	}

	got, err := runTransformExecutable(dir, cloudresourcekind.CloudResourceKind_Auth0ResourceServer, raw)
	if err != nil {
		t.Fatal(err)
	}

	if got["id"] != "abc-123" {
		t.Errorf("expected id=abc-123, got %q", got["id"])
	}
	if got["name"] != "my-resource" {
		t.Errorf("expected name=my-resource, got %q", got["name"])
	}
	if _, exists := got["custom_id"]; exists {
		t.Error("custom_id should have been renamed to id")
	}
}

func TestRunTransformExecutable_NonZeroExit(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "bad-executable")
	raw := map[string]interface{}{"key": "value"}

	_, err := runTransformExecutable(dir, cloudresourcekind.CloudResourceKind_Auth0ResourceServer, raw)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}

func TestRunTransformExecutable_FileNotFound(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "empty")
	raw := map[string]interface{}{"key": "value"}

	_, err := runTransformExecutable(dir, cloudresourcekind.CloudResourceKind_Auth0ResourceServer, raw)
	if err == nil {
		t.Fatal("expected error for missing executable, got nil")
	}
}

func TestRunTransformExecutable_EmptyOutputs(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "with-executable")
	raw := map[string]interface{}{}

	got, err := runTransformExecutable(dir, cloudresourcekind.CloudResourceKind_Auth0ResourceServer, raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result for empty input, got %v", got)
	}
}

func TestRunTransformExecutable_MalformedOutputScript(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, executableFileName)
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'not-json'"), 0755); err != nil {
		t.Fatal(err)
	}

	raw := map[string]interface{}{"key": "value"}
	_, err := runTransformExecutable(dir, cloudresourcekind.CloudResourceKind_Auth0ResourceServer, raw)
	if err == nil {
		t.Fatal("expected error for malformed JSON output, got nil")
	}
}
