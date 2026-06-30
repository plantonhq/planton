package generators

import "fmt"

// TypeRule defines how a specific proto message type should be treated when
// generating Terraform artifacts. Rules are registered once and consulted by
// both the tfvars and variables.tf generators.
type TypeRule struct {
	// FlattenTo is the Terraform type this message should collapse to. When
	// set, the generators emit that type instead of recursing into the
	// message's fields, and the tfvars generator calls ExtractValue to produce
	// the value. Valid values: "string", "number", "bool", and "any" (for
	// free-form JSON well-known types passed through verbatim).
	// Empty string means "do not flatten, recurse normally."
	FlattenTo string

	// Skip means omit this field entirely from Terraform output. Use for
	// orchestrator-only fields that have no meaning in standalone TF modules.
	Skip bool

	// ExtractValue extracts the flattened primitive from a JSON-unmarshaled
	// value. Called by the tfvars generator when FlattenTo is set. Receives
	// the map[string]interface{} (or other JSON value) that protojson produced
	// for this message.
	//
	// For StringValueOrRef: the JSON is {"value": "..."} and ExtractValue
	// returns the string. When the oneof is value_from instead, the Planton
	// orchestrator resolves it before IaC runs, so only the value arm appears
	// in practice.
	ExtractValue func(jsonVal interface{}) (interface{}, error)
}

// DefaultRules returns the standard Planton type rules. Adding a new wrapper
// type means adding one entry here -- no generator code changes needed.
func DefaultRules() map[string]TypeRule {
	return map[string]TypeRule{
		// StringValueOrRef: proto oneof with {string value, ValueFromRef value_from}.
		// The value_from arm is a Planton orchestrator concept resolved before IaC
		// runs. TF modules should see a plain string.
		"dev.planton.shared.foreignkey.v1.StringValueOrRef": {
			FlattenTo:    "string",
			ExtractValue: extractStringValueOrRef,
		},

		// ValueFromRef: Planton-internal reference type. Should never appear in
		// TF output (the orchestrator resolves it before IaC invocation).
		"dev.planton.shared.foreignkey.v1.ValueFromRef": {
			Skip: true,
		},

		// KubernetesClusterSelector: tells the Planton orchestrator which cluster
		// to target. The TF module gets the cluster via KUBE_CONFIG_PATH / provider
		// config, not this field.
		"dev.planton.provider.kubernetes.KubernetesClusterSelector": {
			Skip: true,
		},

		// google.protobuf well-known JSON types represent free-form JSON. They
		// flatten to Terraform `any` and pass through verbatim as the nested
		// JSON value (object/array/scalar) that protojson produced -- NOT
		// recursing into the proto wrapper structure (whose `fields`/`values`
		// internals are not the user-facing shape). A TF module that targets a
		// JSON-string argument calls jsonencode() on the value; a module that
		// wants the structured object (e.g. a kubernetes_manifest free-form
		// field) passes it through directly.
		"google.protobuf.Struct": {
			FlattenTo:    "any",
			ExtractValue: extractJSONValue,
		},
		"google.protobuf.Value": {
			FlattenTo:    "any",
			ExtractValue: extractJSONValue,
		},
		"google.protobuf.ListValue": {
			FlattenTo:    "any",
			ExtractValue: extractJSONValue,
		},
	}
}

// extractStringValueOrRef extracts the "value" string from a StringValueOrRef
// JSON representation: {"value": "..."}.
//
// When the oneof is value_from, protojson emits {"valueFrom": {...}} instead.
// Since value_from is a Planton orchestrator concept resolved before IaC runs,
// encountering it here means the manifest was not pre-processed. We return an
// error in that case rather than silently producing wrong output.
func extractStringValueOrRef(jsonVal interface{}) (interface{}, error) {
	m, ok := jsonVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("StringValueOrRef: expected map, got %T", jsonVal)
	}

	if v, exists := m["value"]; exists {
		return v, nil
	}

	if _, exists := m["valueFrom"]; exists {
		return nil, fmt.Errorf("StringValueOrRef: valueFrom references must be " +
			"resolved by the orchestrator before IaC invocation")
	}

	// Empty StringValueOrRef (no oneof arm set) -- emit empty string.
	return "", nil
}

// extractJSONValue passes a google.protobuf.Struct/Value/ListValue through
// verbatim: protojson already rendered it as the natural JSON value (a nested
// object, array, or scalar), and the HCL writer emits those nested shapes
// directly. Returning the value unchanged preserves the free-form content
// faithfully (including user-defined keys, which must not be snake_cased), so
// the resulting tfvars carries a real object rather than a stringified blob.
func extractJSONValue(jsonVal interface{}) (interface{}, error) {
	return jsonVal, nil
}
