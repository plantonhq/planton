package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// documentValidation is Auth0's verdict on the metadata document as it stood at
// the last read.
type documentValidation struct {
	Valid      bool
	Warnings   []string
	Violations []string
}

// summarizeValidation reads the provider's validation list, which holds one
// entry when Auth0 previewed the document and none otherwise. An absent entry
// reads as not valid with nothing to report, and the lists are never nil, so
// both engines export the same shape. The Terraform module's outputs.tf applies
// the same rule with try() -- keep them in lockstep.
//
// The SDK's Index(0) accessor cannot be used here: it panics on an empty list.
func summarizeValidation(validations []auth0.ClientCimdValidation) documentValidation {
	summary := documentValidation{Warnings: []string{}, Violations: []string{}}
	if len(validations) == 0 {
		return summary
	}
	first := validations[0]
	if first.Valid != nil {
		summary.Valid = *first.Valid
	}
	if first.Warnings != nil {
		summary.Warnings = first.Warnings
	}
	if first.Violations != nil {
		summary.Violations = first.Violations
	}
	return summary
}

// exportOutputs exports the application as Auth0 registered it: its Management
// API id, what Auth0 took from the document, and the document's validation.
//
// Each validation output applies over the provider's list directly: an
// intermediate Output of the unregistered documentValidation type would be an
// AnyOutput, and a typed applier over it fails at deploy time.
func exportOutputs(ctx *pulumi.Context, client *auth0.ClientCimd) error {
	validations := client.Validations

	ctx.Export("client_id", client.ClientId)
	ctx.Export("external_client_id", client.ExternalClientId)
	ctx.Export("name", client.Name)
	ctx.Export("app_type", client.AppType)
	ctx.Export("grant_types", client.GrantTypes)
	ctx.Export("callbacks", client.Callbacks)
	ctx.Export("logo_uri", client.LogoUri)
	ctx.Export("jwks_uri", client.JwksUri)
	ctx.Export("third_party_security_mode", client.ThirdPartySecurityMode)
	ctx.Export("external_metadata_created_by", client.ExternalMetadataCreatedBy)
	ctx.Export("validation_valid", validations.ApplyT(func(v []auth0.ClientCimdValidation) bool {
		return summarizeValidation(v).Valid
	}))
	ctx.Export("validation_warnings", validations.ApplyT(func(v []auth0.ClientCimdValidation) []string {
		return summarizeValidation(v).Warnings
	}))
	ctx.Export("validation_violations", validations.ApplyT(func(v []auth0.ClientCimdValidation) []string {
		return summarizeValidation(v).Violations
	}))

	return nil
}
