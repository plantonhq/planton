package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output keys for outputs. Keys match outputs.proto field names.
const (
	OutputRoleName    = "role_name"
	OutputRoleKind    = "role_kind"
	OutputBindingName = "binding_name"
	OutputBindingKind = "binding_kind"
	OutputNamespace   = "namespace"
)

// exportOutputs exports all outputs: the names and kinds of the objects in
// the grant (binding fields are empty when no binding was created, namespace is
// empty for cluster-scoped grants).
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputRoleName, pulumi.String(locals.RoleName))
	ctx.Export(OutputRoleKind, pulumi.String(locals.RoleKind))
	ctx.Export(OutputBindingName, pulumi.String(locals.BindingName))
	ctx.Export(OutputBindingKind, pulumi.String(locals.BindingKind))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))

	return nil
}
