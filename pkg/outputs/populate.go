package outputs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// indexPattern matches an array index at the end of a field path segment,
// e.g., "subnets[0]" captures "subnets" and "0".
var indexPattern = regexp.MustCompile(`^(.+)\[(\d+)]$`)

// populateMessage sets fields on a proto message from a flat map of string
// key-value pairs. Keys are dot-separated field paths that may include array
// indices (field[0] or field.0 notation).
//
// This is the Go equivalent of Java's StackOutputsMapToProtoLoader.load().
//
// Unknown fields are logged as warnings and skipped rather than causing errors,
// because IaC modules may export outputs that have no corresponding proto field.
func populateMessage(msg proto.Message, outputs map[string]string) error {
	ref := msg.ProtoReflect()
	for key, value := range outputs {
		parts := strings.Split(key, ".")
		if err := setFieldRecursively(ref, parts, value, 0); err != nil {
			log.WithFields(log.Fields{
				"key":   key,
				"value": value,
				"error": err,
			}).Warn("skipping output field that could not be set")
		}
	}
	return nil
}

// setFieldRecursively walks a dot-separated field path and sets the leaf value
// on the proto message. Handles scalar fields, repeated fields (both primitives
// and messages), map fields, and nested message fields.
//
// The function mirrors Java's StackOutputsMapToProtoLoader.setFieldRecursively()
// with Go-specific protoreflect APIs.
func setFieldRecursively(
	msg protoreflect.Message,
	fieldPath []string,
	value string,
	pathIndex int,
) error {
	segment := fieldPath[pathIndex]

	// Check for bracket-style array index: "field[N]"
	fieldName, arrayIdx, hasArrayIdx := parseArrayIndex(segment)

	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if fd == nil {
		return fmt.Errorf("field %q not found on message %s", fieldName, msg.Descriptor().FullName())
	}

	if fd.IsMap() {
		return handleMapField(msg, fd, value)
	}

	if fd.IsList() {
		return handleRepeatedField(msg, fd, fieldPath, value, pathIndex, fieldName, arrayIdx, hasArrayIdx)
	}

	// Singular field
	isLeaf := pathIndex == len(fieldPath)-1

	if isLeaf {
		if fd.Kind() == protoreflect.MessageKind {
			return setMessageFieldFromJSON(msg, fd, value)
		}
		v, err := convertScalar(value, fd)
		if err != nil {
			return err
		}
		msg.Set(fd, v)
		return nil
	}

	// Non-leaf: recurse into nested message
	if fd.Kind() != protoreflect.MessageKind {
		return fmt.Errorf("field %q is %s, cannot recurse into non-message at path index %d",
			fieldName, fd.Kind(), pathIndex)
	}
	nested := msg.Mutable(fd).Message()
	return setFieldRecursively(nested, fieldPath, value, pathIndex+1)
}

// handleRepeatedField handles both repeated primitives (e.g., repeated string)
// and repeated messages (e.g., repeated SubnetStackOutputs).
//
// Supports two index notations:
//   - Bracket: field[0] — index is in the same path segment
//   - Dot-separated: field.0 — index is in the next path segment
func handleRepeatedField(
	msg protoreflect.Message,
	fd protoreflect.FieldDescriptor,
	fieldPath []string,
	value string,
	pathIndex int,
	fieldName string,
	arrayIdx int,
	hasArrayIdx bool,
) error {
	indexSegmentPos := pathIndex

	if !hasArrayIdx {
		// Index might be in the next path segment: "field.0" or "field.0.subfield"
		if pathIndex+1 >= len(fieldPath) {
			// No index provided. If the value is empty, this represents an empty
			// repeated field (Terraform/Pulumi emit "field_name: \"\"" for empty arrays).
			if value == "" {
				return nil
			}
			return fmt.Errorf("repeated field %q: no array index provided", fieldName)
		}
		nextSegment := fieldPath[pathIndex+1]
		idx, err := strconv.Atoi(nextSegment)
		if err != nil {
			return fmt.Errorf("repeated field %q: invalid index %q", fieldName, nextSegment)
		}
		arrayIdx = idx
		indexSegmentPos = pathIndex + 1
	}

	list := msg.Mutable(fd).List()
	ensureListSize(list, fd, arrayIdx)

	isLeaf := indexSegmentPos == len(fieldPath)-1

	if isLeaf {
		if fd.Kind() == protoreflect.MessageKind {
			return setRepeatedMessageFromJSON(list, fd, arrayIdx, value)
		}
		v, err := convertScalar(value, fd)
		if err != nil {
			return err
		}
		list.Set(arrayIdx, v)
		return nil
	}

	// Non-leaf: recurse into nested message at this index
	if fd.Kind() != protoreflect.MessageKind {
		return fmt.Errorf("repeated field %q: cannot recurse into non-message elements", fieldName)
	}
	nested := list.Get(arrayIdx).Message()
	return setFieldRecursively(nested, fieldPath, value, indexSegmentPos+1)
}

// handleMapField parses a JSON string into a proto map field.
// The JSON value is expected to be a JSON object like {"key": "value"}.
func handleMapField(
	msg protoreflect.Message,
	fd protoreflect.FieldDescriptor,
	jsonValue string,
) error {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonValue), &parsed); err != nil {
		return fmt.Errorf("map field %q: failed to parse JSON: %w", fd.Name(), err)
	}

	mapField := msg.Mutable(fd).Map()
	keyFd := fd.MapKey()
	valueFd := fd.MapValue()

	for k, v := range parsed {
		mapKey, err := convertScalar(k, keyFd)
		if err != nil {
			return fmt.Errorf("map field %q key: %w", fd.Name(), err)
		}

		valueStr := fmt.Sprintf("%v", v)
		if valueFd.Kind() == protoreflect.MessageKind {
			// Nested message map values need JSON marshaling
			jsonBytes, jsonErr := json.Marshal(v)
			if jsonErr != nil {
				return fmt.Errorf("map field %q: failed to marshal value for key %q: %w",
					fd.Name(), k, jsonErr)
			}
			newMsg := mapField.NewValue().Message()
			umOpts := protojson.UnmarshalOptions{DiscardUnknown: true}
			if umErr := umOpts.Unmarshal(jsonBytes, newMsg.Interface()); umErr != nil {
				return fmt.Errorf("map field %q: failed to unmarshal value for key %q: %w",
					fd.Name(), k, umErr)
			}
			mapField.Set(mapKey.MapKey(), protoreflect.ValueOfMessage(newMsg))
		} else {
			mapValue, mapErr := convertScalar(valueStr, valueFd)
			if mapErr != nil {
				return fmt.Errorf("map field %q value for key %q: %w", fd.Name(), k, mapErr)
			}
			mapField.Set(mapKey.MapKey(), mapValue)
		}
	}
	return nil
}

// setMessageFieldFromJSON parses a JSON string and sets it as a message field.
// This is the Go equivalent of Java's JsonStringToProtobufMessageMerger.merge().
func setMessageFieldFromJSON(
	msg protoreflect.Message,
	fd protoreflect.FieldDescriptor,
	jsonValue string,
) error {
	nested := msg.Mutable(fd).Message()
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal([]byte(jsonValue), nested.Interface()); err != nil {
		return fmt.Errorf("field %q: failed to unmarshal JSON into message: %w", fd.Name(), err)
	}
	return nil
}

// setRepeatedMessageFromJSON parses a JSON string and sets it at an index in a
// repeated message field.
func setRepeatedMessageFromJSON(
	list protoreflect.List,
	fd protoreflect.FieldDescriptor,
	index int,
	jsonValue string,
) error {
	elem := list.Get(index).Message().Interface()
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal([]byte(jsonValue), elem); err != nil {
		return fmt.Errorf("repeated field %q[%d]: failed to unmarshal JSON: %w",
			fd.Name(), index, err)
	}
	return nil
}

// ensureListSize pads a repeated field's list with default values so that
// index is a valid position. For message fields, empty sub-messages are
// appended. For scalars, zero values are appended.
func ensureListSize(list protoreflect.List, fd protoreflect.FieldDescriptor, index int) {
	for list.Len() <= index {
		if fd.Kind() == protoreflect.MessageKind {
			list.Append(protoreflect.ValueOfMessage(list.NewElement().Message()))
		} else {
			list.Append(list.NewElement())
		}
	}
}

// parseArrayIndex extracts the field name and integer index from a segment
// like "field[3]". Returns the original segment, -1, and false if no bracket
// notation is present.
func parseArrayIndex(segment string) (string, int, bool) {
	matches := indexPattern.FindStringSubmatch(segment)
	if matches == nil {
		return segment, -1, false
	}
	idx, _ := strconv.Atoi(matches[2])
	return matches[1], idx, true
}
