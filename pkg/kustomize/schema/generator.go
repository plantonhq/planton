// Build tag prevents compilation during codegen (uses crkreflect.NewInstance which
// depends on the generated kind_map_gen.go).

//go:build !codegen
// +build !codegen

package schema

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/plantonhq/planton/pkg/crkreflect"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SchemaFileName is the conventional name for the generated schema file.
const SchemaFileName = "planton-schema.json"

// mergeField represents a repeated list field that should merge by key during
// kustomize strategic merge patch instead of being replaced wholesale.
type mergeField struct {
	// jsonPath is the dot-separated path from the spec root using JSON (camelCase) field names.
	// Example: "container.app.env.variables"
	jsonPath string
	// mergeKey is the field name used as the merge key. Currently always "name".
	mergeKey string
}

// Generate produces a kustomize-compatible OpenAPI schema JSON that declares
// strategic merge patch directives for all cloud resource kinds that have
// repeated message fields with a "name" merge key.
//
// Only kinds that actually contain merge-worthy fields produce entries.
// The output is a single JSON document suitable for use with the kustomize
// "openapi:" directive in kustomization.yaml.
func Generate() ([]byte, error) {
	definitions := make(map[string]any)

	for _, kind := range crkreflect.KindsList() {
		msg, err := crkreflect.NewInstance(kind)
		if err != nil {
			log.Debugf("skipping kind %s: %v", kind.String(), err)
			continue
		}

		md := msg.ProtoReflect().Descriptor()
		specField := md.Fields().ByJSONName("spec")
		if specField == nil || specField.Kind() != protoreflect.MessageKind {
			continue
		}

		visited := make(map[protoreflect.FullName]bool)
		fields := findMergeFields(specField.Message(), visited)
		if len(fields) == 0 {
			continue
		}

		kindName := kind.String()
		groupVersion := crkreflect.GroupVersion(kind)
		if groupVersion == "" {
			continue
		}

		parts := strings.SplitN(groupVersion, "/", 2)
		if len(parts) != 2 {
			continue
		}
		group, version := parts[0], parts[1]

		defKey := group + "." + version + "." + kindName
		definitions[defKey] = buildDefinition(group, version, kindName, fields)
	}

	schema := map[string]any{
		"definitions": definitions,
	}

	return json.MarshalIndent(schema, "", "  ")
}

// findMergeFields recursively walks a proto message descriptor and returns all
// repeated message fields whose element type has a "name" field (the natural
// kustomize merge key). Map fields are skipped. The visited set prevents
// infinite recursion on cyclic message references.
func findMergeFields(md protoreflect.MessageDescriptor, visited map[protoreflect.FullName]bool) []mergeField {
	if visited[md.FullName()] {
		return nil
	}
	visited[md.FullName()] = true
	defer delete(visited, md.FullName())

	var result []mergeField
	fields := md.Fields()

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() {
			continue
		}

		jsonName := fd.JSONName()

		if fd.IsList() {
			if hasNameField(fd.Message()) {
				result = append(result, mergeField{jsonPath: jsonName, mergeKey: "name"})
			}
		} else {
			for _, sub := range findMergeFields(fd.Message(), visited) {
				result = append(result, mergeField{
					jsonPath: jsonName + "." + sub.jsonPath,
					mergeKey: sub.mergeKey,
				})
			}
		}
	}

	return result
}

// hasNameField returns true if the message has a top-level string field named "name".
func hasNameField(md protoreflect.MessageDescriptor) bool {
	f := md.Fields().ByName("name")
	return f != nil && f.Kind() == protoreflect.StringKind
}

// buildDefinition constructs the OpenAPI definition object for a single cloud resource kind.
func buildDefinition(group, version, kind string, fields []mergeField) map[string]any {
	return map[string]any{
		"x-kubernetes-group-version-kind": []map[string]string{
			{"group": group, "version": version, "kind": kind},
		},
		"properties": map[string]any{
			"spec": map[string]any{
				"properties": buildNestedProperties(fields),
			},
		},
	}
}

// buildNestedProperties takes a flat list of dot-separated merge field paths
// and constructs the nested OpenAPI "properties" tree.
//
// For paths ["container.app.env.variables", "container.app.env.secrets", "container.app.ports"],
// it produces:
//
//	{
//	  "container": {
//	    "properties": {
//	      "app": {
//	        "properties": {
//	          "env": {
//	            "properties": {
//	              "variables": { "type": "array", "x-kubernetes-patch-merge-key": "name", ... },
//	              "secrets":   { "type": "array", "x-kubernetes-patch-merge-key": "name", ... }
//	            }
//	          },
//	          "ports": { "type": "array", "x-kubernetes-patch-merge-key": "name", ... }
//	        }
//	      }
//	    }
//	  }
//	}
func buildNestedProperties(fields []mergeField) map[string]any {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].jsonPath < fields[j].jsonPath
	})

	root := make(map[string]any)

	for _, mf := range fields {
		parts := strings.Split(mf.jsonPath, ".")
		current := root

		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = map[string]any{
					"type":                         "array",
					"x-kubernetes-patch-merge-key": mf.mergeKey,
					"x-kubernetes-patch-strategy":  "merge",
				}
			} else {
				if existing, ok := current[part]; ok {
					node := existing.(map[string]any)
					current = node["properties"].(map[string]any)
				} else {
					props := make(map[string]any)
					current[part] = map[string]any{
						"properties": props,
					}
					current = props
				}
			}
		}
	}

	return root
}
