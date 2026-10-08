package module

import (
	"github.com/pkg/errors"
	gcpcloudfunctionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudfunctionsv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudFunction
// component.
func Resources(ctx *pulumi.Context, iacInput *gcpcloudfunctionv1alpha1.GcpCloudFunctionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdFunction, err := function(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud function")
	}

	ctx.Export(OpFunctionId, createdFunction.ID())
	ctx.Export(OpFunctionUrl, createdFunction.Url)
	ctx.Export(OpState, createdFunction.State)
	ctx.Export(OpName, createdFunction.Name)
	ctx.Export(OpEnvironment, createdFunction.Environment)
	ctx.Export(OpUpdateTime, createdFunction.UpdateTime)

	// service_config is an optional block: resolve its attributes through
	// ApplyT so a function deployed without one still exports empty strings
	// instead of failing.
	ctx.Export(OpServiceAccountEmail, createdFunction.ServiceConfig.ApplyT(func(sc *cloudfunctionsv2.FunctionServiceConfig) string {
		if sc == nil || sc.ServiceAccountEmail == nil {
			return ""
		}
		return *sc.ServiceAccountEmail
	}))
	ctx.Export(OpCloudRunServiceId, createdFunction.ServiceConfig.ApplyT(func(sc *cloudfunctionsv2.FunctionServiceConfig) string {
		if sc == nil || sc.Service == nil {
			return ""
		}
		return *sc.Service
	}))
	ctx.Export(OpUri, createdFunction.ServiceConfig.ApplyT(func(sc *cloudfunctionsv2.FunctionServiceConfig) string {
		if sc == nil || sc.Uri == nil {
			return ""
		}
		return *sc.Uri
	}))
	ctx.Export(OpEventarcTriggerId, createdFunction.EventTrigger.ApplyT(func(et *cloudfunctionsv2.FunctionEventTrigger) string {
		if et == nil || et.Trigger == nil {
			return ""
		}
		return *et.Trigger
	}))

	return nil
}
