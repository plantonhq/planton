package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesKeycloakOperatorOutputs
// field.
const (
	OpNamespace  = "namespace"
	OpDeployment = "deployment"
	OpService    = "service"
)

// exportOutputs publishes the bundle's fixed handles (the final apply
// group is passed so the export carries its dependency). Deployment and
// Service names are upstream-fixed: `keycloak-operator`.
func exportOutputs(ctx *pulumi.Context, locals *Locals, _ pulumi.Resource) error {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpDeployment, pulumi.String(vars.DeploymentName))
	ctx.Export(OpService, pulumi.String(vars.ServiceName))
	return nil
}
