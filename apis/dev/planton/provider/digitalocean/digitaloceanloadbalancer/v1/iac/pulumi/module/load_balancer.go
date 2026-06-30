package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// loadBalancer provisions the Load Balancer itself and exports its outputs.
func loadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.LoadBalancer, error) {

	spec := locals.DigitalOceanLoadBalancer.Spec

	// 2. -----  Droplet IDs / Tag  ------------------------------------------
	var dropletIds pulumi.IntArray
	if len(spec.DropletIds) > 0 {
		for _, dropletId := range spec.DropletIds {
			if id, err := strconv.Atoi(dropletId.GetValue()); err == nil {
				dropletIds = append(dropletIds, pulumi.Int(id))
			}
		}
	}

	var dropletTag pulumi.StringPtrInput
	if spec.DropletTag != "" {
		dropletTag = pulumi.String(spec.DropletTag)
	}

	// 3. -----  Forwarding Rules  -------------------------------------------
	var forwardingRules digitalocean.LoadBalancerForwardingRuleArray
	for _, fr := range spec.ForwardingRules {
		rule := digitalocean.LoadBalancerForwardingRuleArgs{
			EntryPort:      pulumi.Int(int(fr.EntryPort)),
			EntryProtocol:  pulumi.String(fr.EntryProtocol.String()),
			TargetPort:     pulumi.Int(int(fr.TargetPort)),
			TargetProtocol: pulumi.String(fr.TargetProtocol.String()),
		}
		// Add certificate name for HTTPS termination if specified
		if fr.CertificateName != "" {
			rule.CertificateName = pulumi.StringPtr(fr.CertificateName)
		}
		forwardingRules = append(forwardingRules, rule)
	}

	// 4. -----  Health Check  ------------------------------------------------
	var healthcheck *digitalocean.LoadBalancerHealthcheckArgs
	if spec.HealthCheck != nil {
		healthcheck = &digitalocean.LoadBalancerHealthcheckArgs{
			Port:                 pulumi.Int(int(spec.HealthCheck.Port)),
			Protocol:             pulumi.String(spec.HealthCheck.Protocol.String()),
			Path:                 pulumi.StringPtr(spec.HealthCheck.Path),
			CheckIntervalSeconds: pulumi.IntPtr(int(spec.HealthCheck.CheckIntervalSec)),
		}
	}

	// 5. -----  Sticky Sessions  --------------------------------------------
	var stickySessions *digitalocean.LoadBalancerStickySessionsArgs
	if spec.EnableStickySessions {
		stickySessions = &digitalocean.LoadBalancerStickySessionsArgs{
			Type: pulumi.String("cookies"),
		}
	}

	// 6. -----  Build resource arguments  ------------------------------------
	args := &digitalocean.LoadBalancerArgs{
		Name:            pulumi.String(spec.LoadBalancerName),
		Region:          pulumi.String(spec.Region.String()),
		VpcUuid:         pulumi.String(spec.Vpc.GetValue()),
		DropletIds:      dropletIds,
		DropletTag:      dropletTag,
		ForwardingRules: forwardingRules,
		Healthcheck:     healthcheck,
		StickySessions:  stickySessions,
	}

	// 7. -----  Create the load balancer  ------------------------------------
	createdLoadBalancer, err := digitalocean.NewLoadBalancer(
		ctx,
		"load_balancer",
		args,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean load balancer")
	}

	// 8. -----  Export stack outputs  ----------------------------------------
	ctx.Export(OpLoadBalancerId, createdLoadBalancer.ID())
	ctx.Export(OpIp, createdLoadBalancer.Ip)
	ctx.Export(OpDnsName, createdLoadBalancer.Name) // DigitalOcean LB has no DNS field; use name as placeholder.

	return createdLoadBalancer, nil
}
