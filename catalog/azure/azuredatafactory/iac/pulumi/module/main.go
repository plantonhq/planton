package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuredatafactoryv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactory/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens. Like the Event Grid namespace, a factory supports
// the combined mode -- the third token carries both flavors.
var identityTypeStrings = map[azuredatafactoryv1alpha1.AzureDataFactoryIdentityType]string{
	azuredatafactoryv1alpha1.AzureDataFactoryIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuredatafactoryv1alpha1.AzureDataFactoryIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuredatafactoryv1alpha1.AzureDataFactoryIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azuredatafactoryv1alpha1.AzureDataFactoryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactory.Spec

	// Platform default true -- always sent (mirrors Azure's own
	// default; the provider maps the bool onto ARM's Enabled/Disabled
	// tokens itself). (Proto defaults applied here, matching the TF
	// module's coalesce.)
	publicNetworkEnabled := true
	if spec.PublicNetworkEnabled != nil {
		publicNetworkEnabled = *spec.PublicNetworkEnabled
	}

	// Enabling is an in-place update (the provider creates the managed
	// network named "default" after the factory); DISABLING it is
	// ForceNew (the provider's one CustomizeDiff on this resource) --
	// documented on the spec field. False and omitted are the same wire
	// shape, so the platform default is sent for an explicit plan.
	managedVirtualNetworkEnabled := false
	if spec.ManagedVirtualNetworkEnabled != nil {
		managedVirtualNetworkEnabled = *spec.ManagedVirtualNetworkEnabled
	}

	factoryArgs := &datafactory.FactoryArgs{
		Name:                         pulumi.String(spec.Name),
		ResourceGroupName:            pulumi.String(spec.ResourceGroup.GetValue()),
		Location:                     pulumi.String(spec.Region),
		PublicNetworkEnabled:         pulumi.Bool(publicNetworkEnabled),
		ManagedVirtualNetworkEnabled: pulumi.Bool(managedVirtualNetworkEnabled),
		Tags:                         pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.PurviewId.GetValue() != "" {
		factoryArgs.PurviewId = pulumi.String(spec.PurviewId.GetValue())
	}

	// Customer-managed-key encryption composed onto the factory's own
	// inline fields (the provider's standalone CMK resource writes the
	// same encryption object; it demands a VERSIONED key where these
	// inline fields accept versionless too -- the spec prefers
	// versionless so rotation propagates). The unwrap identity is
	// required with the key (spec makes it required in the block,
	// front-loading the provider's create-time CustomizeDiff).
	if spec.CustomerManagedKey != nil {
		factoryArgs.CustomerManagedKeyId = pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue())
		factoryArgs.CustomerManagedKeyIdentityId = pulumi.String(spec.CustomerManagedKey.UserAssignedIdentityId.GetValue())
	}

	if spec.Identity != nil {
		identityArgs := &datafactory.FactoryIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		factoryArgs.Identity = identityArgs
	}

	// The repository binding (at most one -- spec CEL mirrors the
	// provider's ConflictsWith) travels through a separate
	// configure-repo call the provider makes AFTER the factory exists,
	// and REMOVING the block does not detach the repository (the
	// provider calls no repo-clear API) -- documented on the spec
	// fields. Platform default: publishing enabled (Azure stores the
	// inverse disablePublish flag; the provider translates).
	if spec.GithubConfiguration != nil {
		githubPublishingEnabled := true
		if spec.GithubConfiguration.PublishingEnabled != nil {
			githubPublishingEnabled = *spec.GithubConfiguration.PublishingEnabled
		}
		githubArgs := &datafactory.FactoryGithubConfigurationArgs{
			AccountName:       pulumi.String(spec.GithubConfiguration.AccountName),
			BranchName:        pulumi.String(spec.GithubConfiguration.BranchName),
			RepositoryName:    pulumi.String(spec.GithubConfiguration.RepositoryName),
			RootFolder:        pulumi.String(spec.GithubConfiguration.RootFolder),
			PublishingEnabled: pulumi.BoolPtr(githubPublishingEnabled),
		}
		if spec.GithubConfiguration.GitUrl != "" {
			githubArgs.GitUrl = pulumi.String(spec.GithubConfiguration.GitUrl)
		}
		factoryArgs.GithubConfiguration = githubArgs
	}

	if spec.VstsConfiguration != nil {
		vstsPublishingEnabled := true
		if spec.VstsConfiguration.PublishingEnabled != nil {
			vstsPublishingEnabled = *spec.VstsConfiguration.PublishingEnabled
		}
		factoryArgs.VstsConfiguration = &datafactory.FactoryVstsConfigurationArgs{
			AccountName:       pulumi.String(spec.VstsConfiguration.AccountName),
			BranchName:        pulumi.String(spec.VstsConfiguration.BranchName),
			ProjectName:       pulumi.String(spec.VstsConfiguration.ProjectName),
			RepositoryName:    pulumi.String(spec.VstsConfiguration.RepositoryName),
			RootFolder:        pulumi.String(spec.VstsConfiguration.RootFolder),
			TenantId:          pulumi.String(spec.VstsConfiguration.TenantId),
			PublishingEnabled: pulumi.BoolPtr(vstsPublishingEnabled),
		}
	}

	// Workspace-wide parameters (names unique -- spec CEL front-loads
	// the provider's own duplicate check). Array/Object values travel
	// as JSON text; Azure stores the typed value.
	if len(spec.GlobalParameters) > 0 {
		globalParameters := datafactory.FactoryGlobalParameterArray{}
		for _, globalParameter := range spec.GlobalParameters {
			globalParameters = append(globalParameters, &datafactory.FactoryGlobalParameterArgs{
				Name:  pulumi.String(globalParameter.Name),
				Type:  pulumi.String(globalParameter.Type),
				Value: pulumi.String(globalParameter.Value),
			})
		}
		factoryArgs.GlobalParameters = globalParameters
	}

	createdFactory, err := datafactory.NewFactory(ctx,
		locals.AzureDataFactory.Metadata.Name,
		factoryArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data factory %s",
			locals.AzureDataFactory.Metadata.Name)
	}

	// Composed credentials -- one provider resource per named entry,
	// keyed by name (renames replace only that credential, siblings
	// stay untouched), in lockstep with the TF module's for_each.
	// Credential names share one namespace under the factory (spec CEL
	// enforces cross-list uniqueness), so one map output carries both
	// flavors.
	credentialIds := pulumi.Map{}
	for _, credential := range spec.UserManagedIdentityCredentials {
		credentialArgs := &datafactory.CredentialUserManagedIdentityArgs{
			Name:          pulumi.String(credential.Name),
			DataFactoryId: createdFactory.ID(),
			IdentityId:    pulumi.String(credential.IdentityId.GetValue()),
		}
		if credential.Description != "" {
			credentialArgs.Description = pulumi.String(credential.Description)
		}
		if len(credential.Annotations) > 0 {
			credentialArgs.Annotations = pulumi.ToStringArray(credential.Annotations)
		}
		createdCredential, err := datafactory.NewCredentialUserManagedIdentity(ctx,
			fmt.Sprintf("%s-%s", locals.AzureDataFactory.Metadata.Name, credential.Name),
			credentialArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdFactory))
		if err != nil {
			return errors.Wrapf(err, "failed to create user-managed-identity credential %s", credential.Name)
		}
		credentialIds[credential.Name] = createdCredential.ID()
	}

	// The service-principal flavor: the principal's key is resolved
	// through a Key Vault LINKED SERVICE by name -- the secret itself
	// never travels through this module.
	for _, credential := range spec.ServicePrincipalCredentials {
		credentialArgs := &datafactory.CredentialServicePrincipalArgs{
			Name:               pulumi.String(credential.Name),
			DataFactoryId:      createdFactory.ID(),
			TenantId:           pulumi.String(credential.TenantId),
			ServicePrincipalId: pulumi.String(credential.ServicePrincipalId),
		}
		if credential.ServicePrincipalKey != nil {
			keyArgs := &datafactory.CredentialServicePrincipalServicePrincipalKeyArgs{
				LinkedServiceName: pulumi.String(credential.ServicePrincipalKey.LinkedServiceName),
				SecretName:        pulumi.String(credential.ServicePrincipalKey.SecretName),
			}
			if credential.ServicePrincipalKey.SecretVersion != "" {
				keyArgs.SecretVersion = pulumi.String(credential.ServicePrincipalKey.SecretVersion)
			}
			credentialArgs.ServicePrincipalKey = keyArgs
		}
		if credential.Description != "" {
			credentialArgs.Description = pulumi.String(credential.Description)
		}
		if len(credential.Annotations) > 0 {
			credentialArgs.Annotations = pulumi.ToStringArray(credential.Annotations)
		}
		createdCredential, err := datafactory.NewCredentialServicePrincipal(ctx,
			fmt.Sprintf("%s-%s", locals.AzureDataFactory.Metadata.Name, credential.Name),
			credentialArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdFactory))
		if err != nil {
			return errors.Wrapf(err, "failed to create service-principal credential %s", credential.Name)
		}
		credentialIds[credential.Name] = createdCredential.ID()
	}

	// Composed managed private endpoints -- private egress from the
	// factory's managed virtual network (spec CEL requires the
	// network). Each endpoint is CREATE-ONLY in the provider (no
	// Update method): every field change replaces that endpoint,
	// siblings stay untouched. Exactly one arm is set per entry (spec
	// CEL): subresource_name for regular ARM targets, fqdns for
	// Private Link Service targets. The TARGET side must approve the
	// endpoint's connection before traffic flows -- approval happens
	// outside this resource, and the endpoint provisions to Succeeded
	// while the connection is still Pending.
	managedPrivateEndpointIds := pulumi.Map{}
	for _, endpoint := range spec.ManagedPrivateEndpoints {
		endpointArgs := &datafactory.ManagedPrivateEndpointArgs{
			Name:             pulumi.String(endpoint.Name),
			DataFactoryId:    createdFactory.ID(),
			TargetResourceId: pulumi.String(endpoint.TargetResourceId.GetValue()),
		}
		if endpoint.SubresourceName != "" {
			endpointArgs.SubresourceName = pulumi.String(endpoint.SubresourceName)
		}
		if len(endpoint.Fqdns) > 0 {
			endpointArgs.Fqdns = pulumi.ToStringArray(endpoint.Fqdns)
		}
		createdEndpoint, err := datafactory.NewManagedPrivateEndpoint(ctx,
			fmt.Sprintf("%s-%s", locals.AzureDataFactory.Metadata.Name, endpoint.Name),
			endpointArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdFactory))
		if err != nil {
			return errors.Wrapf(err, "failed to create managed private endpoint %s", endpoint.Name)
		}
		managedPrivateEndpointIds[endpoint.Name] = createdEndpoint.ID()
	}

	ctx.Export(OpDataFactoryId, createdFactory.ID())
	ctx.Export(OpDataFactoryName, createdFactory.Name)
	// Empty unless a system-assigned flavor is enabled -- mirrors the
	// TF module's try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdFactory.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))
	ctx.Export(OpCredentialIds, credentialIds)
	ctx.Export(OpManagedPrivateEndpointIds, managedPrivateEndpointIds)

	return nil
}
