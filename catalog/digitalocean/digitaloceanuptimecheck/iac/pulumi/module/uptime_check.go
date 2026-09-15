package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// uptimeCheck provisions the uptime check, one alert resource per spec
// alert row, and exports the check's outputs.
func uptimeCheck(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.UptimeCheck, error) {
	spec := locals.DigitalOceanUptimeCheck.Spec

	// Always declared (spec-required): the provider never reconciles a
	// DigitalOcean-defaulted region set, so an omitted value would leave
	// every subsequent plan trying to remove what the API chose.
	var regions pulumi.StringArray
	for _, region := range spec.Regions {
		regions = append(regions, pulumi.String(region))
	}

	checkArgs := &digitalocean.UptimeCheckArgs{
		Name:    pulumi.StringPtr(spec.CheckName),
		Target:  pulumi.String(spec.Target),
		Regions: regions,
	}

	// Unset defers to the provider's default, https.
	if spec.Type != "" {
		checkArgs.Type = pulumi.StringPtr(spec.Type)
	}

	// Unset defers to the provider's default, enabled.
	if spec.Enabled != nil {
		checkArgs.Enabled = pulumi.BoolPtr(spec.GetEnabled())
	}

	createdCheck, err := digitalocean.NewUptimeCheck(
		ctx,
		"check",
		checkArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean uptime check")
	}

	// One alert resource per spec row, composed on the check (the
	// standalone resource's mutable parent id is a corruption class this
	// composition makes unrepresentable). Each row's resource name is
	// "<row index>-<alert name>" -- the same key the Terraform module uses
	// as its for_each key and both engines use as the alert_ids output key,
	// so a blind import can find each row's id from state alone. The index
	// lets two rows share a display name without colliding.
	alertIds := pulumi.StringMap{}
	for idx, alert := range spec.Alerts {
		alertKey := fmt.Sprintf("%d-%s", idx, alert.AlertName)

		// The SDK keeps the provider's unbounded notifications list; the
		// provider reads only the first element, so exactly one is sent.
		notificationArgs := digitalocean.UptimeAlertNotificationArgs{}
		if len(alert.Notifications.Emails) > 0 {
			var emails pulumi.StringArray
			for _, email := range alert.Notifications.Emails {
				emails = append(emails, pulumi.String(email))
			}
			notificationArgs.Emails = emails
		}
		if len(alert.Notifications.Slack) > 0 {
			var slacks digitalocean.UptimeAlertNotificationSlackArray
			for _, slack := range alert.Notifications.Slack {
				slacks = append(slacks, digitalocean.UptimeAlertNotificationSlackArgs{
					Channel: pulumi.String(slack.Channel),
					// The SDK does not flag the webhook URL as secret, so
					// wrap it explicitly -- a credential must never ship
					// as plaintext state.
					Url: pulumi.ToSecret(pulumi.String(slack.Url)).(pulumi.StringOutput),
				})
			}
			notificationArgs.Slacks = slacks
		}

		alertArgs := &digitalocean.UptimeAlertArgs{
			CheckId:       createdCheck.ID(),
			Name:          pulumi.StringPtr(alert.AlertName),
			Type:          pulumi.String(alert.Type),
			Notifications: digitalocean.UptimeAlertNotificationArray{notificationArgs},
			// Required by the API for every alert type (spec-required).
			Period: pulumi.StringPtr(alert.Period),
		}

		// DigitalOcean fixes threshold and comparison for some alert types
		// no matter what is sent (measured live: a down alert created with
		// 3 / greater_than reads back as 1 / less_than; an ssl_expiry alert
		// always reads back less_than). The spec rejects those fields on
		// those types, and the module sends the API's own values so state
		// and read-back agree on every apply -- sending nothing reads back
		// as a perpetual diff.
		switch alert.Type {
		case "down", "down_global":
			alertArgs.Threshold = pulumi.IntPtr(1)
			alertArgs.Comparison = pulumi.StringPtr("less_than")
		case "ssl_expiry":
			// Days before expiry, authored (spec-required for this type).
			alertArgs.Threshold = pulumi.IntPtr(int(alert.GetThreshold()))
			alertArgs.Comparison = pulumi.StringPtr("less_than")
		default:
			// latency: milliseconds and direction, both authored
			// (spec-required for this type).
			alertArgs.Threshold = pulumi.IntPtr(int(alert.GetThreshold()))
			alertArgs.Comparison = pulumi.StringPtr(alert.Comparison)
		}

		createdAlert, err := digitalocean.NewUptimeAlert(
			ctx,
			alertKey,
			alertArgs,
			pulumi.Provider(digitalOceanProvider),
			pulumi.Parent(createdCheck),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create uptime alert %q", alert.AlertName)
		}
		alertIds[alertKey] = createdAlert.ID().ToStringOutput()
	}

	ctx.Export(OpCheckId, createdCheck.ID())
	ctx.Export(OpAlertIds, alertIds)

	return createdCheck, nil
}
