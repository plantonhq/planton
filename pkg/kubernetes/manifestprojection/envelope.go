package manifestprojection

import (
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Envelope names the JSON keys of the spec fields that describe the object
// rather than its spec. An empty key means the kind has no such field: a
// cluster-scoped kind has no namespace, and most kinds carry no labels or
// annotations of their own.
type Envelope struct {
	NamespaceKey   string
	LabelsKey      string
	AnnotationsKey string
}

// Namespaced reports whether the custom resource is namespaced.
func (e Envelope) Namespaced() bool { return e.NamespaceKey != "" }

// Holds reports whether key is one of the envelope's keys, and so never part
// of the custom resource's spec.
func (e Envelope) Holds(key string) bool {
	if key == "" {
		return false
	}
	return key == e.NamespaceKey || key == e.LabelsKey || key == e.AnnotationsKey
}

// EnvelopeOf reads a projection kind's envelope from its spec descriptor.
func EnvelopeOf(spec protoreflect.MessageDescriptor) Envelope {
	var e Envelope
	fields := spec.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if refannotations.Of(fd).DefaultKind == cloudresourcekind.CloudResourceKind_KubernetesNamespace {
			e.NamespaceKey = fd.JSONName()
			continue
		}
		switch ObjectMetadataOf(fd) {
		case options.KubernetesObjectMetadataField_object_labels:
			e.LabelsKey = fd.JSONName()
		case options.KubernetesObjectMetadataField_object_annotations:
			e.AnnotationsKey = fd.JSONName()
		}
	}
	return e
}

// ObjectMetadataOf returns the part of the object's metadata a field is marked
// to hold, or the unspecified value for an unmarked field.
func ObjectMetadataOf(fd protoreflect.FieldDescriptor) options.KubernetesObjectMetadataField {
	opts := fd.Options()
	if opts == nil {
		return options.KubernetesObjectMetadataField_kubernetes_object_metadata_field_unspecified
	}
	v, _ := proto.GetExtension(opts, options.E_KubernetesObjectMetadata).(options.KubernetesObjectMetadataField)
	return v
}
