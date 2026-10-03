package module

import (
	"github.com/pkg/errors"
	azureappinsightswebtestv1 "github.com/plantonhq/planton/catalog/azure/azureapplicationinsightsstandardwebtest/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appinsights"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureappinsightswebtestv1.AzureApplicationInsightsStandardWebTestIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureApplicationInsightsStandardWebTest.Spec

	// The request block is required. Optional fields are sent only when set
	// so an unspecified spec and Azure's defaults (GET, follow-redirects,
	// parse-dependent-requests) deploy identically on both engines.
	req := spec.Request
	requestArgs := appinsights.StandardWebTestRequestArgs{
		Url: pulumi.String(req.Url),
	}
	if req.HttpVerb != "" {
		requestArgs.HttpVerb = pulumi.String(req.HttpVerb)
	}
	if req.Body != "" {
		requestArgs.Body = pulumi.String(req.Body)
	}
	if req.FollowRedirectsEnabled != nil {
		requestArgs.FollowRedirectsEnabled = pulumi.Bool(req.GetFollowRedirectsEnabled())
	}
	if req.ParseDependentRequestsEnabled != nil {
		requestArgs.ParseDependentRequestsEnabled = pulumi.Bool(req.GetParseDependentRequestsEnabled())
	}
	if len(req.Headers) > 0 {
		headers := appinsights.StandardWebTestRequestHeaderArray{}
		for _, h := range req.Headers {
			headers = append(headers, appinsights.StandardWebTestRequestHeaderArgs{
				Name:  pulumi.String(h.Name),
				Value: pulumi.String(h.Value),
			})
		}
		requestArgs.Headers = headers
	}

	webTestArgs := &appinsights.StandardWebTestArgs{
		Name:                  pulumi.String(spec.Name),
		ResourceGroupName:     pulumi.String(locals.ResourceGroupName),
		ApplicationInsightsId: pulumi.String(locals.ApplicationInsightsId),
		Location:              pulumi.String(spec.Region),
		GeoLocations:          pulumi.ToStringArray(spec.GeoLocations),
		Request:               requestArgs,
		Tags:                  pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Frequency != nil {
		webTestArgs.Frequency = pulumi.Int(int(spec.GetFrequency()))
	}
	if spec.Timeout != nil {
		webTestArgs.Timeout = pulumi.Int(int(spec.GetTimeout()))
	}
	if spec.Enabled != nil {
		webTestArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}
	if spec.RetryEnabled != nil {
		webTestArgs.RetryEnabled = pulumi.Bool(spec.GetRetryEnabled())
	}
	if spec.Description != "" {
		webTestArgs.Description = pulumi.String(spec.Description)
	}

	// Validation rules are optional. Each sub-field is sent only when set so
	// the module lets Azure apply its defaults (200 status, SSL check off).
	if vr := spec.ValidationRules; vr != nil {
		vrArgs := &appinsights.StandardWebTestValidationRulesArgs{}
		if vr.ExpectedStatusCode != nil {
			vrArgs.ExpectedStatusCode = pulumi.Int(int(vr.GetExpectedStatusCode()))
		}
		if vr.SslCertRemainingLifetime != nil {
			vrArgs.SslCertRemainingLifetime = pulumi.Int(int(vr.GetSslCertRemainingLifetime()))
		}
		if vr.SslCheckEnabled != nil {
			vrArgs.SslCheckEnabled = pulumi.Bool(vr.GetSslCheckEnabled())
		}
		if c := vr.Content; c != nil {
			contentArgs := &appinsights.StandardWebTestValidationRulesContentArgs{
				ContentMatch: pulumi.String(c.ContentMatch),
			}
			if c.IgnoreCase != nil {
				contentArgs.IgnoreCase = pulumi.Bool(c.GetIgnoreCase())
			}
			if c.PassIfTextFound != nil {
				contentArgs.PassIfTextFound = pulumi.Bool(c.GetPassIfTextFound())
			}
			vrArgs.Content = contentArgs
		}
		webTestArgs.ValidationRules = vrArgs
	}

	createdWebTest, err := appinsights.NewStandardWebTest(ctx,
		spec.Name,
		webTestArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create standard web test %s", spec.Name)
	}

	ctx.Export(OpWebTestId, createdWebTest.ID())
	ctx.Export(OpWebTestName, createdWebTest.Name)
	ctx.Export(OpSyntheticMonitorId, createdWebTest.SyntheticMonitorId)

	return nil
}
