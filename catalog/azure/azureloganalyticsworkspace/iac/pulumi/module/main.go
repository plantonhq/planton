package module

import (
	"github.com/pkg/errors"
	azureloganalyticsworkspacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureloganalyticsworkspace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/operationalinsights"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureloganalyticsworkspacev1alpha1.AzureLogAnalyticsWorkspaceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureLogAnalyticsWorkspace.Spec

	// The workspace is the central Azure Monitor data platform: diagnostic
	// settings, Container Insights, Application Insights, and Sentinel all
	// store into it. Switching between PerGB2018 and CapacityReservation
	// updates in place; any other SKU change is ForceNew (the provider's
	// transition rule).
	workspaceArgs := &operationalinsights.AnalyticsWorkspaceArgs{
		Name:              pulumi.String(spec.WorkspaceName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(skuStrings[spec.Sku]),
		// Presence-guarded to the proto defaults: IaC inputs built from a
		// manifest materialize defaults, but direct iac-input paths do not.
		RetentionInDays: pulumi.Int(int(spec.GetRetentionInDays())),
		// -1 means unlimited -- the provider's own default, sent explicitly so
		// both engines carry the same value.
		DailyQuotaGb: pulumi.Float64(spec.GetDailyQuotaGb()),
		Tags:         pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.RetentionInDays == nil {
		workspaceArgs.RetentionInDays = pulumi.Int(30)
	}
	if spec.DailyQuotaGb == nil {
		workspaceArgs.DailyQuotaGb = pulumi.Float64(-1)
	}

	// The commitment tier is only legal with the CapacityReservation SKU
	// (spec-enforced pairing); sent only when set so pay-as-you-go
	// workspaces never carry a capacity level.
	if spec.ReservationCapacityInGbPerDay != nil {
		workspaceArgs.ReservationCapacityInGbPerDay = pulumi.Int(int(spec.GetReservationCapacityInGbPerDay()))
	}

	// Security and network posture. All four default to Azure's own defaults
	// (true); explicit false is preserved because the spec models presence.
	workspaceArgs.LocalAuthenticationEnabled = pulumi.Bool(presenceGuardedBool(spec.LocalAuthenticationEnabled, true))
	workspaceArgs.InternetIngestionEnabled = pulumi.Bool(presenceGuardedBool(spec.InternetIngestionEnabled, true))
	workspaceArgs.InternetQueryEnabled = pulumi.Bool(presenceGuardedBool(spec.InternetQueryEnabled, true))
	workspaceArgs.AllowResourceOnlyPermissions = pulumi.Bool(presenceGuardedBool(spec.AllowResourceOnlyPermissions, true))

	workspaceArgs.CmkForQueryForced = pulumi.Bool(spec.CmkForQueryForced)
	workspaceArgs.ImmediateDataPurgeOn30DaysEnabled = pulumi.Bool(spec.ImmediateDataPurgeOn_30DaysEnabled)

	// The default DCR arrives as a resolved ARM id (references resolve
	// before the module runs); the provider applies it via a follow-up
	// update call because ARM rejects a default DCR at workspace creation.
	if spec.DataCollectionRuleId.GetValue() != "" {
		workspaceArgs.DataCollectionRuleId = pulumi.String(spec.DataCollectionRuleId.GetValue())
	}

	// Managed identity -- used when the workspace itself reads other
	// resources (dedicated-cluster CMK, linked storage).
	if spec.Identity != nil {
		identityArgs := &operationalinsights.AnalyticsWorkspaceIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		workspaceArgs.Identity = identityArgs
	}

	createdWorkspace, err := operationalinsights.NewAnalyticsWorkspace(ctx,
		spec.WorkspaceName,
		workspaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Log Analytics Workspace %s", spec.WorkspaceName)
	}

	// Export outputs. workspace_id (the ARM resource ID) is the FK
	// seam downstream kinds reference; the provider's WorkspaceId attribute
	// is the CUSTOMER GUID, exported under the unambiguous name.
	ctx.Export(OpWorkspaceId, createdWorkspace.ID())
	ctx.Export(OpWorkspaceName, createdWorkspace.Name)
	ctx.Export(OpWorkspaceCustomerId, createdWorkspace.WorkspaceId)
	ctx.Export(OpResourceGroupName, createdWorkspace.ResourceGroupName)
	ctx.Export(OpPrimarySharedKey, pulumi.ToSecret(createdWorkspace.PrimarySharedKey))
	ctx.Export(OpSecondarySharedKey, pulumi.ToSecret(createdWorkspace.SecondarySharedKey))
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdWorkspace.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}

// presenceGuardedBool returns the field's value when set and the proto
// default otherwise -- default materialization is middleware behavior, not a
// wire guarantee.
func presenceGuardedBool(field *bool, defaultValue bool) bool {
	if field == nil {
		return defaultValue
	}
	return *field
}
