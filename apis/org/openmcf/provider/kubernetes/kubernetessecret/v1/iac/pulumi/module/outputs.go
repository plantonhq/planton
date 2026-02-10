package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output keys for stack outputs
const (
	OutputSecretName      = "secret_name"
	OutputSecretNamespace = "secret_namespace"
	OutputSecretType      = "secret_type"
)

// exportOutputs exports all stack outputs
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputSecretName, pulumi.String(locals.SecretName))
	ctx.Export(OutputSecretNamespace, pulumi.String(locals.SecretNamespace))
	ctx.Export(OutputSecretType, pulumi.String(locals.SecretType))

	return nil
}
