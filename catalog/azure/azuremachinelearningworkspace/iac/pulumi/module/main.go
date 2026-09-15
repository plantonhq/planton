package module

import (
	"github.com/pkg/errors"
	azuremachinelearningworkspacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningworkspace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/machinelearning"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient
	// chain). The machine_learning features flag makes destroy purge the soft-delete ghost
	// that would otherwise keep holding the workspace NAME (the provider default leaves it);
	// a soft-delete recovery window is not part of this module's contract -- mirrors the
	// Terraform module's provider features block.
	azureProvider, err := pulumiazureprovider.GetWithFeatures(ctx, stackInput.ProviderConfig,
		azure.ProviderFeaturesArgs{
			MachineLearning: azure.ProviderFeaturesMachineLearningArgs{
				PurgeSoftDeletedWorkspaceOnDestroy: pulumi.Bool(true),
			},
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMachineLearningWorkspace.Spec

	// PARITY-EXCEPTION: the classic Pulumi SDK does not expose the
	// workspace's storage_account_access_type (azurerm v5's
	// SystemDatastoresAuthMode surface), so a manifest that sets it
	// deploys via the Terraform engine only. Failing loudly here beats
	// silently keeping the account-key auth mode the user turned off.
	if spec.StorageAccountAccessType != azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceStorageAccountAccessType_azure_machine_learning_workspace_storage_account_access_type_unspecified {
		return errors.New("the Pulumi engine cannot express storage_account_access_type " +
			"-- deploy this workspace with the Terraform engine")
	}

	// Create the Azure Machine Learning workspace -- the central home
	// a data-science team works in, a thin coordination object over
	// three REQUIRED companion services (storage, key vault,
	// application insights) and an optional container registry; all
	// four attachments are ForceNew. The spec's CEL contracts already
	// enforce the provider's code-level rules. Deletion is a SOFT
	// delete at the service: the provider purges the name-holding ghost
	// because this module enables the machine_learning features flag
	// at provider construction above.
	workspaceArgs := &machinelearning.WorkspaceArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The three required companion services (all ForceNew).
		ApplicationInsightsId: pulumi.String(spec.ApplicationInsightsId.GetValue()),
		KeyVaultId:            pulumi.String(spec.KeyVaultId.GetValue()),
		StorageAccountId:      pulumi.String(spec.StorageAccountId.GetValue()),
		HighBusinessImpact:    pulumi.Bool(spec.HighBusinessImpact),
		V1LegacyModeEnabled:   pulumi.Bool(spec.V1LegacyModeEnabled),
		Tags:                  pulumi.ToStringMap(locals.AzureTags),
	}

	// service_side_encryption_enabled is SENT ONLY WHEN TRUE: the
	// provider pairs it with the encryption block via RequiredWith,
	// which fires on any SPECIFIED value -- an explicit false without
	// the block is rejected at validation ("all of
	// encryption,service_side_encryption_enabled must be specified" --
	// live-caught). Omitting it applies the provider default (false),
	// and the spec CEL guarantees the encryption block is present
	// whenever this is true. ForceNew.
	if spec.ServiceSideEncryptionEnabled {
		workspaceArgs.ServiceSideEncryptionEnabled = pulumi.Bool(true)
	}

	if spec.Identity != nil {
		identityArgs := &machinelearning.WorkspaceIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		workspaceArgs.Identity = identityArgs
	}

	// Enum name -> wire value; unspecified omits the property so the
	// provider applies its default, "Default". FEATURE_STORE requires
	// the feature_store block (spec CEL, both directions).
	if kind, ok := kindWire[spec.Kind]; ok {
		workspaceArgs.Kind = pulumi.String(kind)
	}

	// The block's switch (on by default once declared, filled by the platform
	// before this runs) decides whether the feature store renders.
	if spec.FeatureStore != nil && spec.FeatureStore.GetEnabled() {
		featureStoreArgs := &machinelearning.WorkspaceFeatureStoreArgs{}
		if spec.FeatureStore.ComputerSparkRuntimeVersion != "" {
			featureStoreArgs.ComputerSparkRuntimeVersion = pulumi.String(spec.FeatureStore.ComputerSparkRuntimeVersion)
		}
		if spec.FeatureStore.OfflineConnectionName != "" {
			featureStoreArgs.OfflineConnectionName = pulumi.String(spec.FeatureStore.OfflineConnectionName)
		}
		if spec.FeatureStore.OnlineConnectionName != "" {
			featureStoreArgs.OnlineConnectionName = pulumi.String(spec.FeatureStore.OnlineConnectionName)
		}
		workspaceArgs.FeatureStore = featureStoreArgs
	}

	if spec.PrimaryUserAssignedIdentity.GetValue() != "" {
		workspaceArgs.PrimaryUserAssignedIdentity = pulumi.String(spec.PrimaryUserAssignedIdentity.GetValue())
	}

	// ForceNew: attaching or re-pointing a registry replaces the
	// workspace.
	if spec.ContainerRegistryId.GetValue() != "" {
		workspaceArgs.ContainerRegistryId = pulumi.String(spec.ContainerRegistryId.GetValue())
	}

	// Optional-with-default-true on the provider: omit when the spec
	// leaves it unset so the provider default applies.
	if spec.PublicNetworkAccessEnabled != nil {
		workspaceArgs.PublicNetworkAccessEnabled = pulumi.Bool(*spec.PublicNetworkAccessEnabled)
	}

	if spec.ImageBuildComputeName != "" {
		workspaceArgs.ImageBuildComputeName = pulumi.String(spec.ImageBuildComputeName)
	}
	if spec.Description != "" {
		workspaceArgs.Description = pulumi.String(spec.Description)
	}
	if spec.FriendlyName != "" {
		workspaceArgs.FriendlyName = pulumi.String(spec.FriendlyName)
	}

	// Customer-managed-key encryption; the whole block is ForceNew.
	// The key id is a Key Vault key data-plane URL (versionless
	// follows rotation).
	if spec.Encryption != nil {
		encryptionArgs := &machinelearning.WorkspaceEncryptionArgs{
			KeyVaultId: pulumi.String(spec.Encryption.KeyVaultId.GetValue()),
			KeyId:      pulumi.String(spec.Encryption.KeyId.GetValue()),
		}
		if spec.Encryption.UserAssignedIdentityId.GetValue() != "" {
			encryptionArgs.UserAssignedIdentityId = pulumi.String(spec.Encryption.UserAssignedIdentityId.GetValue())
		}
		workspaceArgs.Encryption = encryptionArgs
	}

	// The managed virtual network. isolation_mode is Optional+Computed
	// on the provider -- unspecified omits it and the value is read
	// back.
	if spec.ManagedNetwork != nil {
		managedNetworkArgs := &machinelearning.WorkspaceManagedNetworkArgs{
			ProvisionOnCreationEnabled: pulumi.Bool(spec.ManagedNetwork.ProvisionOnCreationEnabled),
		}
		if isolationMode, ok := isolationModeWire[spec.ManagedNetwork.IsolationMode]; ok {
			managedNetworkArgs.IsolationMode = pulumi.String(isolationMode)
		}
		workspaceArgs.ManagedNetwork = managedNetworkArgs
	}

	// "Basic" is the only value the provider accepts at v5; unset
	// applies the provider default (also "Basic").
	if spec.SkuName != "" {
		workspaceArgs.SkuName = pulumi.String(spec.SkuName)
	}

	// Serverless compute. NOTE (update behavior, provider-enforced):
	// public_ip_enabled cannot flip true -> false while subnet_id is
	// unset; the static create-time rule is already spec CEL.
	if spec.ServerlessCompute != nil {
		serverlessComputeArgs := &machinelearning.WorkspaceServerlessComputeArgs{
			PublicIpEnabled: pulumi.Bool(spec.ServerlessCompute.PublicIpEnabled),
		}
		if spec.ServerlessCompute.SubnetId.GetValue() != "" {
			serverlessComputeArgs.SubnetId = pulumi.String(spec.ServerlessCompute.SubnetId.GetValue())
		}
		workspaceArgs.ServerlessCompute = serverlessComputeArgs
	}

	createdWorkspace, err := machinelearning.NewWorkspace(ctx,
		spec.Name,
		workspaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create machine learning workspace %s", spec.Name)
	}

	// The composed outbound rules: standalone ARM children of the
	// workspace's managed network, one per spec entry, keyed by name.
	// The three rule types share ONE ARM collection -- cross-type name
	// uniqueness is spec CEL. Only effective under
	// AllowOnlyApprovedOutbound isolation.
	fqdnOutboundRuleIds := pulumi.Map{}
	for _, rule := range spec.FqdnOutboundRules {
		createdRule, err := machinelearning.NewWorkspaceNetworkOutboundRuleFqdn(ctx,
			spec.Name+"-"+rule.Name,
			&machinelearning.WorkspaceNetworkOutboundRuleFqdnArgs{
				Name:            pulumi.String(rule.Name),
				WorkspaceId:     createdWorkspace.ID(),
				DestinationFqdn: pulumi.String(rule.DestinationFqdn),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdWorkspace))
		if err != nil {
			return errors.Wrapf(err, "failed to create fqdn outbound rule %s", rule.Name)
		}
		fqdnOutboundRuleIds[rule.Name] = createdRule.ID()
	}

	// Every field is ForceNew (the provider ships no update for this
	// rule type); the target/sub-resource pairing is spec CEL for
	// literal ids and provider-checked for references.
	privateEndpointOutboundRuleIds := pulumi.Map{}
	for _, rule := range spec.PrivateEndpointOutboundRules {
		createdRule, err := machinelearning.NewWorkspaceNetworkOutboundRulePrivateEndpoint(ctx,
			spec.Name+"-"+rule.Name,
			&machinelearning.WorkspaceNetworkOutboundRulePrivateEndpointArgs{
				Name:              pulumi.String(rule.Name),
				WorkspaceId:       createdWorkspace.ID(),
				ServiceResourceId: pulumi.String(rule.ServiceResourceId.GetValue()),
				SubResourceTarget: pulumi.String(rule.SubResourceTarget),
				SparkEnabled:      pulumi.Bool(rule.SparkEnabled),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdWorkspace))
		if err != nil {
			return errors.Wrapf(err, "failed to create private endpoint outbound rule %s", rule.Name)
		}
		privateEndpointOutboundRuleIds[rule.Name] = createdRule.ID()
	}

	serviceTagOutboundRuleIds := pulumi.Map{}
	for _, rule := range spec.ServiceTagOutboundRules {
		createdRule, err := machinelearning.NewWorkspaceNetworkOutboundRuleServiceTag(ctx,
			spec.Name+"-"+rule.Name,
			&machinelearning.WorkspaceNetworkOutboundRuleServiceTagArgs{
				Name:        pulumi.String(rule.Name),
				WorkspaceId: createdWorkspace.ID(),
				ServiceTag:  pulumi.String(rule.ServiceTag),
				Protocol:    pulumi.String(rule.Protocol),
				PortRanges:  pulumi.String(rule.PortRanges),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdWorkspace))
		if err != nil {
			return errors.Wrapf(err, "failed to create service tag outbound rule %s", rule.Name)
		}
		serviceTagOutboundRuleIds[rule.Name] = createdRule.ID()
	}

	ctx.Export(OpMachineLearningWorkspaceId, createdWorkspace.ID())
	ctx.Export(OpMachineLearningWorkspaceName, createdWorkspace.Name)
	ctx.Export(OpWorkspaceGuid, createdWorkspace.WorkspaceId)
	ctx.Export(OpDiscoveryUrl, createdWorkspace.DiscoveryUrl)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdWorkspace.Identity.PrincipalId())
	ctx.Export(OpFqdnOutboundRuleIds, fqdnOutboundRuleIds)
	ctx.Export(OpPrivateEndpointOutboundRuleIds, privateEndpointOutboundRuleIds)
	ctx.Export(OpServiceTagOutboundRuleIds, serviceTagOutboundRuleIds)

	return nil
}
