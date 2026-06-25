package module

import (
	"github.com/pkg/errors"
	cloudflareemailroutingzonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareemailroutingzone/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// emailRoutingZone enables Email Routing on the zone (which provisions the
// required DNS records), optionally folds in the single catch-all rule, and
// optionally locks the DNS records.
func emailRoutingZone(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareEmailRoutingZone.Spec

	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}

	settings, err := cloudflare.NewEmailRoutingSettings(
		ctx,
		"email-routing-settings",
		&cloudflare.EmailRoutingSettingsArgs{
			ZoneId: pulumi.String(zoneId),
		},
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to enable email routing settings")
	}

	if ca := spec.CatchAll; ca != nil {
		// Map the typed catch-all action onto the provider's generic {type, value[]}.
		values := pulumi.StringArray{}
		switch ca.Type {
		case cloudflareemailroutingzonev1.CloudflareEmailRoutingCatchAllActionType_forward:
			for _, f := range ca.ForwardTo {
				values = append(values, pulumi.String(f.GetValue()))
			}
		case cloudflareemailroutingzonev1.CloudflareEmailRoutingCatchAllActionType_worker:
			if ca.Worker != nil {
				values = append(values, pulumi.String(ca.Worker.GetValue()))
			}
		}

		_, err := cloudflare.NewEmailRoutingCatchAll(
			ctx,
			"email-routing-catch-all",
			&cloudflare.EmailRoutingCatchAllArgs{
				ZoneId:  pulumi.String(zoneId),
				Enabled: pulumi.Bool(ca.Enabled),
				Matchers: cloudflare.EmailRoutingCatchAllMatcherArray{
					cloudflare.EmailRoutingCatchAllMatcherArgs{Type: pulumi.String("all")},
				},
				Actions: cloudflare.EmailRoutingCatchAllActionArray{
					cloudflare.EmailRoutingCatchAllActionArgs{
						Type:   pulumi.String(ca.Type.String()),
						Values: values,
					},
				},
			},
			pulumi.Provider(cloudflareProvider),
			pulumi.DependsOn([]pulumi.Resource{settings}),
		)
		if err != nil {
			return errors.Wrap(err, "failed to create email routing catch-all")
		}
	}

	if spec.LockDnsRecords {
		_, err := cloudflare.NewEmailRoutingDns(
			ctx,
			"email-routing-dns",
			&cloudflare.EmailRoutingDnsArgs{
				ZoneId: pulumi.String(zoneId),
			},
			pulumi.Provider(cloudflareProvider),
			pulumi.DependsOn([]pulumi.Resource{settings}),
		)
		if err != nil {
			return errors.Wrap(err, "failed to lock email routing dns records")
		}
	}

	ctx.Export(OpZoneId, pulumi.String(zoneId))
	ctx.Export(OpEnabled, settings.Enabled)
	ctx.Export(OpStatus, settings.Status)
	ctx.Export(OpName, settings.Name)

	return nil
}
