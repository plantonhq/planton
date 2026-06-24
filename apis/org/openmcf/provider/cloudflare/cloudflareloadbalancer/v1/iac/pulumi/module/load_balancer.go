package module

import (
	"fmt"

	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// load_balancer provisions:
//  1. A Monitor (health check)
//  2. A Pool (collection of origins)
//  3. The Load Balancer itself
//
// It also exports stack outputs defined in outputs.go.
func load_balancer(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.LoadBalancer, error) {

	resourceName := locals.CloudflareLoadBalancer.Metadata.Name

	// The load balancer is zone-scoped, while the pool and monitor are
	// account-scoped. Derive the account id from the zone so the spec only
	// needs the zone reference.
	zoneId := locals.CloudflareLoadBalancer.Spec.ZoneId.GetValue()
	zone, err := cloudflare.LookupZone(ctx, &cloudflare.LookupZoneArgs{
		ZoneId: pulumi.StringRef(zoneId),
	}, pulumi.Provider(cloudflareProvider))
	if err != nil {
		return nil, fmt.Errorf("failed to look up zone for account id: %w", err)
	}
	accountId := zone.Account.Id

	// ---------------------------------------------------------------------
	// 1. Monitor (health check) – uses the health_probe_path from the spec.
	// ---------------------------------------------------------------------
	createdMonitor, err := cloudflare.NewLoadBalancerMonitor(ctx, "monitor",
		&cloudflare.LoadBalancerMonitorArgs{
			AccountId:     pulumi.String(accountId),
			Type:          pulumi.String("https"),
			Method:        pulumi.String("GET"),
			Path:          pulumi.String(locals.CloudflareLoadBalancer.Spec.HealthProbePath),
			ExpectedCodes: pulumi.String("2xx"),
			Timeout:       pulumi.Int(5),
			Interval:      pulumi.Int(60),
			Retries:       pulumi.Int(2),
			Description:   pulumi.String(fmt.Sprintf("Health check for %s", resourceName)),
		}, pulumi.Provider(cloudflareProvider))
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	// ---------------------------------------------------------------------
	// 2. Pool – one pool containing all declared origins.
	// ---------------------------------------------------------------------
	var poolOrigins cloudflare.LoadBalancerPoolOriginArray
	for _, o := range locals.CloudflareLoadBalancer.Spec.Origins {
		poolOrigins = append(poolOrigins, cloudflare.LoadBalancerPoolOriginArgs{
			Name:    pulumi.String(o.Name),
			Address: pulumi.String(o.Address),
			Enabled: pulumi.Bool(true),
			Weight:  pulumi.Float64Ptr(float64(o.Weight)),
		})
	}

	createdPool, err := cloudflare.NewLoadBalancerPool(ctx, "pool", &cloudflare.LoadBalancerPoolArgs{
		AccountId:   pulumi.String(accountId),
		Name:        pulumi.String(resourceName + "-pool"),
		Origins:     poolOrigins,
		Monitor:     createdMonitor.ID(),
		Enabled:     pulumi.Bool(true),
		Description: pulumi.String(fmt.Sprintf("Pool for %s", resourceName)),
	}, pulumi.Provider(cloudflareProvider))
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// ---------------------------------------------------------------------
	// 3. Load Balancer – wires everything together.
	// ---------------------------------------------------------------------
	// Get steering policy and session affinity directly from enum strings.
	// Enum values match Cloudflare API expected strings.
	steering := pulumi.StringPtr(locals.CloudflareLoadBalancer.Spec.SteeringPolicy.String())
	affinity := pulumi.StringPtr(locals.CloudflareLoadBalancer.Spec.SessionAffinity.String())

	createdLoadBalancer, err := cloudflare.NewLoadBalancer(ctx, "load_balancer", &cloudflare.LoadBalancerArgs{
		ZoneId:          pulumi.String(zoneId),
		Name:            pulumi.String(locals.CloudflareLoadBalancer.Spec.Hostname),
		DefaultPools:    pulumi.StringArray{createdPool.ID()},
		FallbackPool:    createdPool.ID(),
		Proxied:         pulumi.BoolPtr(locals.CloudflareLoadBalancer.Spec.Proxied),
		SteeringPolicy:  steering,
		SessionAffinity: affinity,
		Description:     pulumi.String(fmt.Sprintf("Load balancer for %s", locals.CloudflareLoadBalancer.Spec.Hostname)),
	}, pulumi.Provider(cloudflareProvider))
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer: %w", err)
	}

	// ---------------------------------------------------------------------
	// 4. Stack outputs.
	// ---------------------------------------------------------------------
	ctx.Export(OpLoadBalancerId, createdLoadBalancer.ID())
	ctx.Export(OpLoadBalancerDnsRecordName, pulumi.String(locals.CloudflareLoadBalancer.Spec.Hostname))
	ctx.Export(OpLoadBalancerCnameTarget, createdLoadBalancer.ID())

	return createdLoadBalancer, nil
}
