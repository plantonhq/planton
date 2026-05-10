package generators

import "testing"

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
		{"google.protobuf.Struct", false, "string"},
		{"google.protobuf.Value", false, "string"},
		{"google.protobuf.ListValue", false, "string"},
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

func TestExtractJSONString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"map", map[string]interface{}{"k": "v"}, "map[k:v]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractJSONString(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %q", got, tc.want)
			}
		})
	}
}
