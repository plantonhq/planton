package module

import (
	"github.com/pkg/errors"
	azuremachinelearningcomputeinstancev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningcomputeinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/machinelearning"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningcomputeinstancev1alpha1.AzureMachineLearningComputeInstanceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMachineLearningComputeInstance.Spec

	// Create the Machine Learning compute instance -- a single always-on
	// VM serving as one data scientist's cloud workstation, as an ARM
	// child of its workspace (.../workspaces/{ws}/computes/{name}).
	//
	// The provider has NO update path for this resource: EVERY argument
	// is ForceNew, tags included -- any change replaces the instance
	// (its OS disk and local files go with it). The instance always
	// runs in its workspace's region (the service's own rule; there is
	// no location argument), and its name is reserved region-wide per
	// subscription.
	//
	// One contract lives at apply time, not manifest time: when
	// node_public_ip_enabled is false, the provider requires
	// subnet_resource_id UNLESS the workspace runs a managed network --
	// it depends on the live workspace's isolation mode (recorded on
	// the spec fields).
	instanceArgs := &machinelearning.ComputeInstanceArgs{
		Name:                       pulumi.String(spec.Name),
		MachineLearningWorkspaceId: pulumi.String(locals.WorkspaceId),
		VirtualMachineSize:         pulumi.String(spec.VirtualMachineSize),
		Tags:                       pulumi.ToStringMap(locals.AzureTags),
	}

	// "personal" is the only value the provider accepts; unset omits
	// the property and leaves the service default.
	if spec.AuthorizationType != "" {
		instanceArgs.AuthorizationType = pulumi.String(spec.AuthorizationType)
	}

	// The admin-provisions-for-the-team pattern: assign the instance to
	// a user other than the deploying principal.
	if spec.AssignToUser != nil {
		assignArgs := &machinelearning.ComputeInstanceAssignToUserArgs{}
		if spec.AssignToUser.TenantId != "" {
			assignArgs.TenantId = pulumi.String(spec.AssignToUser.TenantId)
		}
		if spec.AssignToUser.ObjectId != "" {
			assignArgs.ObjectId = pulumi.String(spec.AssignToUser.ObjectId)
		}
		instanceArgs.AssignToUser = assignArgs
	}

	if spec.Identity != nil {
		identityArgs := &machinelearning.ComputeInstanceIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		instanceArgs.Identity = identityArgs
	}

	// Optional-with-default-true on the provider: omit when the spec
	// leaves them unset so the provider default applies.
	if spec.LocalAuthEnabled != nil {
		instanceArgs.LocalAuthEnabled = pulumi.Bool(*spec.LocalAuthEnabled)
	}
	if spec.NodePublicIpEnabled != nil {
		instanceArgs.NodePublicIpEnabled = pulumi.Bool(*spec.NodePublicIpEnabled)
	}

	// Absent block means the SSH port is DISABLED (the provider's own
	// contract); the service assigns the username and port, surfaced as
	// outputs.
	if spec.Ssh != nil {
		instanceArgs.Ssh = &machinelearning.ComputeInstanceSshArgs{
			PublicKey: pulumi.String(spec.Ssh.PublicKey),
		}
	}

	// Only legal when the workspace does NOT use a managed network.
	if spec.SubnetId.GetValue() != "" {
		instanceArgs.SubnetResourceId = pulumi.String(spec.SubnetId.GetValue())
	}

	if spec.Description != "" {
		instanceArgs.Description = pulumi.String(spec.Description)
	}

	createdInstance, err := machinelearning.NewComputeInstance(ctx,
		spec.Name,
		instanceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create compute instance %s", spec.Name)
	}

	ctx.Export(OpMachineLearningComputeInstanceId, createdInstance.ID())
	ctx.Export(OpMachineLearningComputeInstanceName, createdInstance.Name)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdInstance.Identity.PrincipalId())
	// The ssh outputs are populated only when the ssh block is configured.
	// Elem() dereferences nil to the zero value ("" / 0), mirroring the
	// Terraform module's try(..., "") / try(..., 0) fallbacks -- both
	// engines must emit the same output shape, and a raw nil export fails
	// the harness's int32 parse.
	ctx.Export(OpSshUsername, createdInstance.Ssh.Username().Elem())
	ctx.Export(OpSshPort, createdInstance.Ssh.Port().Elem())

	return nil
}
