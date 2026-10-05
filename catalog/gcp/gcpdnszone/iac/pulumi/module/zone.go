package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gcpdnszonev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdnszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dns"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpdnszonev1alpha1.GcpDnsZoneIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	return managedZone(ctx, locals, gcpProvider)
}

// managedZone provisions a Cloud DNS managed zone (`google_dns_managed_zone`).
// DNS records belong in the separate GcpDnsRecord kind.
func managedZone(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	target := locals.GcpDnsZone
	spec := target.Spec
	projectId := spec.ProjectId.GetValue()

	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("dns.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if projectId != "" {
		serviceArgs.Project = pulumi.String(projectId)
	}
	createdProjectService, err := projects.NewService(ctx,
		"dns-dns.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable dns.googleapis.com api")
	}

	managedZoneName := strings.ReplaceAll(target.Metadata.Name, ".", "-")

	dnsName := spec.DnsName
	if dnsName == "" {
		dnsName = fmt.Sprintf("%s.", target.Metadata.Name)
	}

	description := spec.Description
	if description == "" {
		description = fmt.Sprintf("managed-zone for %s", target.Metadata.Name)
	}

	visibility := spec.GetVisibility()
	if visibility == "" {
		visibility = "public"
	}

	args := &dns.ManagedZoneArgs{
		Name:         pulumi.String(managedZoneName),
		Description:  pulumi.String(description),
		DnsName:      pulumi.String(dnsName),
		Visibility:   pulumi.String(visibility),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
		ForceDestroy: pulumi.BoolPtr(spec.GetForceDestroy()),
	}
	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if projectId != "" {
		args.Project = pulumi.StringPtr(projectId)
	}

	// What destroy does to the zone shell: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep answering queries).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if spec.DnssecConfig != nil {
		dc := spec.DnssecConfig
		dnssecArgs := &dns.ManagedZoneDnssecConfigArgs{}
		if dc.State != nil {
			dnssecArgs.State = pulumi.StringPtr(dc.GetState())
		}
		if dc.NonExistence != "" {
			dnssecArgs.NonExistence = pulumi.StringPtr(dc.NonExistence)
		}
		if len(dc.DefaultKeySpecs) > 0 {
			specs := dns.ManagedZoneDnssecConfigDefaultKeySpecArray{}
			for _, ks := range dc.DefaultKeySpecs {
				keySpec := dns.ManagedZoneDnssecConfigDefaultKeySpecArgs{}
				if ks.Algorithm != "" {
					keySpec.Algorithm = pulumi.StringPtr(ks.Algorithm)
				}
				if ks.KeyLength > 0 {
					keySpec.KeyLength = pulumi.IntPtr(int(ks.KeyLength))
				}
				if ks.KeyType != "" {
					keySpec.KeyType = pulumi.StringPtr(ks.KeyType)
				}
				specs = append(specs, keySpec)
			}
			dnssecArgs.DefaultKeySpecs = specs
		}
		args.DnssecConfig = dnssecArgs
	}

	if spec.PrivateVisibilityConfig != nil {
		pvc := spec.PrivateVisibilityConfig
		pvcArgs := &dns.ManagedZonePrivateVisibilityConfigArgs{}
		if len(pvc.Networks) > 0 {
			networks := dns.ManagedZonePrivateVisibilityConfigNetworkArray{}
			for _, n := range pvc.Networks {
				networks = append(networks, dns.ManagedZonePrivateVisibilityConfigNetworkArgs{
					NetworkUrl: pulumi.String(n.NetworkUrl.GetValue()),
				})
			}
			pvcArgs.Networks = networks
		}
		if len(pvc.GkeClusters) > 0 {
			clusters := dns.ManagedZonePrivateVisibilityConfigGkeClusterArray{}
			for _, c := range pvc.GkeClusters {
				clusters = append(clusters, dns.ManagedZonePrivateVisibilityConfigGkeClusterArgs{
					GkeClusterName: pulumi.String(c.GkeClusterName.GetValue()),
				})
			}
			pvcArgs.GkeClusters = clusters
		}
		args.PrivateVisibilityConfig = pvcArgs
	}

	if spec.ForwardingConfig != nil && len(spec.ForwardingConfig.TargetNameServers) > 0 {
		servers := dns.ManagedZoneForwardingConfigTargetNameServerArray{}
		for _, s := range spec.ForwardingConfig.TargetNameServers {
			server := dns.ManagedZoneForwardingConfigTargetNameServerArgs{}
			if s.Ipv4Address != "" {
				server.Ipv4Address = pulumi.StringPtr(s.Ipv4Address)
			}
			// One address family per target (spec CEL enforces pre-deploy).
			if s.Ipv6Address != "" {
				server.Ipv6Address = pulumi.StringPtr(s.Ipv6Address)
			}
			if s.DomainName != "" {
				server.DomainName = pulumi.StringPtr(s.DomainName)
			}
			if s.ForwardingPath != "" {
				server.ForwardingPath = pulumi.StringPtr(s.ForwardingPath)
			}
			servers = append(servers, server)
		}
		args.ForwardingConfig = &dns.ManagedZoneForwardingConfigArgs{
			TargetNameServers: servers,
		}
	}

	if spec.PeeringConfig != nil && spec.PeeringConfig.TargetNetwork != nil {
		args.PeeringConfig = &dns.ManagedZonePeeringConfigArgs{
			TargetNetwork: &dns.ManagedZonePeeringConfigTargetNetworkArgs{
				NetworkUrl: pulumi.String(spec.PeeringConfig.TargetNetwork.GetValue()),
			},
		}
	}

	if spec.CloudLoggingConfig != nil {
		args.CloudLoggingConfig = &dns.ManagedZoneCloudLoggingConfigArgs{
			EnableLogging: pulumi.Bool(spec.CloudLoggingConfig.EnableLogging),
		}
	}

	createdManagedZone, err := dns.NewManagedZone(ctx,
		managedZoneName,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrapf(err, "failed to create managed zone for %s", dnsName)
	}

	ctx.Export(OpZoneId, createdManagedZone.ManagedZoneId)
	ctx.Export(OpZoneName, createdManagedZone.Name)
	ctx.Export(OpNameservers, createdManagedZone.NameServers)

	return nil
}
