//go:build !codegen
// +build !codegen

package outputs

import (
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestTransformRaw_GenericFallback(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{
		"id":   "rs-001",
		"name": "My API",
	}

	msg, flat, err := TransformRaw(kind, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil proto message")
	}
	if flat["id"] != "rs-001" {
		t.Errorf("flat[id] = %q, want rs-001", flat["id"])
	}
	if flat["name"] != "My API" {
		t.Errorf("flat[name] = %q, want 'My API'", flat["name"])
	}
}

func TestTransformRaw_NilOpts(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{"id": "test"}

	msg, _, err := TransformRaw(kind, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message with nil opts")
	}
}

func TestTransformRaw_EmptyModuleDir(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{"id": "test"}

	msg, _, err := TransformRaw(kind, raw, &TransformOptions{ModuleDir: ""})
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message with empty moduleDir")
	}
}

func TestTransformRaw_MappingOverride(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{
		"custom_id":      "rs-mapped",
		"custom_name":    "Mapped API",
		"internal_debug": "should-be-skipped",
	}

	opts := &TransformOptions{
		ModuleDir: filepath.Join("testdata", "modules", "with-mapping"),
	}

	msg, flat, err := TransformRaw(kind, raw, opts)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil proto message")
	}

	if flat["id"] != "rs-mapped" {
		t.Errorf("flat[id] = %q, want rs-mapped (renamed from custom_id)", flat["id"])
	}
	if flat["name"] != "Mapped API" {
		t.Errorf("flat[name] = %q, want 'Mapped API' (renamed from custom_name)", flat["name"])
	}
	if _, exists := flat["internal_debug"]; exists {
		t.Error("internal_debug should have been skipped by mapping")
	}
	if _, exists := flat["custom_id"]; exists {
		t.Error("custom_id should have been renamed to id")
	}
}

func TestTransformRaw_ExecutableOverride(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{
		"custom_id": "rs-exec",
		"name":      "Exec API",
	}

	opts := &TransformOptions{
		ModuleDir: filepath.Join("testdata", "modules", "with-executable"),
	}

	msg, flat, err := TransformRaw(kind, raw, opts)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil proto message")
	}

	if flat["id"] != "rs-exec" {
		t.Errorf("flat[id] = %q, want rs-exec (renamed by executable)", flat["id"])
	}
	if flat["name"] != "Exec API" {
		t.Errorf("flat[name] = %q, want 'Exec API'", flat["name"])
	}
}

func TestTransformRaw_ExecutableTakesPrecedence(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{
		"executable_was_used": "proof",
	}

	opts := &TransformOptions{
		ModuleDir: filepath.Join("testdata", "modules", "with-both"),
	}

	_, flat, err := TransformRaw(kind, raw, opts)
	if err != nil {
		t.Fatal(err)
	}

	// The executable renames "executable_was_used" -> "id".
	// The YAML would rename "yaml_was_used" -> "id" instead.
	if flat["id"] != "proof" {
		t.Errorf("flat[id] = %q, want 'proof' (executable should take precedence over YAML)", flat["id"])
	}
}

func TestTransformRaw_EmptyOutputs(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{}

	msg, flat, err := TransformRaw(kind, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message for empty outputs")
	}
	if len(flat) != 0 {
		t.Errorf("expected empty flat map, got %v", flat)
	}
}

func TestTransformRaw_BadExecutable(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{"key": "value"}

	opts := &TransformOptions{
		ModuleDir: filepath.Join("testdata", "modules", "bad-executable"),
	}

	_, _, err := TransformRaw(kind, raw, opts)
	if err == nil {
		t.Fatal("expected error for failing executable, got nil")
	}
}

func TestTransformRaw_BadMapping(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	raw := map[string]interface{}{"key": "value"}

	opts := &TransformOptions{
		ModuleDir: filepath.Join("testdata", "modules", "bad-mapping"),
	}

	_, _, err := TransformRaw(kind, raw, opts)
	if err == nil {
		t.Fatal("expected error for malformed YAML mapping, got nil")
	}
}
