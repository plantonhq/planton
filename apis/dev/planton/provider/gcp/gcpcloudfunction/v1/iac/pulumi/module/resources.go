package module

import (
	"github.com/pkg/errors"
	gcpcloudfunctionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudfunction/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudFunction component.
func Resources(ctx *pulumi.Context, stackInput *gcpcloudfunctionv1.GcpCloudFunctionStackInput) error {
	// Initialize locals
	locals := initializeLocals(stackInput)

	// Create GCP provider using the credentials from the input
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Create the Cloud Function
	function, err := createCloudFunction(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud function")
	}

	// Export stack outputs
	ctx.Export(OpFunctionId, function.ID())

	// Function URL is only available for HTTP triggers
	if locals.IsHttpTrigger {
		ctx.Export(OpFunctionUrl, function.ServiceConfig.ApplyT(func(sc interface{}) string {
			if sc == nil {
				return ""
			}
			serviceConfig := sc.(map[string]interface{})
			if uri, ok := serviceConfig["uri"].(string); ok {
				return uri
			}
			return ""
		}))
	} else {
		ctx.Export(OpFunctionUrl, pulumi.String(""))
	}

	ctx.Export(OpServiceAccountEmail, function.ServiceConfig.ApplyT(func(sc interface{}) string {
		if sc == nil {
			return ""
		}
		serviceConfig := sc.(map[string]interface{})
		if email, ok := serviceConfig["serviceAccountEmail"].(string); ok {
			return email
		}
		return ""
	}))

	ctx.Export(OpState, function.State)
	ctx.Export(OpCloudRunServiceId, function.Name)

	// Eventarc trigger ID is only available for event-driven functions
	if !locals.IsHttpTrigger {
		ctx.Export(OpEventarcTriggerId, function.EventTrigger.ApplyT(func(et interface{}) string {
			if et == nil {
				return ""
			}
			eventTrigger := et.(map[string]interface{})
			if triggerId, ok := eventTrigger["trigger"].(string); ok {
				return triggerId
			}
			return ""
		}))
	} else {
		ctx.Export(OpEventarcTriggerId, pulumi.String(""))
	}

	return nil
}
