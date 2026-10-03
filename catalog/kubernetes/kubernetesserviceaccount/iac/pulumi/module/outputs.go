// Exports outputs; keys mirror KubernetesServiceAccountOutputs field names.
package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output keys for outputs
const (
	OutputServiceAccountName     = "service_account_name"
	OutputNamespace              = "namespace"
	OutputRbacSubject            = "rbac_subject"
	OutputWorkloadIdentityHandle = "workload_identity_handle"
)

// exportOutputs exports all outputs
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputServiceAccountName, pulumi.String(locals.ServiceAccountName))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputRbacSubject, pulumi.String(locals.RbacSubject))
	ctx.Export(OutputWorkloadIdentityHandle, pulumi.String(locals.WorkloadIdentityHandle))

	return nil
}
