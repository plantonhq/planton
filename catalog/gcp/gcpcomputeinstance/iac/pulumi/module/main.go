package module

import (
	"github.com/pkg/errors"
	gcpcomputeinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputeinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpComputeInstance
// component.
func Resources(ctx *pulumi.Context, iacInput *gcpcomputeinstancev1alpha1.GcpComputeInstanceIacInput) error {
	locals := initializeLocals(iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdInstance, err := computeInstance(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create compute instance")
	}

	// Semantic outputs — names and shapes byte-identical to the Terraform
	// module's outputs.
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpInstanceId, createdInstance.InstanceId)
	ctx.Export(OpSelfLink, createdInstance.SelfLink)
	ctx.Export(OpStatus, createdInstance.CurrentStatus)
	ctx.Export(OpZone, createdInstance.Zone)
	ctx.Export(OpMachineType, createdInstance.MachineType)
	ctx.Export(OpCpuPlatform, createdInstance.CpuPlatform)

	// Primary internal IP from the first interface; external IP exports
	// "" when the VM has no access config (private VM) rather than
	// failing, so downstream consumers can branch on presence. A single
	// struct-based apply — indexing lazy outputs panics on private VMs
	// whose access-config list is empty.
	ctx.Export(OpInternalIp, createdInstance.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) string {
		if len(nis) > 0 && nis[0].NetworkIp != nil {
			return *nis[0].NetworkIp
		}
		return ""
	}).(pulumi.StringOutput))
	ctx.Export(OpExternalIp, createdInstance.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) string {
		if len(nis) > 0 && len(nis[0].AccessConfigs) > 0 && nis[0].AccessConfigs[0].NatIp != nil {
			return *nis[0].AccessConfigs[0].NatIp
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}
