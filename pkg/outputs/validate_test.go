//go:build !codegen
// +build !codegen

package outputs

import (
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestValidateOverride_GenericNoSamples(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "empty")

	result, err := ValidateOverride(kind, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverrideType != OverrideNone {
		t.Errorf("expected OverrideNone, got %s", result.OverrideType)
	}
	if len(result.SchemaErrors) != 0 {
		t.Errorf("expected no schema errors, got %v", result.SchemaErrors)
	}
	if result.DryRun != nil {
		t.Error("expected nil DryRun without sample outputs")
	}
}

func TestValidateOverride_ValidMapping(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "with-mapping")

	result, err := ValidateOverride(kind, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverrideType != OverrideMapping {
		t.Errorf("expected OverrideMapping, got %s", result.OverrideType)
	}
	// "custom_endpoint" maps to "endpoint" which is not a field on
	// Auth0ResourceServerStackOutputs, so we expect a schema error.
	foundEndpointError := false
	for _, e := range result.SchemaErrors {
		if containsSubstring(e, "endpoint") {
			foundEndpointError = true
		}
	}
	if !foundEndpointError {
		t.Error("expected schema error for 'endpoint' target not being a proto field")
	}
}

func TestValidateOverride_BadMapping(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "bad-mapping")

	result, err := ValidateOverride(kind, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.SchemaErrors) == 0 {
		t.Error("expected schema errors for malformed YAML, got none")
	}
}

func TestValidateOverride_Executable(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "with-executable")

	result, err := ValidateOverride(kind, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.OverrideType != OverrideExecutable {
		t.Errorf("expected OverrideExecutable, got %s", result.OverrideType)
	}
	if len(result.SchemaErrors) != 0 {
		t.Errorf("expected no schema errors for valid executable, got %v", result.SchemaErrors)
	}
}

func TestValidateOverride_DryRunGeneric(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "empty")
	samples := map[string]interface{}{
		"id":   "rs-001",
		"name": "My API",
	}

	result, err := ValidateOverride(kind, dir, samples)
	if err != nil {
		t.Fatal(err)
	}
	if result.DryRun == nil {
		t.Fatal("expected DryRun result with sample outputs")
	}
	if result.DryRun.PopulatedCount < 2 {
		t.Errorf("expected at least 2 populated fields, got %d", result.DryRun.PopulatedCount)
	}
	if result.DryRun.TotalProtoFields == 0 {
		t.Error("expected non-zero TotalProtoFields")
	}
}

func TestValidateOverride_DryRunWithMapping(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "with-mapping")
	samples := map[string]interface{}{
		"custom_id":      "mapped-001",
		"custom_name":    "Mapped API",
		"internal_debug": "should-skip",
	}

	// Schema errors exist (endpoint not a field), but we still run dry-run
	// because sampleOutputs is provided and errors are about unmapped targets.
	result, err := ValidateOverride(kind, dir, samples)
	if err != nil {
		t.Fatal(err)
	}

	// Schema has errors, so DryRun should be nil (we skip dry-run on schema errors).
	if len(result.SchemaErrors) > 0 && result.DryRun != nil {
		t.Error("DryRun should be nil when there are schema errors")
	}
}

func TestValidateOverride_DryRunWithExecutable(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	dir := filepath.Join("testdata", "modules", "with-executable")
	samples := map[string]interface{}{
		"custom_id": "exec-001",
		"name":      "Exec API",
	}

	result, err := ValidateOverride(kind, dir, samples)
	if err != nil {
		t.Fatal(err)
	}
	if result.DryRun == nil {
		t.Fatal("expected DryRun result with executable override")
	}
	if result.DryRun.PopulatedCount < 2 {
		t.Errorf("expected at least 2 populated fields, got %d", result.DryRun.PopulatedCount)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringImpl(s, substr))
}

func containsSubstringImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
