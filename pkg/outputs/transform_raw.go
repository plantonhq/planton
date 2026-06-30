//go:build !codegen
// +build !codegen

package outputs

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// TransformOptions configures optional behavior for TransformRaw.
type TransformOptions struct {
	// ModuleDir is the path to the IaC module directory. When set, the
	// directory is checked for output transformation overrides
	// (transform-outputs executable or output_transform.yaml) before
	// falling back to the generic reflection-based transformer.
	ModuleDir string
}

// TransformRaw takes raw IaC outputs (as produced by Pulumi automation API
// or Terraform JSON) and returns a typed StackOutputs proto message. It
// supports three override levels discovered from the module directory:
//
//  1. transform-outputs executable -- full programmatic control
//  2. output_transform.yaml -- declarative key remapping
//  3. generic reflection transformer -- zero-config default
//
// All three paths ultimately feed into Transform() for proto population.
//
// The second return value is the flat map[string]string that was actually
// fed to the generic transformer, useful for logging and diagnostics.
//
// If opts is nil or opts.ModuleDir is empty, the generic path is used.
func TransformRaw(
	kind cloudresourcekind.CloudResourceKind,
	rawOutputs map[string]interface{},
	opts *TransformOptions,
) (proto.Message, map[string]string, error) {
	moduleDir := ""
	if opts != nil {
		moduleDir = opts.ModuleDir
	}

	override := discoverOverride(moduleDir)

	var flatOutputs map[string]string
	var err error

	switch override {
	case OverrideExecutable:
		log.WithField("moduleDir", moduleDir).Info("using transform-outputs executable override")
		flatOutputs, err = runTransformExecutable(moduleDir, kind, rawOutputs)
		if err != nil {
			return nil, nil, errors.Wrap(err, "executable override failed")
		}

	case OverrideMapping:
		log.WithField("moduleDir", moduleDir).Info("using output_transform.yaml mapping override")
		flatOutputs = Flatten(rawOutputs)

		mapping, loadErr := loadMapping(moduleDir)
		if loadErr != nil {
			return nil, flatOutputs, errors.Wrap(loadErr, "mapping override failed")
		}
		flatOutputs = applyMapping(flatOutputs, mapping)

	default:
		flatOutputs = Flatten(rawOutputs)
	}

	msg, err := Transform(kind, flatOutputs)
	if err != nil {
		return nil, flatOutputs, errors.Wrapf(err, "transform failed (override=%s)", override)
	}

	return msg, flatOutputs, nil
}
