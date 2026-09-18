package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// loadBalancer provisions the load balancer and exports its outputs.
func loadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.LoadBalancer, error) {
	spec := locals.DigitalOceanLoadBalancer.Spec

	args := &digitalocean.LoadBalancerArgs{
		Name:                         pulumi.String(spec.LoadBalancerName),
		RedirectHttpToHttps:          pulumi.Bool(spec.RedirectHttpToHttps),
		EnableProxyProtocol:          pulumi.Bool(spec.EnableProxyProtocol),
		EnableBackendKeepalive:       pulumi.Bool(spec.EnableBackendKeepalive),
		DisableLetsEncryptDnsRecords: pulumi.Bool(spec.DisableLetsEncryptDnsRecords),
	}

	if spec.Region != 0 {
		args.Region = pulumi.StringPtr(spec.Region.String())
	}
	if spec.Type != "" {
		args.Type = pulumi.StringPtr(spec.Type)
	}
	if spec.Vpc != nil && spec.Vpc.GetValue() != "" {
		args.VpcUuid = pulumi.StringPtr(spec.Vpc.GetValue())
	}
	// VPC subnet placement (the spec's CEL rule requires vpc alongside it) and
	// the bring-your-own address, both create-time and both sent only when
	// the manifest states them -- the Terraform module's null coalescing.
	if spec.SubnetUuid != "" {
		args.SubnetUuid = pulumi.StringPtr(spec.SubnetUuid)
	}
	if spec.Ip != "" {
		args.Ip = pulumi.StringPtr(spec.Ip)
	}
	if spec.Size != "" {
		args.Size = pulumi.StringPtr(spec.Size)
	}
	if spec.SizeUnit > 0 {
		args.SizeUnit = pulumi.IntPtr(int(spec.SizeUnit))
	}
	if spec.HttpIdleTimeoutSeconds > 0 {
		args.HttpIdleTimeoutSeconds = pulumi.IntPtr(int(spec.HttpIdleTimeoutSeconds))
	}
	if spec.TlsCipherPolicy != "" {
		args.TlsCipherPolicy = pulumi.StringPtr(spec.TlsCipherPolicy)
	}
	if spec.Network != "" {
		args.Network = pulumi.StringPtr(spec.Network)
	}
	if spec.NetworkStack != "" {
		args.NetworkStack = pulumi.StringPtr(spec.NetworkStack)
	}
	if spec.ProjectId != "" {
		args.ProjectId = pulumi.StringPtr(spec.ProjectId)
	}
	if spec.DropletTag != "" {
		args.DropletTag = pulumi.StringPtr(spec.DropletTag)
	}

	if len(spec.DropletIds) > 0 {
		var dropletIds pulumi.IntArray
		for _, dropletID := range spec.DropletIds {
			id, err := strconv.Atoi(dropletID.GetValue())
			if err != nil {
				return nil, errors.Wrapf(err, "droplet_ids entry %q is not a numeric Droplet ID", dropletID.GetValue())
			}
			dropletIds = append(dropletIds, pulumi.Int(id))
		}
		args.DropletIds = dropletIds
	}

	if len(spec.TargetLoadBalancerIds) > 0 {
		var targets pulumi.StringArray
		for _, target := range spec.TargetLoadBalancerIds {
			if v := target.GetValue(); v != "" {
				targets = append(targets, pulumi.String(v))
			}
		}
		if len(targets) > 0 {
			args.TargetLoadBalancerIds = targets
		}
	}

	if len(spec.ForwardingRules) > 0 {
		var rules digitalocean.LoadBalancerForwardingRuleArray
		for _, fr := range spec.ForwardingRules {
			rule := digitalocean.LoadBalancerForwardingRuleArgs{
				EntryPort:      pulumi.Int(int(fr.EntryPort)),
				EntryProtocol:  pulumi.String(fr.EntryProtocol.String()),
				TargetPort:     pulumi.Int(int(fr.TargetPort)),
				TargetProtocol: pulumi.String(fr.TargetProtocol.String()),
				TlsPassthrough: pulumi.BoolPtr(fr.TlsPassthrough),
			}
			if fr.CertificateName != nil && fr.CertificateName.GetValue() != "" {
				rule.CertificateName = pulumi.StringPtr(fr.CertificateName.GetValue())
			}
			rules = append(rules, rule)
		}
		args.ForwardingRules = rules
	}

	if spec.HealthCheck != nil {
		hc := &digitalocean.LoadBalancerHealthcheckArgs{
			Port:     pulumi.Int(int(spec.HealthCheck.Port)),
			Protocol: pulumi.String(spec.HealthCheck.Protocol.String()),
		}
		if spec.HealthCheck.Path != "" {
			hc.Path = pulumi.StringPtr(spec.HealthCheck.Path)
		}
		if spec.HealthCheck.CheckIntervalSec > 0 {
			hc.CheckIntervalSeconds = pulumi.IntPtr(int(spec.HealthCheck.CheckIntervalSec))
		}
		if spec.HealthCheck.ResponseTimeoutSeconds > 0 {
			hc.ResponseTimeoutSeconds = pulumi.IntPtr(int(spec.HealthCheck.ResponseTimeoutSeconds))
		}
		if spec.HealthCheck.UnhealthyThreshold > 0 {
			hc.UnhealthyThreshold = pulumi.IntPtr(int(spec.HealthCheck.UnhealthyThreshold))
		}
		if spec.HealthCheck.HealthyThreshold > 0 {
			hc.HealthyThreshold = pulumi.IntPtr(int(spec.HealthCheck.HealthyThreshold))
		}
		args.Healthcheck = hc
	}

	if spec.StickySessions != nil {
		sticky := &digitalocean.LoadBalancerStickySessionsArgs{
			Type: pulumi.StringPtr(spec.StickySessions.Type),
		}
		if spec.StickySessions.CookieName != "" {
			sticky.CookieName = pulumi.StringPtr(spec.StickySessions.CookieName)
		}
		if spec.StickySessions.CookieTtlSeconds > 0 {
			sticky.CookieTtlSeconds = pulumi.IntPtr(int(spec.StickySessions.CookieTtlSeconds))
		}
		args.StickySessions = sticky
	}

	if spec.Firewall != nil {
		fw := &digitalocean.LoadBalancerFirewallArgs{}
		if len(spec.Firewall.Allow) > 0 {
			var allows pulumi.StringArray
			for _, a := range spec.Firewall.Allow {
				allows = append(allows, pulumi.String(a))
			}
			fw.Allows = allows
		}
		if len(spec.Firewall.Deny) > 0 {
			var denies pulumi.StringArray
			for _, d := range spec.Firewall.Deny {
				denies = append(denies, pulumi.String(d))
			}
			fw.Denies = denies
		}
		args.Firewall = fw
	}

	if len(spec.Domains) > 0 {
		var domains digitalocean.LoadBalancerDomainArray
		for _, d := range spec.Domains {
			domain := digitalocean.LoadBalancerDomainArgs{
				Name:      pulumi.String(d.Name),
				IsManaged: pulumi.BoolPtr(d.IsManaged),
			}
			if d.CertificateName != nil && d.CertificateName.GetValue() != "" {
				domain.CertificateName = pulumi.StringPtr(d.CertificateName.GetValue())
			}
			domains = append(domains, domain)
		}
		args.Domains = domains
	}

	if spec.GlbSettings != nil {
		glb := &digitalocean.LoadBalancerGlbSettingsArgs{
			TargetProtocol: pulumi.String(spec.GlbSettings.TargetProtocol),
			TargetPort:     pulumi.Int(int(spec.GlbSettings.TargetPort)),
		}
		if spec.GlbSettings.FailoverThreshold > 0 {
			glb.FailoverThreshold = pulumi.IntPtr(int(spec.GlbSettings.FailoverThreshold))
		}
		if len(spec.GlbSettings.RegionPriorities) > 0 {
			priorities := pulumi.IntMap{}
			for region, priority := range spec.GlbSettings.RegionPriorities {
				priorities[region] = pulumi.Int(int(priority))
			}
			glb.RegionPriorities = priorities
		}
		if spec.GlbSettings.Cdn != nil {
			glb.Cdn = &digitalocean.LoadBalancerGlbSettingsCdnArgs{
				IsEnabled: pulumi.BoolPtr(spec.GlbSettings.Cdn.IsEnabled),
			}
		}
		args.GlbSettings = glb
	}

	created, err := digitalocean.NewLoadBalancer(
		ctx,
		"load_balancer",
		args,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean load balancer")
	}

	ctx.Export(OpLoadBalancerId, created.ID())
	ctx.Export(OpIp, created.Ip)
	ctx.Export(OpUrn, created.LoadBalancerUrn)
	ctx.Export(OpIpv6, created.Ipv6)

	return created, nil
}
