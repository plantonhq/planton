package module

import (
	"github.com/pkg/errors"
	cloudflareloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflareloadbalancer/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// load_balancer provisions the zone-scoped Cloudflare Load Balancer, wiring it to
// account-scoped pools referenced by ID/reference. Pools and their monitors are
// separate resources (CloudflareLoadBalancerPool / CloudflareLoadBalancerMonitor).
func load_balancer(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.LoadBalancer, error) {
	spec := locals.CloudflareLoadBalancer.Spec

	var defaultPools pulumi.StringArray
	for _, p := range spec.DefaultPools {
		defaultPools = append(defaultPools, pulumi.String(p.GetValue()))
	}

	args := &cloudflare.LoadBalancerArgs{
		ZoneId:       pulumi.String(spec.ZoneId.GetValue()),
		Name:         pulumi.String(spec.Hostname),
		DefaultPools: defaultPools,
		FallbackPool: pulumi.String(spec.FallbackPool.GetValue()),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Proxied != nil {
		args.Proxied = pulumi.BoolPtr(*spec.Proxied)
	}
	if spec.Enabled != nil {
		args.Enabled = pulumi.BoolPtr(*spec.Enabled)
	}
	if spec.Ttl > 0 {
		args.Ttl = pulumi.Float64Ptr(spec.Ttl)
	}
	if spec.SessionAffinity != cloudflareloadbalancerv1.CloudflareLoadBalancerSessionAffinity_none {
		args.SessionAffinity = pulumi.StringPtr(spec.SessionAffinity.String())
	}
	if spec.SteeringPolicy != cloudflareloadbalancerv1.CloudflareLoadBalancerSteeringPolicy_off {
		args.SteeringPolicy = pulumi.StringPtr(spec.SteeringPolicy.String())
	}
	if spec.SessionAffinityTtl > 0 {
		args.SessionAffinityTtl = pulumi.Float64Ptr(spec.SessionAffinityTtl)
	}

	if saa := spec.SessionAffinityAttributes; saa != nil {
		saaArgs := &cloudflare.LoadBalancerSessionAffinityAttributesArgs{
			RequireAllHeaders: pulumi.BoolPtr(saa.RequireAllHeaders),
		}
		if saa.DrainDuration > 0 {
			saaArgs.DrainDuration = pulumi.Float64Ptr(saa.DrainDuration)
		}
		if len(saa.Headers) > 0 {
			saaArgs.Headers = pulumi.ToStringArray(saa.Headers)
		}
		if saa.Samesite != "" {
			saaArgs.Samesite = pulumi.StringPtr(saa.Samesite)
		}
		if saa.Secure != "" {
			saaArgs.Secure = pulumi.StringPtr(saa.Secure)
		}
		if saa.ZeroDowntimeFailover != "" {
			saaArgs.ZeroDowntimeFailover = pulumi.StringPtr(saa.ZeroDowntimeFailover)
		}
		args.SessionAffinityAttributes = saaArgs
	}

	if m := geoPoolMap(spec.RegionPools); len(m) > 0 {
		args.RegionPools = m
	}
	if m := geoPoolMap(spec.CountryPools); len(m) > 0 {
		args.CountryPools = m
	}
	if m := geoPoolMap(spec.PopPools); len(m) > 0 {
		args.PopPools = m
	}

	if ar := spec.AdaptiveRouting; ar != nil {
		args.AdaptiveRouting = &cloudflare.LoadBalancerAdaptiveRoutingArgs{
			FailoverAcrossPools: pulumi.BoolPtr(ar.FailoverAcrossPools),
		}
	}
	if ls := spec.LocationStrategy; ls != nil && (ls.Mode != "" || ls.PreferEcs != "") {
		lsArgs := &cloudflare.LoadBalancerLocationStrategyArgs{}
		if ls.Mode != "" {
			lsArgs.Mode = pulumi.StringPtr(ls.Mode)
		}
		if ls.PreferEcs != "" {
			lsArgs.PreferEcs = pulumi.StringPtr(ls.PreferEcs)
		}
		args.LocationStrategy = lsArgs
	}
	if rs := spec.RandomSteering; rs != nil {
		rsArgs := &cloudflare.LoadBalancerRandomSteeringArgs{}
		if rs.DefaultWeight > 0 {
			rsArgs.DefaultWeight = pulumi.Float64Ptr(rs.DefaultWeight)
		}
		if len(rs.PoolWeights) > 0 {
			weights := pulumi.Float64Map{}
			for k, v := range rs.PoolWeights {
				weights[k] = pulumi.Float64(v)
			}
			rsArgs.PoolWeights = weights
		}
		args.RandomSteering = rsArgs
	}

	created, err := cloudflare.NewLoadBalancer(
		ctx,
		"load_balancer",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare load balancer")
	}

	ctx.Export(OpLoadBalancerId, created.ID())
	ctx.Export(OpLoadBalancerDnsRecordName, created.Name)
	// The CNAME target for a Cloudflare load balancer is its hostname (clients
	// point their DNS at it); it is not the opaque load-balancer ID.
	ctx.Export(OpLoadBalancerCnameTarget, created.Name)

	return created, nil
}

// geoPoolMap converts a list of geo-pool mappings into the provider's
// map[code] -> ordered pool IDs shape.
func geoPoolMap(entries []*cloudflareloadbalancerv1.CloudflareLoadBalancerGeoPools) pulumi.StringArrayMap {
	m := pulumi.StringArrayMap{}
	for _, gp := range entries {
		var ids pulumi.StringArray
		for _, p := range gp.PoolIds {
			ids = append(ids, pulumi.String(p.GetValue()))
		}
		m[gp.Code] = ids
	}
	return m
}
