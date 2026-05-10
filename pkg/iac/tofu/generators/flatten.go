package generators

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Flatten walks a JSON map (produced by protojson.Marshal + json.Unmarshal)
// alongside a proto message descriptor, applying type rules to transform the
// map in-place. Wrapper types are flattened to primitives, skipped fields are
// removed, and proto field keys are renamed from JSON camelCase to snake_case.
//
// Key renaming is done here (not in the HCL writer) because only the flatten
// step has the proto descriptor to distinguish proto field names (which need
// snake_case conversion) from user-defined map keys like environment variable
// names (which must be preserved verbatim).
//
// The function handles all three shapes in which a flattened type can appear:
//
//   - Singular message field: {"namespace": {"value": "ns"}} -> {"namespace": "ns"}
//   - Map with message values: {"variables": {"K": {"value": "v"}}} -> {"variables": {"K": "v"}}
//   - Repeated message field: [{"value": "a"}, {"value": "b"}] -> ["a", "b"]
//
// After processing each field, if the value remains a nested map (not flattened
// and not skipped), the function recurses into it with the corresponding nested
// message descriptor.
func Flatten(data map[string]interface{}, md protoreflect.MessageDescriptor, rules map[string]TypeRule) {
	fields := md.Fields()

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		jsonKey := jsonFieldName(fd)

		val, exists := data[jsonKey]
		if !exists {
			continue
		}

		// Rename proto field key from JSON camelCase to snake_case.
		// The proto field name is already snake_case by convention.
		snakeKey := string(fd.Name())
		if snakeKey != jsonKey {
			delete(data, jsonKey)
			data[snakeKey] = val
		}
		activeKey := snakeKey

		if fd.Kind() != protoreflect.MessageKind {
			continue
		}

		if fd.IsMap() {
			flattenMapField(data, activeKey, fd, val, rules)
			continue
		}

		if fd.IsList() {
			flattenListField(data, activeKey, fd, val, rules)
			continue
		}

		flattenSingularField(data, activeKey, fd, val, rules)
	}
}

// flattenSingularField handles a non-repeated, non-map message field.
func flattenSingularField(data map[string]interface{}, jsonKey string, fd protoreflect.FieldDescriptor, val interface{}, rules map[string]TypeRule) {
	msgName := string(fd.Message().FullName())
	rule, hasRule := rules[msgName]

	if hasRule && rule.Skip {
		delete(data, jsonKey)
		return
	}

	if hasRule && rule.FlattenTo != "" && rule.ExtractValue != nil {
		extracted, err := rule.ExtractValue(val)
		if err != nil {
			// On extraction error, remove the field rather than emitting
			// malformed output. The error likely means valueFrom was present,
			// which is an orchestrator concept that should have been resolved.
			delete(data, jsonKey)
			return
		}
		data[jsonKey] = extracted
		return
	}

	// No rule -- recurse into the nested map if possible.
	if nested, ok := val.(map[string]interface{}); ok {
		Flatten(nested, fd.Message(), rules)
	}
}

// flattenMapField handles a proto map<K, V> field. In JSON, proto maps are
// objects: {"key1": value1, "key2": value2}. When the map's value type has a
// flatten rule, each value in the map is flattened individually.
func flattenMapField(data map[string]interface{}, jsonKey string, fd protoreflect.FieldDescriptor, val interface{}, rules map[string]TypeRule) {
	mapObj, ok := val.(map[string]interface{})
	if !ok {
		return
	}

	valueDesc := fd.MapValue()
	if valueDesc.Kind() != protoreflect.MessageKind {
		return
	}

	valueMsgName := string(valueDesc.Message().FullName())
	rule, hasRule := rules[valueMsgName]

	if hasRule && rule.Skip {
		delete(data, jsonKey)
		return
	}

	if hasRule && rule.FlattenTo != "" && rule.ExtractValue != nil {
		for k, v := range mapObj {
			extracted, err := rule.ExtractValue(v)
			if err != nil {
				delete(mapObj, k)
				continue
			}
			mapObj[k] = extracted
		}
		return
	}

	// No flatten rule on the value type -- recurse into each value.
	for _, v := range mapObj {
		if nested, ok := v.(map[string]interface{}); ok {
			Flatten(nested, valueDesc.Message(), rules)
		}
	}
}

// flattenListField handles a repeated message field. In JSON, these are arrays.
// When the element type has a flatten rule, each element is flattened.
func flattenListField(data map[string]interface{}, jsonKey string, fd protoreflect.FieldDescriptor, val interface{}, rules map[string]TypeRule) {
	arr, ok := val.([]interface{})
	if !ok {
		return
	}

	elemMsgName := string(fd.Message().FullName())
	rule, hasRule := rules[elemMsgName]

	if hasRule && rule.Skip {
		delete(data, jsonKey)
		return
	}

	if hasRule && rule.FlattenTo != "" && rule.ExtractValue != nil {
		for i, elem := range arr {
			extracted, err := rule.ExtractValue(elem)
			if err != nil {
				arr[i] = nil
				continue
			}
			arr[i] = extracted
		}
		return
	}

	// No flatten rule -- recurse into each element.
	for _, elem := range arr {
		if nested, ok := elem.(map[string]interface{}); ok {
			Flatten(nested, fd.Message(), rules)
		}
	}
}

// jsonFieldName returns the JSON field name that protojson uses for a proto
// field. protojson uses lowerCamelCase (the JSON name from the proto
// descriptor), not the proto field name.
func jsonFieldName(fd protoreflect.FieldDescriptor) string {
	// protojson uses the field's JSONName(), which is the lowerCamelCase
	// form of the proto field name (or an explicit json_name option).
	name := fd.JSONName()
	if name == "" {
		return string(fd.Name())
	}
	return name
}

// lookupRule returns the TypeRule for a message field's type, if one exists.
// Exported for use by the variablestf generator.
func lookupRule(fd protoreflect.FieldDescriptor, rules map[string]TypeRule) (TypeRule, bool) {
	if fd.Kind() != protoreflect.MessageKind {
		return TypeRule{}, false
	}

	var msgDesc protoreflect.MessageDescriptor
	if fd.IsMap() {
		valDesc := fd.MapValue()
		if valDesc.Kind() != protoreflect.MessageKind {
			return TypeRule{}, false
		}
		msgDesc = valDesc.Message()
	} else {
		msgDesc = fd.Message()
	}

	fullName := string(msgDesc.FullName())
	rule, ok := rules[fullName]
	return rule, ok
}

// fieldMessageFullName returns the full proto name of the message type for a
// field, handling map fields by looking at the map value descriptor. Returns
// empty string for non-message fields.
func fieldMessageFullName(fd protoreflect.FieldDescriptor) string {
	if fd.Kind() != protoreflect.MessageKind {
		return ""
	}
	if fd.IsMap() {
		valDesc := fd.MapValue()
		if valDesc.Kind() != protoreflect.MessageKind {
			return ""
		}
		return string(valDesc.Message().FullName())
	}
	return string(fd.Message().FullName())
}

// resolveMessageRule checks the rules map for a given proto full name and
// returns the rule if found. This is a convenience for callers that already
// have the full name string.
func resolveMessageRule(fullName string, rules map[string]TypeRule) (TypeRule, bool) {
	if fullName == "" {
		return TypeRule{}, false
	}
	rule, ok := rules[fullName]
	return rule, ok
}

// init-time validation: ensure DefaultRules ExtractValue functions are non-nil
// where FlattenTo is set. Catches programming errors at import time.
func init() {
	for name, rule := range DefaultRules() {
		if rule.FlattenTo != "" && rule.ExtractValue == nil {
			panic(fmt.Sprintf("generators: rule %q has FlattenTo=%q but nil ExtractValue", name, rule.FlattenTo))
		}
	}
}
