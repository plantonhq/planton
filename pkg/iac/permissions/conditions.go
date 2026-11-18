package permissions

import (
	"regexp"
	"strings"
)

// ConditionNeeded reports whether a permission entry whose condition names
// specFieldSet is needed by any of the given manifest specs (each a manifest's
// `spec`, decoded from YAML or JSON into generic maps). An entry with no
// condition (an empty path) is always needed.
//
// "Set" follows the catalog's definition: a list or map is non-empty, an
// object is present, a scalar holds a non-default value. The condition names
// the field by its proto name (enabled_clients) while a manifest may spell it
// in its JSON name (enabledClients); both spellings are read. A path that
// crosses a list (containers.env.secret_value) is set when any element sets
// the rest of it.
//
// The platform decides the same entries for a chart's runner policy and a
// stack plan (PermissionConditions in the platform's infra-hub commons, and
// the console's permission-conditions module). All three read one case table,
// testdata/permission-conditions.cases.json here, copied verbatim from the
// platform, so none can drift from the others unnoticed.
func ConditionNeeded(specFieldSet string, specs []map[string]interface{}) bool {
	if specFieldSet == "" {
		return true
	}
	path := strings.Split(specFieldSet, ".")
	for _, spec := range specs {
		if pathIsSet(spec, path) {
			return true
		}
	}
	return false
}

func pathIsSet(value interface{}, path []string) bool {
	if len(path) == 0 {
		return isSet(value)
	}
	switch typed := value.(type) {
	case []interface{}:
		for _, element := range typed {
			if pathIsSet(element, path) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		next, ok := typed[path[0]]
		if !ok {
			next, ok = typed[jsonName(path[0])]
		}
		return ok && pathIsSet(next, path[1:])
	default:
		return false
	}
}

var snakeSegment = regexp.MustCompile(`_([a-z0-9])`)

// jsonName is a field's proto JSON name: lowerCamelCase of its snake_case name.
func jsonName(protoName string) string {
	return snakeSegment.ReplaceAllStringFunc(protoName, func(match string) string {
		return strings.ToUpper(match[1:])
	})
}

func isSet(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case []interface{}:
		return len(typed) > 0
	case map[string]interface{}:
		return true
	case string:
		return typed != ""
	case bool:
		return typed
	case float64:
		return typed != 0
	case int:
		return typed != 0
	case int64:
		return typed != 0
	default:
		return true
	}
}
