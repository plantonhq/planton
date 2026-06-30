//go:build !codegen
// +build !codegen

// Package outputs transforms flat IaC output maps into typed StackOutputs protos.
//
// IaC engines (Terraform and Pulumi) produce outputs as flat map[string]string.
// Each CloudResourceKind in Planton has a typed StackOutputs proto that defines
// the expected output fields with their concrete types. This package bridges the
// two representations using proto reflection, so it works for all 365+ component
// kinds without per-component code.
//
// The public entry point is Transform(). See resolve.go for how the StackOutputs
// message type is discovered, populate.go for how fields are set, and
// preprocess.go for key normalization.
package outputs

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

// Transform takes a CloudResourceKind and a flat map of IaC outputs (as produced
// by Terraform/Pulumi runners) and returns a typed StackOutputs proto message
// populated via proto reflection.
//
// The returned message is the concrete generated Go type for the kind's
// StackOutputs (e.g., *Auth0ResourceServerStackOutputs). Callers that know the
// expected type can safely type-assert the result.
//
// Empty or nil output maps produce an empty (zero-value) StackOutputs message,
// not an error. Unknown output keys that have no corresponding proto field are
// logged as warnings and skipped.
func Transform(
	kind cloudresourcekind.CloudResourceKind,
	outputs map[string]string,
) (proto.Message, error) {
	msg, err := resolveStackOutputsMessage(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve StackOutputs message for kind %s", kind.String())
	}

	if len(outputs) == 0 {
		return msg, nil
	}

	normalized := preprocessKeys(outputs)

	if err := populateMessage(msg, normalized); err != nil {
		return nil, errors.Wrapf(err, "failed to populate StackOutputs for kind %s", kind.String())
	}

	return msg, nil
}
