package module

import (
	"github.com/pkg/errors"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// route creates DNS record and attaches the Worker script to a URL pattern (if DNS is configured).
func route(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudfl.Provider,
	createdWorkerScript *cloudfl.WorkersScript,
) ([]pulumi.StringOutput, error) {

	// Check if DNS configuration is provided and enabled
	if locals.CloudflareWorker.Spec.Dns == nil || !locals.CloudflareWorker.Spec.Dns.Enabled {
		// No DNS/route configuration requested or explicitly disabled.
		return nil, nil
	}

	dns := locals.CloudflareWorker.Spec.Dns

	// Resolve zone_id from literal value or reference.
	zoneId := ""
	if dns.ZoneId != nil {
		zoneId = dns.ZoneId.GetValue()
	}
	if zoneId == "" {
		return nil, errors.New("dns.zone_id is required when dns is enabled")
	}
	if dns.Hostname == "" {
		return nil, errors.New("dns.hostname is required when dns is enabled")
	}

	zoneIdOutput := pulumi.String(zoneId).ToStringOutput()

	// Create DNS record for the hostname
	createdDnsRecord, err := createDnsRecord(ctx, locals, cloudflareProvider, zoneIdOutput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dns record")
	}

	// Determine the route pattern: use specified pattern or default to "hostname/*"
	routePattern := dns.RoutePattern
	if routePattern == "" {
		routePattern = dns.Hostname + "/*"
	}

	routeArgs := &cloudfl.WorkersRouteArgs{
		ZoneId:  zoneIdOutput,
		Pattern: pulumi.String(routePattern),
		Script:  pulumi.String(locals.CloudflareWorker.Spec.WorkerName),
	}

	// Create the route, ensuring it depends on both DNS record and worker script
	var routeOptions []pulumi.ResourceOption
	routeOptions = append(routeOptions, pulumi.Provider(cloudflareProvider))

	// Build dependencies list
	var dependencies []pulumi.Resource
	if createdDnsRecord != nil {
		dependencies = append(dependencies, createdDnsRecord)
	}
	if createdWorkerScript != nil {
		dependencies = append(dependencies, createdWorkerScript)
	}
	if len(dependencies) > 0 {
		routeOptions = append(routeOptions, pulumi.DependsOn(dependencies))
	}

	createdWorkerRoute, err := cloudfl.NewWorkersRoute(
		ctx,
		"workers-route",
		routeArgs,
		routeOptions...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare workers route")
	}

	ctx.Export(OpRouteUrls, pulumi.StringArray{createdWorkerRoute.Pattern})
	return []pulumi.StringOutput{createdWorkerRoute.Pattern}, nil
}
