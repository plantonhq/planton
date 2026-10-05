package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — the outputs contract
// (outputs.proto). The Terraform twin exports the same handles.
const (
	OpNamespace           = "namespace"
	OpExposeService       = "expose_service"
	OpKubeEndpoint        = "kube_endpoint"
	OpExternalUrl         = "external_url"
	OpCoreService         = "core_service"
	OpPortalService       = "portal_service"
	OpRegistryService     = "registry_service"
	OpJobserviceService   = "jobservice_service"
	OpTrivyService        = "trivy_service"
	OpDatabaseService     = "database_service"
	OpRedisService        = "redis_service"
	OpAdminUsername       = "admin_username"
	OpAdminPasswordSecret = "admin_password_secret"
	OpPortForwardCommand  = "port_forward_command"
)

// exportOutputs publishes the composition handles. Component Service
// names are chart-derived (`<fullname>-<component>` with the fullname
// pinned to metadata.name); the front-door Service name is pinned to
// metadata.name by the module.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpExposeService, pulumi.String(locals.ReleaseName))
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.kubeEndpoint()))
	ctx.Export(OpExternalUrl, pulumi.String(locals.Spec.GetExternalUrl()))
	ctx.Export(OpCoreService, pulumi.String(locals.ReleaseName+"-core"))
	ctx.Export(OpPortalService, pulumi.String(locals.ReleaseName+"-portal"))
	ctx.Export(OpRegistryService, pulumi.String(locals.ReleaseName+"-registry"))
	ctx.Export(OpJobserviceService, pulumi.String(locals.ReleaseName+"-jobservice"))

	trivyService := ""
	if locals.TrivyEnabled {
		trivyService = locals.ReleaseName + "-trivy"
	}
	ctx.Export(OpTrivyService, pulumi.String(trivyService))

	databaseService := ""
	if locals.InternalDatabase {
		databaseService = locals.ReleaseName + "-database"
	}
	ctx.Export(OpDatabaseService, pulumi.String(databaseService))

	redisService := ""
	if locals.InternalRedis {
		redisService = locals.ReleaseName + "-redis"
	}
	ctx.Export(OpRedisService, pulumi.String(redisService))

	ctx.Export(OpAdminUsername, pulumi.String(vars.AdminUsername))
	ctx.Export(OpAdminPasswordSecret, pulumi.StringMap{
		"name": pulumi.String(locals.AdminSecretName),
		"key":  pulumi.String(locals.AdminSecretKey),
	})
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.portForwardCommand()))
}
