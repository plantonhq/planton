package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// turnstileWidget provisions a Turnstile widget and exports its site key plus the
// (sensitive) secret key. Optional flags are sent only when set, matching the
// Terraform module so both engines rely on the provider's defaults identically.
func turnstileWidget(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.TurnstileWidget, error) {
	spec := locals.CloudflareTurnstileWidget.Spec

	domains := make(pulumi.StringArray, 0, len(spec.Domains))
	for _, d := range spec.Domains {
		domains = append(domains, pulumi.String(d))
	}

	args := &cloudflare.TurnstileWidgetArgs{
		AccountId: pulumi.String(spec.AccountId),
		Name:      pulumi.String(spec.Name),
		Domains:   domains,
		Mode:      pulumi.String(spec.Mode),
	}
	if spec.ClearanceLevel != "" {
		args.ClearanceLevel = pulumi.StringPtr(spec.ClearanceLevel)
	}
	if spec.BotFightMode {
		args.BotFightMode = pulumi.BoolPtr(true)
	}
	if spec.EphemeralId {
		args.EphemeralId = pulumi.BoolPtr(true)
	}
	if spec.Offlabel {
		args.Offlabel = pulumi.BoolPtr(true)
	}
	if spec.Region != "" {
		args.Region = pulumi.StringPtr(spec.Region)
	}

	created, err := cloudflare.NewTurnstileWidget(
		ctx,
		"turnstile-widget",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare turnstile widget")
	}

	ctx.Export(OpSitekey, created.Sitekey)
	ctx.Export(OpSecret, pulumi.ToSecret(created.Secret))
	ctx.Export(OpCreatedOn, created.CreatedOn)
	ctx.Export(OpModifiedOn, created.ModifiedOn)

	return created, nil
}
