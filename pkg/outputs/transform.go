//go:build !codegen
// +build !codegen

// Package outputs transforms flat IaC output maps into typed Outputs protos.
//
// IaC engines (Terraform and Pulumi) produce outputs as flat map[string]string.
// Each CatalogKind in Planton has a typed Outputs proto that defines
// the expected output fields with their concrete types. This package bridges the
// two representations using proto reflection, so it works for all 365+ kind
// kinds without per-kind code.
//
// The public entry point is Transform(). See resolve.go for how the Outputs
// message type is discovered, populate.go for how fields are set, and
// preprocess.go for key normalization.
package outputs

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// Transform takes a CatalogKind and a flat map of IaC outputs (as produced
// by Terraform/Pulumi runners) and returns a typed Outputs proto message
// populated via proto reflection.
//
// The returned message is the concrete generated Go type for the kind's
// Outputs (e.g., *Auth0ResourceServerOutputs). Callers that know the
// expected type can safely type-assert the result.
//
// Empty or nil output maps produce an empty (zero-value) Outputs message,
// not an error. Unknown output keys that have no corresponding proto field are
// logged as warnings and skipped.
func Transform(
	kind catalogkind.CatalogKind,
	outputs map[string]string,
) (proto.Message, error) {
	msg, err := resolveOutputsMessage(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve Outputs message for kind %s", kind.String())
	}

	if len(outputs) == 0 {
		return msg, nil
	}

	normalized := preprocessKeys(outputs)

	if err := populateMessage(msg, normalized); err != nil {
		return nil, errors.Wrapf(err, "failed to populate Outputs for kind %s", kind.String())
	}

	return msg, nil
}
