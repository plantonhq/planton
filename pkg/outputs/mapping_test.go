package outputs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMapping_Valid(t *testing.T) {
	m, err := loadMapping(filepath.Join("testdata", "modules", "with-mapping"))
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != "v1" {
		t.Errorf("version = %q, want v1", m.Version)
	}
	if len(m.Mappings) != 3 {
		t.Errorf("mappings count = %d, want 3", len(m.Mappings))
	}
	if m.Mappings["custom_id"] != "id" {
		t.Errorf("mapping custom_id = %q, want id", m.Mappings["custom_id"])
	}
	if len(m.Skip) != 2 {
		t.Errorf("skip count = %d, want 2", len(m.Skip))
	}
}

func TestLoadMapping_MalformedYAML(t *testing.T) {
	_, err := loadMapping(filepath.Join("testdata", "modules", "bad-mapping"))
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoadMapping_FileNotFound(t *testing.T) {
	_, err := loadMapping(filepath.Join("testdata", "modules", "empty"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMapping_UnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, mappingFileName), []byte("version: v99\nmappings:\n  a: b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := loadMapping(dir)
	if err == nil {
		t.Fatal("expected error for unsupported version, got nil")
	}
}

func TestApplyMapping_RenameKeys(t *testing.T) {
	outputs := map[string]string{
		"custom_id":   "abc-123",
		"custom_name": "my-resource",
		"passthrough": "value",
	}
	m := &OutputMapping{
		Mappings: map[string]string{
			"custom_id":   "id",
			"custom_name": "name",
		},
	}

	got := applyMapping(outputs, m)
	want := map[string]string{
		"id":          "abc-123",
		"name":        "my-resource",
		"passthrough": "value",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("applyMapping() = %v, want %v", got, want)
	}
}

func TestApplyMapping_SkipKeys(t *testing.T) {
	outputs := map[string]string{
		"id":             "abc-123",
		"internal_debug": "debug-info",
		"temp_value":     "temp",
	}
	m := &OutputMapping{
		Skip: []string{"internal_debug", "temp_value"},
	}

	got := applyMapping(outputs, m)
	want := map[string]string{
		"id": "abc-123",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("applyMapping() = %v, want %v", got, want)
	}
}

func TestApplyMapping_RenameAndSkip(t *testing.T) {
	outputs := map[string]string{
		"custom_id":      "abc",
		"internal_debug": "debug",
		"other":          "val",
	}
	m := &OutputMapping{
		Mappings: map[string]string{"custom_id": "id"},
		Skip:     []string{"internal_debug"},
	}

	got := applyMapping(outputs, m)
	want := map[string]string{
		"id":    "abc",
		"other": "val",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("applyMapping() = %v, want %v", got, want)
	}
}

func TestApplyMapping_DotPathSourceKey(t *testing.T) {
	outputs := map[string]string{
		"connection.host": "db.example.com",
		"connection.port": "5432",
	}
	m := &OutputMapping{
		Mappings: map[string]string{
			"connection.host": "hostname",
			"connection.port": "port",
		},
	}

	got := applyMapping(outputs, m)
	want := map[string]string{
		"hostname": "db.example.com",
		"port":     "5432",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("applyMapping() = %v, want %v", got, want)
	}
}

func TestApplyMapping_EmptyMapping(t *testing.T) {
	outputs := map[string]string{"id": "abc"}
	m := &OutputMapping{}

	got := applyMapping(outputs, m)
	if !reflect.DeepEqual(got, outputs) {
		t.Errorf("empty mapping should pass through all keys; got %v, want %v", got, outputs)
	}
}

func TestApplyMapping_NilMapping(t *testing.T) {
	outputs := map[string]string{"id": "abc"}
	got := applyMapping(outputs, nil)
	if !reflect.DeepEqual(got, outputs) {
		t.Errorf("nil mapping should return original; got %v", got)
	}
}

func TestApplyMapping_MissingSources(t *testing.T) {
	outputs := map[string]string{"id": "abc"}
	m := &OutputMapping{
		Mappings: map[string]string{"nonexistent_key": "target"},
	}

	got := applyMapping(outputs, m)
	want := map[string]string{"id": "abc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("missing source should be a no-op; got %v, want %v", got, want)
	}
}

func TestApplyMapping_DoesNotModifyOriginal(t *testing.T) {
	outputs := map[string]string{
		"custom_id": "abc",
		"other":     "val",
	}
	m := &OutputMapping{
		Mappings: map[string]string{"custom_id": "id"},
	}

	applyMapping(outputs, m)

	if _, ok := outputs["custom_id"]; !ok {
		t.Error("original map was modified: custom_id missing")
	}
	if _, ok := outputs["id"]; ok {
		t.Error("original map was modified: id key was added")
	}
}
