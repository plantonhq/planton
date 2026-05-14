package outputs

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Flatten converts a map[string]interface{} (as produced by Pulumi's
// automation API or any JSON-deserialized output map) into a flat
// map[string]string suitable for the IacStackOutputsPayload wire format.
//
// Type coercion rules:
//   - string: direct copy
//   - float64: integer formatting when the value has no fractional part
//     (e.g., 3600 -> "3600"), decimal otherwise (e.g., 3.14 -> "3.14")
//   - bool: "true" / "false"
//   - json.Number: string representation
//   - nil: empty string
//   - []interface{}: recursive flattening with dot-indexed keys (key.0, key.1, ...)
//   - map[string]interface{}: recursive dot-path flattening (parent.child = value)
//
// This function replaces the lossy ConvertInterfaceMapToStringMap which
// turned non-string scalars (int, float, bool) into the literal "unknown".
func Flatten(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	if len(m) == 0 {
		return result
	}
	flattenMap(m, "", result)
	return result
}

// flattenMap recursively walks a map and writes dot-path keys into out.
func flattenMap(m map[string]interface{}, prefix string, out map[string]string) {
	keys := sortedKeys(m)
	for _, key := range keys {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		flattenValue(fullKey, m[key], out)
	}
}

// flattenValue dispatches a single value to the appropriate flattening strategy.
func flattenValue(key string, val interface{}, out map[string]string) {
	if val == nil {
		out[key] = ""
		return
	}

	switch v := val.(type) {
	case string:
		out[key] = v

	case float64:
		out[key] = formatFloat64(v)

	case bool:
		out[key] = strconv.FormatBool(v)

	case json.Number:
		out[key] = v.String()

	case []interface{}:
		if len(v) == 0 {
			out[key] = ""
			return
		}
		for i, item := range v {
			subKey := fmt.Sprintf("%s.%d", key, i)
			flattenValue(subKey, item, out)
		}

	case map[string]interface{}:
		if len(v) == 0 {
			out[key] = ""
			return
		}
		flattenMap(v, key, out)

	default:
		out[key] = fmt.Sprintf("%v", v)
	}
}

// formatFloat64 converts a float64 to its most compact string form.
// Whole numbers (3600.0) are formatted as integers ("3600").
// Fractional values are formatted without trailing zeros ("3.14").
func formatFloat64(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Sprintf("%v", f)
	}
	if f == math.Trunc(f) && !math.IsInf(f, 0) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// sortedKeys returns the keys of a map in sorted order for deterministic output.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
