package module

import (
	gcpiamoauthclientv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpiamoauthclient/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpIamOauthClient *gcpiamoauthclientv1alpha1.GcpIamOauthClient

	// The cloud-side client ID defaults to metadata.name when the spec
	// leaves oauth_client_id empty — the same naming basis every kind
	// uses.
	OauthClientId string

	// The client's location defaults to "global" — the documented home
	// for workforce OAuth clients (the spec comment records the
	// contract).
	Location string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpiamoauthclientv1alpha1.GcpIamOauthClientIacInput) *Locals {
	target := iacInput.Target

	oauthClientId := target.Spec.OauthClientId
	if oauthClientId == "" {
		oauthClientId = target.Metadata.Name
	}

	location := target.Spec.Location
	if location == "" {
		location = "global"
	}

	return &Locals{
		GcpIamOauthClient: target,
		OauthClientId:     oauthClientId,
		Location:          location,
	}
}
