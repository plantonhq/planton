package module

import (
	"github.com/pkg/errors"
	gcpcomputeinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcomputeinstance/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpComputeInstance component.
func Resources(ctx *pulumi.Context, stackInput *gcpcomputeinstancev1.GcpComputeInstanceStackInput) error {
	locals := initializeLocals(stackInput)

	// Set up the GCP provider from the supplied credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Create the Compute Engine instance.
	createdInstance, err := computeInstance(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create compute instance")
	}

	// Export stack outputs
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpInstanceId, createdInstance.InstanceId)
	ctx.Export(OpSelfLink, createdInstance.SelfLink)
	ctx.Export(OpZone, createdInstance.Zone)
	ctx.Export(OpMachineType, createdInstance.MachineType)
	ctx.Export(OpCpuPlatform, createdInstance.CpuPlatform)

	// Export network information from the first network interface
	ctx.Export(OpInternalIp, createdInstance.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
	ctx.Export(OpExternalIp, createdInstance.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())

	return nil
}
