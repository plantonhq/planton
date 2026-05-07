package generators

import "fmt"

// TypeRule defines how a specific proto message type should be treated when
// generating Terraform artifacts. Rules are registered once and consulted by
// both the tfvars and variables.tf generators.
type TypeRule struct {
	// FlattenTo is the Terraform primitive type this message should collapse
	// to. When set, the generators emit a primitive instead of recursing into
	// the message's fields. Valid values: "string", "number", "bool".
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

// DefaultRules returns the standard OpenMCF type rules. Adding a new wrapper
// type means adding one entry here -- no generator code changes needed.
func DefaultRules() map[string]TypeRule {
	return map[string]TypeRule{
		// StringValueOrRef: proto oneof with {string value, ValueFromRef value_from}.
		// The value_from arm is a Planton orchestrator concept resolved before IaC
		// runs. TF modules should see a plain string.
		"org.openmcf.shared.foreignkey.v1.StringValueOrRef": {
			FlattenTo:    "string",
			ExtractValue: extractStringValueOrRef,
		},

		// ValueFromRef: Planton-internal reference type. Should never appear in
		// TF output (the orchestrator resolves it before IaC invocation).
		"org.openmcf.shared.foreignkey.v1.ValueFromRef": {
			Skip: true,
		},

		// KubernetesClusterSelector: tells the Planton orchestrator which cluster
		// to target. The TF module gets the cluster via KUBE_CONFIG_PATH / provider
		// config, not this field.
		"org.openmcf.provider.kubernetes.KubernetesClusterSelector": {
			Skip: true,
		},

		// google.protobuf well-known JSON types: map to string to avoid deep
		// recursion into proto wrapper structures.
		"google.protobuf.Struct": {
			FlattenTo:    "string",
			ExtractValue: extractJSONString,
		},
		"google.protobuf.Value": {
			FlattenTo:    "string",
			ExtractValue: extractJSONString,
		},
		"google.protobuf.ListValue": {
			FlattenTo:    "string",
			ExtractValue: extractJSONString,
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

// extractJSONString converts a google.protobuf.Struct/Value/ListValue JSON
// representation to its string form. These are opaque JSON blobs that TF
// modules treat as strings.
func extractJSONString(jsonVal interface{}) (interface{}, error) {
	if jsonVal == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", jsonVal), nil
}
