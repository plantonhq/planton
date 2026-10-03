package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesPersistentVolumeClaimOutputs field names.
const (
	OutputPvcName        = "pvc_name"
	OutputNamespace      = "namespace"
	OutputStorageRequest = "storage_request"
)

// exportOutputs exports the outputs from the created PersistentVolumeClaim.
// Deliberately no bind-time status: a claim under a wait_for_first_consumer
// class is correctly Pending until consumed.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputPvcName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputStorageRequest, pulumi.String(locals.Spec.GetStorageRequest()))

	return nil
}
