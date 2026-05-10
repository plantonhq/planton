package generators

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteMapToHCL_TopLevelKeysUnquoted(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "test",
		},
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "metadata = {") {
		t.Errorf("top-level key should be unquoted, got:\n%s", out)
	}
}

func TestWriteMapToHCL_NestedKeysQuoted(t *testing.T) {
	data := map[string]interface{}{
		"labels": map[string]interface{}{
			"env": "prod",
		},
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 1); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"labels"`) {
		t.Errorf("nested key should be quoted, got:\n%s", out)
	}
}

func TestWriteMapToHCL_SkipsTopLevelEnvelopeFields(t *testing.T) {
	data := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Test",
		"status":     map[string]interface{}{},
		"metadata":   map[string]interface{}{"name": "ok"},
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "api_version") || strings.Contains(out, "apiVersion") {
		t.Error("apiVersion should be skipped")
	}
	if strings.Contains(out, "kind =") {
		t.Error("kind should be skipped")
	}
	if strings.Contains(out, "status") {
		t.Error("status should be skipped")
	}
	if !strings.Contains(out, "metadata") {
		t.Error("metadata should be preserved")
	}
}

func TestWriteMapToHCL_Primitives(t *testing.T) {
	// Keys come pre-converted (snake_case from Flatten, or verbatim for
	// user-defined keys). The HCL writer does not apply case conversion.
	data := map[string]interface{}{
		"str_val":  "hello",
		"bool_val": true,
		"num_val":  float64(42),
		"null_val": nil,
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `str_val = "hello"`) {
		t.Errorf("string formatting wrong, got:\n%s", out)
	}
	if !strings.Contains(out, "bool_val = true") {
		t.Errorf("bool formatting wrong, got:\n%s", out)
	}
	if !strings.Contains(out, "num_val = 42") {
		t.Errorf("number formatting wrong, got:\n%s", out)
	}
	if !strings.Contains(out, "null_val = null") {
		t.Errorf("null formatting wrong, got:\n%s", out)
	}
}

func TestWriteMapToHCL_PreservesKeyAsIs(t *testing.T) {
	// HCL writer does not apply case conversion -- keys come pre-converted
	// from the Flatten step (proto field names) or are user-defined (map keys).
	data := map[string]interface{}{
		"create_namespace": true,
		"DB_HOST":          "localhost",
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "create_namespace = true") {
		t.Errorf("snake_case key should be preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "DB_HOST = ") {
		t.Errorf("user-defined key should be preserved verbatim, got:\n%s", out)
	}
}

func TestWriteMapToHCL_Array(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{"a", "b"},
	}
	var buf bytes.Buffer
	if err := WriteMapToHCL(&buf, data, 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "items = [") {
		t.Errorf("array formatting wrong, got:\n%s", out)
	}
	if !strings.Contains(out, `"a",`) {
		t.Errorf("array element missing, got:\n%s", out)
	}
}
