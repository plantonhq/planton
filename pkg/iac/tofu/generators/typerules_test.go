package generators

import (
	"reflect"
	"testing"
)

func TestDefaultRules_ContainsExpectedEntries(t *testing.T) {
	rules := DefaultRules()

	expected := []struct {
		name     string
		wantSkip bool
		wantFlat string
	}{
		{"org.openmcf.shared.foreignkey.v1.StringValueOrRef", false, "string"},
		{"org.openmcf.shared.foreignkey.v1.ValueFromRef", true, ""},
		{"org.openmcf.provider.kubernetes.KubernetesClusterSelector", true, ""},
		{"google.protobuf.Struct", false, "any"},
		{"google.protobuf.Value", false, "any"},
		{"google.protobuf.ListValue", false, "any"},
	}

	for _, tc := range expected {
		rule, ok := rules[tc.name]
		if !ok {
			t.Errorf("DefaultRules() missing entry for %q", tc.name)
			continue
		}
		if rule.Skip != tc.wantSkip {
			t.Errorf("rule %q: Skip = %v, want %v", tc.name, rule.Skip, tc.wantSkip)
		}
		if rule.FlattenTo != tc.wantFlat {
			t.Errorf("rule %q: FlattenTo = %q, want %q", tc.name, rule.FlattenTo, tc.wantFlat)
		}
	}

	if len(rules) != len(expected) {
		t.Errorf("DefaultRules() has %d entries, want %d", len(rules), len(expected))
	}
}

func TestExtractStringValueOrRef_Value(t *testing.T) {
	input := map[string]interface{}{"value": "my-namespace"}
	got, err := extractStringValueOrRef(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-namespace" {
		t.Errorf("got %v, want %q", got, "my-namespace")
	}
}

func TestExtractStringValueOrRef_EmptyOneof(t *testing.T) {
	input := map[string]interface{}{}
	got, err := extractStringValueOrRef(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %v, want empty string", got)
	}
}

func TestExtractStringValueOrRef_ValueFrom_ReturnsError(t *testing.T) {
	input := map[string]interface{}{
		"valueFrom": map[string]interface{}{
			"name": "my-postgres",
		},
	}
	_, err := extractStringValueOrRef(input)
	if err == nil {
		t.Fatal("expected error for valueFrom, got nil")
	}
}

func TestExtractStringValueOrRef_WrongType(t *testing.T) {
	_, err := extractStringValueOrRef("not-a-map")
	if err == nil {
		t.Fatal("expected error for non-map input, got nil")
	}
}

func TestExtractJSONValue_PassesThroughVerbatim(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"nil", nil},
		{"string", "hello"},
		{"number", float64(5)},
		{"bool", true},
		{"nested object preserves keys verbatim", map[string]interface{}{
			"connect_timeout": "5s",
			"typed_config":    map[string]interface{}{"@type": "x"},
		}},
		{"array", []interface{}{"a", float64(1), map[string]interface{}{"k": "v"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractJSONValue(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.input) {
				t.Errorf("extractJSONValue mutated the value: got %#v, want %#v", got, tc.input)
			}
		})
	}
}
