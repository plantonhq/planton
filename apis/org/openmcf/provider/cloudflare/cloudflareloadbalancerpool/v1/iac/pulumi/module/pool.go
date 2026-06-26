package module

import (
	"github.com/pkg/errors"
	cloudflareloadbalancerpoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareloadbalancerpool/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// pool provisions the account-scoped Cloudflare Load Balancer pool and exports
// its outputs.
func pool(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.LoadBalancerPool, error) {
	spec := locals.CloudflareLoadBalancerPool.Spec

	var origins cloudflare.LoadBalancerPoolOriginArray
	for _, o := range spec.Origins {
		originArgs := cloudflare.LoadBalancerPoolOriginArgs{
			Name:    pulumi.String(o.Name),
			Address: pulumi.String(o.Address.GetValue()),
		}
		if o.Weight != nil {
			originArgs.Weight = pulumi.Float64Ptr(*o.Weight)
		}
		if o.Enabled != nil {
			originArgs.Enabled = pulumi.BoolPtr(*o.Enabled)
		}
		if o.FlattenCname != nil {
			originArgs.FlattenCname = pulumi.BoolPtr(*o.FlattenCname)
		}
		if o.Port > 0 {
			originArgs.Port = pulumi.IntPtr(int(o.Port))
		}
		if o.VirtualNetworkId != "" {
			originArgs.VirtualNetworkId = pulumi.StringPtr(o.VirtualNetworkId)
		}
		if o.HostHeader != "" {
			originArgs.Header = cloudflare.LoadBalancerPoolOriginHeaderArgs{
				Hosts: pulumi.StringArray{pulumi.String(o.HostHeader)},
			}
		}
		origins = append(origins, originArgs)
	}

	args := &cloudflare.LoadBalancerPoolArgs{
		AccountId: pulumi.String(spec.AccountId),
		Name:      pulumi.String(spec.Name),
		Origins:   origins,
	}

	if spec.Monitor != nil && spec.Monitor.GetValue() != "" {
		args.Monitor = pulumi.StringPtr(spec.Monitor.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Enabled != nil {
		args.Enabled = pulumi.BoolPtr(*spec.Enabled)
	}
	if spec.MinimumOrigins > 0 {
		args.MinimumOrigins = pulumi.IntPtr(int(spec.MinimumOrigins))
	}
	if spec.Latitude != nil {
		args.Latitude = pulumi.Float64Ptr(*spec.Latitude)
	}
	if spec.Longitude != nil {
		args.Longitude = pulumi.Float64Ptr(*spec.Longitude)
	}

	if len(spec.CheckRegions) > 0 {
		var regions pulumi.StringArray
		for _, r := range spec.CheckRegions {
			if r != cloudflareloadbalancerpoolv1.CloudflareLoadBalancerPoolCheckRegion_check_region_unspecified {
				regions = append(regions, pulumi.String(r.String()))
			}
		}
		if len(regions) > 0 {
			args.CheckRegions = regions
		}
	}

	if ls := spec.LoadShedding; ls != nil {
		lsArgs := &cloudflare.LoadBalancerPoolLoadSheddingArgs{
			DefaultPercent: pulumi.Float64Ptr(ls.DefaultPercent),
			SessionPercent: pulumi.Float64Ptr(ls.SessionPercent),
		}
		if ls.DefaultPolicy != "" {
			lsArgs.DefaultPolicy = pulumi.StringPtr(ls.DefaultPolicy)
		}
		if ls.SessionPolicy != "" {
			lsArgs.SessionPolicy = pulumi.StringPtr(ls.SessionPolicy)
		}
		args.LoadShedding = lsArgs
	}

	if os := spec.OriginSteering; os != nil && os.Policy != "" {
		args.OriginSteering = &cloudflare.LoadBalancerPoolOriginSteeringArgs{
			Policy: pulumi.StringPtr(os.Policy),
		}
	}

	if nf := spec.NotificationFilter; nf != nil {
		nfArgs := &cloudflare.LoadBalancerPoolNotificationFilterArgs{}
		if nf.Origin != nil {
			originFilter := &cloudflare.LoadBalancerPoolNotificationFilterOriginArgs{
				Disable: pulumi.BoolPtr(nf.Origin.Disable),
			}
			if nf.Origin.Healthy != nil {
				originFilter.Healthy = pulumi.BoolPtr(*nf.Origin.Healthy)
			}
			nfArgs.Origin = originFilter
		}
		if nf.Pool != nil {
			poolFilter := &cloudflare.LoadBalancerPoolNotificationFilterPoolArgs{
				Disable: pulumi.BoolPtr(nf.Pool.Disable),
			}
			if nf.Pool.Healthy != nil {
				poolFilter.Healthy = pulumi.BoolPtr(*nf.Pool.Healthy)
			}
			nfArgs.Pool = poolFilter
		}
		args.NotificationFilter = nfArgs
	}

	created, err := cloudflare.NewLoadBalancerPool(
		ctx,
		"pool",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare load balancer pool")
	}

	ctx.Export(OpPoolId, created.ID())
	ctx.Export(OpPoolName, created.Name)

	return created, nil
}
