package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants - one per KubernetesLangfuseOutputs field (nested
// outputs use dotted names).
const (
	OpNamespace                  = "namespace"
	OpService                    = "service"
	OpEndpoint                   = "endpoint"
	OpExternalUrl                = "external_url"
	OpOtelEndpoint               = "otel_endpoint"
	OpPortForwardCommand         = "port_forward_command"
	OpInstanceSecretName         = "instance_secret_name"
	OpOrganizationId             = "organization_id"
	OpProjectId                  = "project_id"
	OpProjectPublicKey           = "project_public_key"
	OpProjectSecretKeySecretName = "project_secret_key_secret.name"
	OpProjectSecretKeySecretKey  = "project_secret_key_secret.key"
	OpAdminEmail                 = "admin_email"
	OpAdminPasswordSecretName    = "admin_password_secret.name"
	OpAdminPasswordSecretKey     = "admin_password_secret.key"
)

// exportOutputs publishes the composition handles (Terraform twin:
// outputs.tf). The seeding outputs are empty when seeding is off, and the
// admin outputs when no admin is seeded.
func exportOutputs(ctx *pulumi.Context, locals *Locals, projectPublicKey pulumi.StringOutput) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.WebServiceName))
	ctx.Export(OpEndpoint, pulumi.String(locals.Endpoint))
	ctx.Export(OpExternalUrl, pulumi.String(locals.ExternalUrl))
	ctx.Export(OpOtelEndpoint, pulumi.String(locals.OtelEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	ctx.Export(OpInstanceSecretName, pulumi.String(locals.InstanceSecretName))

	if locals.InitEnabled {
		ctx.Export(OpOrganizationId, pulumi.String(locals.OrganizationId))
		ctx.Export(OpProjectId, pulumi.String(locals.ProjectId))
		ctx.Export(OpProjectPublicKey, projectPublicKey)
		ctx.Export(OpProjectSecretKeySecretName, pulumi.String(locals.InstanceSecretName))
		ctx.Export(OpProjectSecretKeySecretKey, pulumi.String(keyInitProjectSecretKey))
	} else {
		ctx.Export(OpOrganizationId, pulumi.String(""))
		ctx.Export(OpProjectId, pulumi.String(""))
		ctx.Export(OpProjectPublicKey, pulumi.String(""))
		ctx.Export(OpProjectSecretKeySecretName, pulumi.String(""))
		ctx.Export(OpProjectSecretKeySecretKey, pulumi.String(""))
	}

	if locals.AdminSeeded {
		ctx.Export(OpAdminEmail, pulumi.String(locals.AdminEmail))
		ctx.Export(OpAdminPasswordSecretName, pulumi.String(locals.InstanceSecretName))
		ctx.Export(OpAdminPasswordSecretKey, pulumi.String(keyInitUserPassword))
	} else {
		ctx.Export(OpAdminEmail, pulumi.String(""))
		ctx.Export(OpAdminPasswordSecretName, pulumi.String(""))
		ctx.Export(OpAdminPasswordSecretKey, pulumi.String(""))
	}
}
