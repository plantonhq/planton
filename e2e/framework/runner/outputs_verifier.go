package runner

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/outputs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// VerifyOutputTransformation validates that raw IaC engine outputs can be
// correctly transformed into a typed StackOutputs proto message.
//
// The pipeline uses TransformRaw which supports module-directory overrides:
//   - transform-outputs executable (highest priority)
//   - output_transform.yaml declarative mapping
//   - generic reflection transformer (default fallback)
//
// Returns the populated proto message on success, or an error if any step fails.
// Unknown output fields are logged as warnings by Transform() but do not cause failure.
func VerifyOutputTransformation(component string, rawOutputs map[string]interface{}, moduleDir string) (proto.Message, map[string]string, error) {
	kind := crkreflect.KindFromString(component)
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return nil, nil, errors.Errorf("cannot resolve CloudResourceKind from component name %q", component)
	}

	var opts *outputs.TransformOptions
	if moduleDir != "" {
		opts = &outputs.TransformOptions{ModuleDir: moduleDir}
	}

	msg, flatOutputs, err := outputs.TransformRaw(kind, rawOutputs, opts)
	if err != nil {
		return nil, flatOutputs, errors.Wrapf(err, "output transformation failed for %s (kind=%s)", component, kind.String())
	}

	logTransformationSummary(component, kind, flatOutputs, msg)

	return msg, flatOutputs, nil
}

// logTransformationSummary prints how many proto fields were populated vs available.
func logTransformationSummary(component string, kind cloudresourcekind.CloudResourceKind, flatOutputs map[string]string, msg proto.Message) {
	ref := msg.ProtoReflect()
	totalFields := ref.Descriptor().Fields().Len()

	populated := 0
	ref.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		populated++
		return true
	})

	fmt.Printf("  [outputs] %s (%s): %d/%d proto fields populated from %d raw outputs\n",
		component, kind.String(), populated, totalFields, len(flatOutputs))
}
