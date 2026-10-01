package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// webAnalyticsSite creates the Web Analytics (RUM) site plus its folded
// measurement rules. The site is identified by host OR zone (spec
// validation enforces exactly one).
//
// Cloudflare stores include/exclude rules as separate objects under the
// site's ruleset; this module manages one rule object per declared row,
// keyed by position. The provider never reads rules back after writing them
// (its refresh is deliberately blind), so each apply re-asserts exactly the
// declared rows.
func webAnalyticsSite(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareWebAnalyticsSite.Spec

	args := &cloudflare.WebAnalyticsSiteArgs{
		AccountId: pulumi.String(spec.AccountId),
	}

	if spec.Host != "" {
		args.Host = pulumi.StringPtr(spec.Host)
	}
	if spec.ZoneTag.GetValue() != "" {
		args.ZoneTag = pulumi.StringPtr(spec.ZoneTag.GetValue())
	}
	// auto_install is ALWAYS sent: Cloudflare echoes a server-side false
	// for it even when never sent (measured on host AND zone sites), so an
	// omitted send drifts forever on Terraform and unset-means-false is
	// exactly Cloudflare's default. enabled/lite stay conditional: they are
	// no_refresh (never read back), so no echo can drift them.
	args.AutoInstall = pulumi.BoolPtr(spec.GetAutoInstall())
	if spec.Enabled != nil {
		args.Enabled = pulumi.BoolPtr(spec.GetEnabled())
	}
	if spec.Lite != nil {
		args.Lite = pulumi.BoolPtr(spec.GetLite())
	}

	createdSite, err := cloudflare.NewWebAnalyticsSite(
		ctx,
		"web_analytics_site",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create web analytics site")
	}

	// Read-after-create (measured 2026-08-27): the create response omits
	// `snippet` (GET-only), so the snippet/ruleset_id outputs and the
	// folded rules ride this lookup, not the resource. The `ruleset`
	// object itself is IDENTITY-DEPENDENT: zone-linked sites carry it in
	// every response; host-identified sites have NO ruleset, ever (the
	// spec walls rules to zone_tag sites for exactly that reason -- on a
	// host site the lookup's Ruleset() is a zero struct and the
	// ruleset_id export is ""). NOTE: reading site_info requires the
	// Account Settings READ permission -- Account Settings Write alone
	// creates sites but cannot read them back (measured 403/10000).
	lookedUpSite := cloudflare.LookupWebAnalyticsSiteOutput(
		ctx,
		cloudflare.LookupWebAnalyticsSiteOutputArgs{
			AccountId: pulumi.StringPtr(spec.AccountId),
			SiteId:    createdSite.ID().ToStringPtrOutput(),
		},
		pulumi.Provider(cloudflareProvider),
	)

	rulesetId := lookedUpSite.Ruleset().Id()

	// Every rule field is ALWAYS sent (measured 2026-08-27): Cloudflare's
	// rule form validates each field's presence and rejects omissions one
	// by one (400 code 10001 "form.host.invalid" / "form.is_paused.invalid"
	// -- the provider passes nulls straight through because upstream never
	// exercises rules). An empty spec host means "every host", which the
	// API spells "*".
	for index, row := range spec.Rules {
		ruleHost := row.Host
		if ruleHost == "" {
			ruleHost = "*"
		}
		ruleArgs := &cloudflare.WebAnalyticsRuleArgs{
			AccountId: pulumi.String(spec.AccountId),
			RulesetId: rulesetId,
			Host:      pulumi.StringPtr(ruleHost),
			Inclusive: pulumi.BoolPtr(row.GetInclusive()),
			IsPaused:  pulumi.BoolPtr(row.GetIsPaused()),
		}
		if len(row.Paths) > 0 {
			ruleArgs.Paths = pulumi.ToStringArray(row.Paths)
		}

		if _, err := cloudflare.NewWebAnalyticsRule(
			ctx,
			fmt.Sprintf("web_analytics_rule_%d", index),
			ruleArgs,
			pulumi.Provider(cloudflareProvider),
			pulumi.Parent(createdSite),
		); err != nil {
			return errors.Wrapf(err, "failed to create web analytics rule %d", index)
		}
	}

	// The site tag is the one value the create response returns (the
	// resource id IS the tag); token and snippet come from the
	// read-after-create lookup. Both ship inside public pages, so neither
	// is exported as a secret.
	ctx.Export(OpSiteTag, createdSite.ID())
	ctx.Export(OpSiteToken, lookedUpSite.SiteToken())
	ctx.Export(OpSnippet, lookedUpSite.Snippet())
	ctx.Export(OpRulesetId, rulesetId)

	return nil
}
