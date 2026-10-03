package module

import (
	"github.com/pkg/errors"
	azuremachinelearningcomputeclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningcomputecluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/machinelearning"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMachineLearningComputeCluster.Spec

	// Create the Machine Learning compute cluster -- the auto-scaling
	// pool of VMs that training jobs and pipelines run on, as an ARM
	// child of its workspace (.../workspaces/{ws}/computes/{name}).
	//
	// Only identity, scale_settings and tags update in place -- every
	// other argument is ForceNew (the provider's own contract).
	// Uniquely in the ML family, `location` here is the NODES' region
	// and may differ from the workspace's; the provider writes the
	// cluster envelope at the WORKSPACE's region, so ARM reads the
	// envelope back there (recorded on the spec's region field).
	clusterArgs := &machinelearning.ComputeClusterArgs{
		Name:                       pulumi.String(spec.Name),
		MachineLearningWorkspaceId: pulumi.String(locals.WorkspaceId),
		Location:                   pulumi.String(spec.Region),
		VmSize:                     pulumi.String(spec.VmSize),
		// Enum name -> wire value; the spec enum is required, so there
		// is no unspecified fallback.
		VmPriority: pulumi.String(vmPriorityWire[spec.VmPriority]),
		// The one substantive in-place-updatable setting.
		ScaleSettings: &machinelearning.ComputeClusterScaleSettingsArgs{
			MinNodeCount:                    pulumi.Int(int(spec.ScaleSettings.MinNodeCount)),
			MaxNodeCount:                    pulumi.Int(int(spec.ScaleSettings.MaxNodeCount)),
			ScaleDownNodesAfterIdleDuration: pulumi.String(spec.ScaleSettings.ScaleDownNodesAfterIdleDuration),
		},
		// Plain bool: false is the provider's own default, so passing
		// the zero value through is exact.
		SshPublicAccessEnabled: pulumi.Bool(spec.SshPublicAccessEnabled),
		Tags:                   pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Identity != nil {
		identityArgs := &machinelearning.ComputeClusterIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		clusterArgs.Identity = identityArgs
	}

	// The admin account created on every node. At least one credential
	// is set (spec CEL mirrors the provider's AtLeastOneOf). The
	// password is sensitive -- resolved from secret references, masked
	// in state/preview by the provider schema.
	if spec.Ssh != nil {
		sshArgs := &machinelearning.ComputeClusterSshArgs{
			AdminUsername: pulumi.String(spec.Ssh.AdminUsername),
		}
		if spec.Ssh.AdminPassword.GetValue() != "" {
			sshArgs.AdminPassword = pulumi.String(spec.Ssh.AdminPassword.GetValue())
		}
		if spec.Ssh.KeyValue != "" {
			sshArgs.KeyValue = pulumi.String(spec.Ssh.KeyValue)
		}
		clusterArgs.Ssh = sshArgs
	}

	// Optional-with-default-true on the provider: omit when the spec
	// leaves them unset so the provider default applies.
	if spec.LocalAuthEnabled != nil {
		clusterArgs.LocalAuthEnabled = pulumi.Bool(*spec.LocalAuthEnabled)
	}
	if spec.NodePublicIpEnabled != nil {
		clusterArgs.NodePublicIpEnabled = pulumi.Bool(*spec.NodePublicIpEnabled)
	}

	// Optional+Computed on the provider: unset lets Azure network the
	// nodes (a workspace managed network assigns one, read back).
	if spec.SubnetId.GetValue() != "" {
		clusterArgs.SubnetResourceId = pulumi.String(spec.SubnetId.GetValue())
	}

	if spec.Description != "" {
		clusterArgs.Description = pulumi.String(spec.Description)
	}

	createdCluster, err := machinelearning.NewComputeCluster(ctx,
		spec.Name,
		clusterArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create compute cluster %s", spec.Name)
	}

	ctx.Export(OpMachineLearningComputeClusterId, createdCluster.ID())
	ctx.Export(OpMachineLearningComputeClusterName, createdCluster.Name)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdCluster.Identity.PrincipalId())

	return nil
}
