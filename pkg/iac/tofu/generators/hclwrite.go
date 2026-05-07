package generators

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// topLevelSkipKeys lists keys that are always omitted from tfvars output.
// These are proto envelope fields that have no meaning in Terraform.
// Keys are listed in both camelCase (pre-flatten JSON) and snake_case
// (post-flatten proto name) to work regardless of whether Flatten was called.
var topLevelSkipKeys = map[string]bool{
	"apiVersion":  true,
	"api_version": true,
	"kind":        true,
	"status":      true,
}

// WriteMapToHCL formats a map[string]interface{} into HCL-compatible syntax
// suitable for Terraform tfvars files. It writes the result to buf.
//
// Top-level keys (indentLevel 0) are unquoted per HCL spec. Nested keys are
// quoted to safely handle special characters (periods, slashes) common in
// Kubernetes-style labels. All keys are converted to snake_case.
func WriteMapToHCL(buf *bytes.Buffer, data map[string]interface{}, indentLevel int) error {
	indent := strings.Repeat("  ", indentLevel)

	for k, val := range data {
		if indentLevel == 0 && topLevelSkipKeys[k] {
			continue
		}

		// Keys are expected to already be in the correct form: proto field
		// names are snake_cased by the Flatten step (which has the proto
		// descriptor), and user-defined map keys (like env var names) are
		// preserved verbatim. The HCL writer does NOT apply case conversion.
		formattedKey := formatKey(k, indentLevel)

		switch typedVal := val.(type) {
		case map[string]interface{}:
			buf.WriteString(fmt.Sprintf("%s%s = {\n", indent, formattedKey))
			if err := WriteMapToHCL(buf, typedVal, indentLevel+1); err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s}\n", indent))

		case []interface{}:
			buf.WriteString(fmt.Sprintf("%s%s = [\n", indent, formattedKey))
			if err := writeArrayToHCL(buf, typedVal, indentLevel+1); err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s]\n", indent))

		case string:
			buf.WriteString(fmt.Sprintf("%s%s = %q\n", indent, formattedKey, typedVal))

		case bool:
			buf.WriteString(fmt.Sprintf("%s%s = %t\n", indent, formattedKey, typedVal))

		case float64:
			buf.WriteString(fmt.Sprintf("%s%s = %v\n", indent, formattedKey, typedVal))

		case nil:
			buf.WriteString(fmt.Sprintf("%s%s = null\n", indent, formattedKey))

		default:
			return errors.Errorf("unsupported type for key %q: %T", k, val)
		}
	}

	return nil
}

// writeArrayToHCL formats a []interface{} as an HCL array body (without the
// surrounding brackets, which the caller writes).
func writeArrayToHCL(buf *bytes.Buffer, arr []interface{}, indentLevel int) error {
	indent := strings.Repeat("  ", indentLevel)

	for _, elem := range arr {
		switch typedElem := elem.(type) {
		case string:
			buf.WriteString(fmt.Sprintf("%s%q,\n", indent, typedElem))

		case bool:
			buf.WriteString(fmt.Sprintf("%s%t,\n", indent, typedElem))

		case float64:
			buf.WriteString(fmt.Sprintf("%s%v,\n", indent, typedElem))

		case map[string]interface{}:
			buf.WriteString(fmt.Sprintf("%s{\n", indent))
			if err := WriteMapToHCL(buf, typedElem, indentLevel+1); err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s},\n", indent))

		case []interface{}:
			buf.WriteString(fmt.Sprintf("%s[\n", indent))
			if err := writeArrayToHCL(buf, typedElem, indentLevel+1); err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s],\n", indent))

		case nil:
			buf.WriteString(fmt.Sprintf("%snull,\n", indent))

		default:
			return errors.Errorf("unsupported array element type: %T", typedElem)
		}
	}

	return nil
}

// formatKey formats a map key for HCL output. Top-level keys (indentLevel 0)
// must NOT be quoted in tfvars files. Nested keys are quoted to safely handle
// special characters.
func formatKey(key string, indentLevel int) string {
	if indentLevel == 0 {
		return key
	}
	return fmt.Sprintf("%q", key)
}
