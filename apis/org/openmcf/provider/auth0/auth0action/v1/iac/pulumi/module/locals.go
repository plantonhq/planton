package module

import (
	auth0actionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0action/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Auth0Action *auth0actionv1.Auth0Action

	ActionName string
	Code       string
	Runtime    string
	Deploy     bool

	SupportedTrigger *auth0actionv1.Auth0ActionSupportedTrigger
	Dependencies     []*auth0actionv1.Auth0ActionDependency
	Secrets          []*auth0actionv1.Auth0ActionSecret
	TriggerBinding   *auth0actionv1.Auth0ActionTriggerBinding
}

func initializeLocals(ctx *pulumi.Context, stackInput *auth0actionv1.Auth0ActionStackInput) *Locals {
	locals := &Locals{}

	locals.Auth0Action = stackInput.Target

	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	locals.ActionName = metadata.Name
	locals.Code = spec.Code
	locals.Runtime = spec.Runtime
	locals.Deploy = spec.Deploy

	locals.SupportedTrigger = spec.SupportedTrigger
	locals.Dependencies = spec.Dependencies
	locals.Secrets = spec.Secrets
	locals.TriggerBinding = spec.TriggerBinding

	return locals
}
