package module

import (
	"github.com/pkg/errors"
	azuremachinelearningonlineendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningonlineendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazurenativeprovider"
	"github.com/pulumi/pulumi-azure-native-sdk/machinelearningservices/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointIacInput) error {
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

	spec := locals.AzureMachineLearningOnlineEndpoint.Spec

	resourceGroupName, workspaceName, err := parseWorkspaceId(locals.WorkspaceId)
	if err != nil {
		return errors.Wrap(err, "failed to parse the workspace ARM id")
	}

	// The ARM properties object, assembled field-by-field so unset
	// optionals are OMITTED (ARM applies its own defaults) -- the same
	// wire shape the Terraform module's body builds.
	endpointProperties := &machinelearningservices.OnlineEndpointTypeArgs{
		AuthMode: pulumi.String(spec.AuthMode),
	}

	if spec.Description != "" {
		endpointProperties.Description = pulumi.String(spec.Description)
	}

	if spec.PublicNetworkAccessEnabled != nil {
		wire := "Disabled"
		if *spec.PublicNetworkAccessEnabled {
			wire = "Enabled"
		}
		endpointProperties.PublicNetworkAccess = pulumi.String(wire)
	}

	if len(spec.Traffic) > 0 {
		traffic := pulumi.IntMap{}
		for deploymentName, percent := range spec.Traffic {
			traffic[deploymentName] = pulumi.Int(int(percent))
		}
		endpointProperties.Traffic = traffic
	}

	if len(spec.MirrorTraffic) > 0 {
		mirrorTraffic := pulumi.IntMap{}
		for deploymentName, percent := range spec.MirrorTraffic {
			mirrorTraffic[deploymentName] = pulumi.Int(int(percent))
		}
		endpointProperties.MirrorTraffic = mirrorTraffic
	}

	if len(spec.Properties) > 0 {
		endpointProperties.Properties = pulumi.ToStringMap(spec.Properties)
	}

	// Bring-your-own auth keys (Key mode): ARM treats these as
	// create-time input and never returns them on any read (retrieval is
	// the separate listKeys action); the values are secret-wrapped so
	// they never land readable in state.
	if spec.InitialAuthKeys != nil {
		keys := &machinelearningservices.EndpointAuthKeysArgs{}
		if spec.InitialAuthKeys.PrimaryKey.GetValue() != "" {
			keys.PrimaryKey = pulumi.ToSecret(pulumi.String(spec.InitialAuthKeys.PrimaryKey.GetValue())).(pulumi.StringOutput)
		}
		if spec.InitialAuthKeys.SecondaryKey.GetValue() != "" {
			keys.SecondaryKey = pulumi.ToSecret(pulumi.String(spec.InitialAuthKeys.SecondaryKey.GetValue())).(pulumi.StringOutput)
		}
		endpointProperties.Keys = keys
	}

	// Required by the spec (a recorded tightening of ARM's optional): an
	// endpoint without an identity cannot pull images or models and every
	// deployment on it fails at provisioning.
	identityArgs := &machinelearningservices.ManagedServiceIdentityArgs{
		Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
	}
	if len(spec.Identity.IdentityIds) > 0 {
		identityIds := pulumi.StringArray{}
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
		}
		identityArgs.UserAssignedIdentities = identityIds
	}

	// Create the online endpoint -- the stable HTTPS address applications
	// call to score against deployed models, as an ARM child of its
	// workspace. Properties update via full PUT (traffic is the
	// blue/green dial); name, region and workspace replace the endpoint.
	// The endpoint's NAME is reserved region-wide per subscription.
	createdEndpoint, err := machinelearningservices.NewOnlineEndpoint(ctx,
		spec.Name,
		&machinelearningservices.OnlineEndpointArgs{
			EndpointName:             pulumi.String(spec.Name),
			ResourceGroupName:        pulumi.String(resourceGroupName),
			WorkspaceName:            pulumi.String(workspaceName),
			Location:                 pulumi.String(spec.Region),
			Identity:                 identityArgs,
			OnlineEndpointProperties: endpointProperties,
			Tags:                     pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureNativeProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create online endpoint %s", spec.Name)
	}

	ctx.Export(OpOnlineEndpointId, createdEndpoint.ID())
	ctx.Export(OpOnlineEndpointName, createdEndpoint.Name)
	ctx.Export(OpScoringUri, createdEndpoint.OnlineEndpointProperties.ScoringUri())
	ctx.Export(OpSwaggerUri, createdEndpoint.OnlineEndpointProperties.SwaggerUri())
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdEndpoint.Identity.PrincipalId())

	return nil
}
