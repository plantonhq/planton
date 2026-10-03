package module

import (
	"github.com/pkg/errors"
	azuremachinelearningbatchendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningbatchendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazurenativeprovider"
	"github.com/pulumi/pulumi-azure-native-sdk/machinelearningservices/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningbatchendpointv1alpha1.AzureMachineLearningBatchEndpointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the azure-native provider from the IaC input via the shared
	// builder, which resolves the right credential mechanism (static client
	// secret, keyless web identity, or ambient chain). This kind rides
	// azure-native, not the classic provider: azurerm/classic carry NO ML
	// endpoint resources, and azure-native's typed resources are generated
	// from the same ARM specification the Terraform module's raw-API shape
	// pins (tracked at hashicorp/terraform-provider-azurerm#32011 with a
	// mandatory move to native resources when azurerm ships them).
	azureNativeProvider, err := pulumiazurenativeprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure-native provider")
	}

	spec := locals.AzureMachineLearningBatchEndpoint.Spec

	resourceGroupName, workspaceName, err := parseWorkspaceId(locals.WorkspaceId)
	if err != nil {
		return errors.Wrap(err, "failed to parse the workspace ARM id")
	}

	// The ARM properties object, assembled field-by-field so unset
	// optionals are OMITTED (ARM applies its own defaults) -- the same
	// wire shape the Terraform module's body builds. authMode always
	// sends: ARM requires it, and AADToken is the only mode the batch
	// service accepts (the spec's vocabulary already enforces it; the
	// unset default applies here). There is NO keys arm (unlike the
	// online endpoint): with Key mode rejected by the service, ARM's
	// create-time keys property is dead surface for this kind.
	authMode := spec.AuthMode
	if authMode == "" {
		authMode = "AADToken"
	}
	endpointProperties := &machinelearningservices.BatchEndpointTypeArgs{
		AuthMode: pulumi.String(authMode),
	}

	if spec.DefaultDeploymentName != "" {
		endpointProperties.Defaults = &machinelearningservices.BatchEndpointDefaultsArgs{
			DeploymentName: pulumi.String(spec.DefaultDeploymentName),
		}
	}

	if spec.Description != "" {
		endpointProperties.Description = pulumi.String(spec.Description)
	}

	if spec.Properties != nil && len(spec.Properties) > 0 {
		endpointProperties.Properties = pulumi.ToStringMap(spec.Properties)
	}

	// Optional here where the online endpoint requires one: batch jobs
	// run under the INVOKER's Entra token plus the COMPUTE pool's managed
	// identity, so the endpoint's own identity sits outside the batch
	// data path.
	var identityArgs *machinelearningservices.ManagedServiceIdentityArgs
	if spec.Identity != nil {
		identityArgs = &machinelearningservices.ManagedServiceIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.UserAssignedIdentities = identityIds
		}
	}

	// Create the batch endpoint -- the stable address batch scoring jobs
	// are submitted to, as an ARM child of its workspace. Properties
	// update via full PUT (the default-deployment pointer is the routing
	// dial); name, region and workspace replace the endpoint. Nothing
	// runs or bills while no job is active.
	createdEndpoint, err := machinelearningservices.NewBatchEndpoint(ctx,
		spec.Name,
		&machinelearningservices.BatchEndpointArgs{
			EndpointName:            pulumi.String(spec.Name),
			ResourceGroupName:       pulumi.String(resourceGroupName),
			WorkspaceName:           pulumi.String(workspaceName),
			Location:                pulumi.String(spec.Region),
			Identity:                identityArgs,
			BatchEndpointProperties: endpointProperties,
			Tags:                    pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureNativeProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create batch endpoint %s", spec.Name)
	}

	ctx.Export(OpBatchEndpointId, createdEndpoint.ID())
	ctx.Export(OpBatchEndpointName, createdEndpoint.Name)
	ctx.Export(OpScoringUri, createdEndpoint.BatchEndpointProperties.ScoringUri())
	ctx.Export(OpSwaggerUri, createdEndpoint.BatchEndpointProperties.SwaggerUri())
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdEndpoint.Identity.PrincipalId())

	return nil
}
