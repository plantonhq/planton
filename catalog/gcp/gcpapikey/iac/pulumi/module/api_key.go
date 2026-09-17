package module

import (
	"github.com/pkg/errors"
	gcpapikeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpapikey/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// apiKey provisions the API key with its restrictions.
//
// The key's identity is the spec's key_id (the provider's `name`); the
// project and the optional service-account binding are immutable with it,
// so a change to any of the three is a destroy-and-recreate that rotates
// the key string every client holds. Restrictions are updated in place.
//
// Restriction arms are sent only when the spec declares them: an omitted
// arm means "no restriction of that class" to the API, and sending an
// empty block would be rejected (each arm's list is required by the API).
// The spec's CEL already guarantees at most one client arm.
func apiKey(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpApiKey.Spec

	args := &projects.ApiKeyArgs{
		Name: pulumi.String(spec.KeyId),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if spec.ServiceAccountEmail.GetValue() != "" {
		args.ServiceAccountEmail = pulumi.StringPtr(spec.ServiceAccountEmail.GetValue())
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	if spec.Restrictions != nil {
		args.Restrictions = restrictionsArgs(spec.Restrictions)
	}

	createdKey, err := projects.NewApiKey(ctx, spec.KeyId, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create api key")
	}

	// The resource ID is the full name the API addresses the key by; the
	// provider's `name` attribute is the short key_id.
	ctx.Export(OpName, createdKey.ID().ToStringOutput())
	ctx.Export(OpUid, createdKey.Uid)
	// The key string is a credential Google bills against the project:
	// marked secret so it never prints, matching the Terraform module's
	// `sensitive = true` on the same output.
	ctx.Export(OpKeyString, pulumi.ToSecret(createdKey.KeyString))

	return nil
}

// restrictionsArgs maps the spec's restrictions onto the provider's block,
// arm by arm. Every list the API requires non-empty is guaranteed non-empty
// by the spec's validation, so no arm is ever sent hollow.
func restrictionsArgs(r *gcpapikeyv1alpha1.GcpApiKeyRestrictions) *projects.ApiKeyRestrictionsArgs {
	out := &projects.ApiKeyRestrictionsArgs{}

	if r.AndroidKeyRestrictions != nil {
		apps := projects.ApiKeyRestrictionsAndroidKeyRestrictionsAllowedApplicationArray{}
		for _, app := range r.AndroidKeyRestrictions.AllowedApplications {
			apps = append(apps, &projects.ApiKeyRestrictionsAndroidKeyRestrictionsAllowedApplicationArgs{
				PackageName:     pulumi.String(app.PackageName),
				Sha1Fingerprint: pulumi.String(app.Sha1Fingerprint),
			})
		}
		out.AndroidKeyRestrictions = &projects.ApiKeyRestrictionsAndroidKeyRestrictionsArgs{
			AllowedApplications: apps,
		}
	}
	if r.IosKeyRestrictions != nil {
		out.IosKeyRestrictions = &projects.ApiKeyRestrictionsIosKeyRestrictionsArgs{
			AllowedBundleIds: pulumi.ToStringArray(r.IosKeyRestrictions.AllowedBundleIds),
		}
	}
	if r.BrowserKeyRestrictions != nil {
		out.BrowserKeyRestrictions = &projects.ApiKeyRestrictionsBrowserKeyRestrictionsArgs{
			AllowedReferrers: pulumi.ToStringArray(r.BrowserKeyRestrictions.AllowedReferrers),
		}
	}
	if r.ServerKeyRestrictions != nil {
		out.ServerKeyRestrictions = &projects.ApiKeyRestrictionsServerKeyRestrictionsArgs{
			AllowedIps: pulumi.ToStringArray(r.ServerKeyRestrictions.AllowedIps),
		}
	}
	if len(r.ApiTargets) > 0 {
		targets := projects.ApiKeyRestrictionsApiTargetArray{}
		for _, t := range r.ApiTargets {
			targetArgs := &projects.ApiKeyRestrictionsApiTargetArgs{
				Service: pulumi.String(t.Service),
			}
			if len(t.Methods) > 0 {
				targetArgs.Methods = pulumi.ToStringArray(t.Methods)
			}
			targets = append(targets, targetArgs)
		}
		out.ApiTargets = targets
	}
	return out
}
