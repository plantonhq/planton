package module

import (
	"github.com/pkg/errors"
	azuremonitoractiongroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitoractiongroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitoractiongroupv1alpha1.AzureMonitorActionGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorActionGroup.Spec

	// The action group is the notification hub alert rules fire into. It is
	// a GLOBAL resource (the provider defaults location to "global"):
	// notifications keep flowing during regional outages, which is when
	// they matter most.
	actionGroupArgs := &monitoring.ActionGroupArgs{
		Name:              pulumi.String(locals.AzureMonitorActionGroup.Metadata.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ShortName:         pulumi.String(spec.ShortName),
		// Presence-guarded to the proto default: a disabled group silently
		// swallows every alert that fires into it.
		Enabled: pulumi.Bool(presenceGuardedBool(spec.Enabled, true)),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	if len(spec.EmailReceivers) > 0 {
		receivers := monitoring.ActionGroupEmailReceiverArray{}
		for _, receiver := range spec.EmailReceivers {
			receivers = append(receivers, monitoring.ActionGroupEmailReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				EmailAddress:         pulumi.String(receiver.EmailAddress),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			})
		}
		actionGroupArgs.EmailReceivers = receivers
	}

	// SMS and voice carry no schema toggle -- they are not payload-aware.
	if len(spec.SmsReceivers) > 0 {
		receivers := monitoring.ActionGroupSmsReceiverArray{}
		for _, receiver := range spec.SmsReceivers {
			receivers = append(receivers, monitoring.ActionGroupSmsReceiverArgs{
				Name:        pulumi.String(receiver.Name),
				CountryCode: pulumi.String(receiver.CountryCode),
				PhoneNumber: pulumi.String(receiver.PhoneNumber),
			})
		}
		actionGroupArgs.SmsReceivers = receivers
	}

	if len(spec.VoiceReceivers) > 0 {
		receivers := monitoring.ActionGroupVoiceReceiverArray{}
		for _, receiver := range spec.VoiceReceivers {
			receivers = append(receivers, monitoring.ActionGroupVoiceReceiverArgs{
				Name:        pulumi.String(receiver.Name),
				CountryCode: pulumi.String(receiver.CountryCode),
				PhoneNumber: pulumi.String(receiver.PhoneNumber),
			})
		}
		actionGroupArgs.VoiceReceivers = receivers
	}

	if len(spec.WebhookReceivers) > 0 {
		receivers := monitoring.ActionGroupWebhookReceiverArray{}
		for _, receiver := range spec.WebhookReceivers {
			receiverArgs := monitoring.ActionGroupWebhookReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				ServiceUri:           pulumi.String(receiver.ServiceUri),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			}
			// Entra-authenticated webhooks: the keyless posture -- no secret
			// baked into the URL.
			if receiver.AadAuth != nil {
				aadAuthArgs := &monitoring.ActionGroupWebhookReceiverAadAuthArgs{
					ObjectId: pulumi.String(receiver.AadAuth.ObjectId),
				}
				if receiver.AadAuth.IdentifierUri != "" {
					aadAuthArgs.IdentifierUri = pulumi.String(receiver.AadAuth.IdentifierUri)
				}
				if receiver.AadAuth.TenantId != "" {
					aadAuthArgs.TenantId = pulumi.String(receiver.AadAuth.TenantId)
				}
				receiverArgs.AadAuth = aadAuthArgs
			}
			receivers = append(receivers, receiverArgs)
		}
		actionGroupArgs.WebhookReceivers = receivers
	}

	if len(spec.AzureAppPushReceivers) > 0 {
		receivers := monitoring.ActionGroupAzureAppPushReceiverArray{}
		for _, receiver := range spec.AzureAppPushReceivers {
			receivers = append(receivers, monitoring.ActionGroupAzureAppPushReceiverArgs{
				Name:         pulumi.String(receiver.Name),
				EmailAddress: pulumi.String(receiver.EmailAddress),
			})
		}
		actionGroupArgs.AzureAppPushReceivers = receivers
	}

	if len(spec.AutomationRunbookReceivers) > 0 {
		receivers := monitoring.ActionGroupAutomationRunbookReceiverArray{}
		for _, receiver := range spec.AutomationRunbookReceivers {
			receivers = append(receivers, monitoring.ActionGroupAutomationRunbookReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				AutomationAccountId:  pulumi.String(receiver.AutomationAccountId),
				RunbookName:          pulumi.String(receiver.RunbookName),
				WebhookResourceId:    pulumi.String(receiver.WebhookResourceId),
				IsGlobalRunbook:      pulumi.Bool(receiver.IsGlobalRunbook),
				ServiceUri:           pulumi.String(receiver.ServiceUri),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			})
		}
		actionGroupArgs.AutomationRunbookReceivers = receivers
	}

	if len(spec.LogicAppReceivers) > 0 {
		receivers := monitoring.ActionGroupLogicAppReceiverArray{}
		for _, receiver := range spec.LogicAppReceivers {
			receivers = append(receivers, monitoring.ActionGroupLogicAppReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				ResourceId:           pulumi.String(receiver.ResourceId),
				CallbackUrl:          pulumi.String(receiver.CallbackUrl),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			})
		}
		actionGroupArgs.LogicAppReceivers = receivers
	}

	if len(spec.AzureFunctionReceivers) > 0 {
		receivers := monitoring.ActionGroupAzureFunctionReceiverArray{}
		for _, receiver := range spec.AzureFunctionReceivers {
			receivers = append(receivers, monitoring.ActionGroupAzureFunctionReceiverArgs{
				Name:                  pulumi.String(receiver.Name),
				FunctionAppResourceId: pulumi.String(receiver.FunctionAppResourceId.GetValue()),
				FunctionName:          pulumi.String(receiver.FunctionName),
				HttpTriggerUrl:        pulumi.String(receiver.HttpTriggerUrl),
				UseCommonAlertSchema:  pulumi.Bool(receiver.UseCommonAlertSchema),
			})
		}
		actionGroupArgs.AzureFunctionReceivers = receivers
	}

	// Role-based fan-out: every user holding the role on the subscription
	// is notified -- no address list to maintain.
	if len(spec.ArmRoleReceivers) > 0 {
		receivers := monitoring.ActionGroupArmRoleReceiverArray{}
		for _, receiver := range spec.ArmRoleReceivers {
			receivers = append(receivers, monitoring.ActionGroupArmRoleReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				RoleId:               pulumi.String(receiver.RoleId.GetValue()),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			})
		}
		actionGroupArgs.ArmRoleReceivers = receivers
	}

	if len(spec.EventHubReceivers) > 0 {
		receivers := monitoring.ActionGroupEventHubReceiverArray{}
		for _, receiver := range spec.EventHubReceivers {
			receiverArgs := monitoring.ActionGroupEventHubReceiverArgs{
				Name:                 pulumi.String(receiver.Name),
				EventHubName:         pulumi.String(receiver.EventHubName.GetValue()),
				EventHubNamespace:    pulumi.String(receiver.EventHubNamespace.GetValue()),
				UseCommonAlertSchema: pulumi.Bool(receiver.UseCommonAlertSchema),
			}
			if receiver.TenantId != "" {
				receiverArgs.TenantId = pulumi.String(receiver.TenantId)
			}
			if receiver.SubscriptionId != "" {
				receiverArgs.SubscriptionId = pulumi.String(receiver.SubscriptionId)
			}
			receivers = append(receivers, receiverArgs)
		}
		actionGroupArgs.EventHubReceivers = receivers
	}

	// ITSM work-item creation; the ticket_configuration JSON must carry
	// PayloadRevision and WorkItemType (spec-enforced).
	if len(spec.ItsmReceivers) > 0 {
		receivers := monitoring.ActionGroupItsmReceiverArray{}
		for _, receiver := range spec.ItsmReceivers {
			receivers = append(receivers, monitoring.ActionGroupItsmReceiverArgs{
				Name:                pulumi.String(receiver.Name),
				WorkspaceId:         pulumi.String(receiver.WorkspaceId),
				ConnectionId:        pulumi.String(receiver.ConnectionId),
				TicketConfiguration: pulumi.String(receiver.TicketConfiguration),
				Region:              pulumi.String(receiver.Region),
			})
		}
		actionGroupArgs.ItsmReceivers = receivers
	}

	createdActionGroup, err := monitoring.NewActionGroup(ctx,
		locals.AzureMonitorActionGroup.Metadata.Name,
		actionGroupArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create action group %s", locals.AzureMonitorActionGroup.Metadata.Name)
	}

	// Export outputs. action_group_id is the composition seam alert
	// rules reference.
	ctx.Export(OpActionGroupId, createdActionGroup.ID())
	ctx.Export(OpActionGroupName, createdActionGroup.Name)

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
