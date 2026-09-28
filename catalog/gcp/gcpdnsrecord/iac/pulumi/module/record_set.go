package module

import (
	"github.com/pkg/errors"
	gcpdnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdnsrecord/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// recordSet provisions one DNS record set (`google_dns_record_set`): a
// (name, type) pair answered either with static rrdatas or with exactly one
// routing policy — never both (the provider enforces ExactlyOneOf; the spec
// enforces the same rule pre-deploy).
func recordSet(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpDnsRecord.Spec

	args := &dns.RecordSetArgs{
		ManagedZone: pulumi.String(locals.ManagedZone),
		Name:        pulumi.String(locals.Name),
		Type:        pulumi.String(locals.RecordType),
		Ttl:         pulumi.IntPtr(locals.TtlSeconds),
	}

	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}

	// Static values arm. Left nil when routing_policy is used so the
	// provider's ExactlyOneOf sees the attribute as absent.
	if len(locals.Values) > 0 {
		args.Rrdatas = pulumi.ToStringArray(locals.Values)
	}

	if spec.RoutingPolicy != nil {
		args.RoutingPolicy = buildRoutingPolicy(spec.RoutingPolicy)
	}

	// What destroy does to the record set: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep answering queries).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdRecordSet, err := dns.NewRecordSet(ctx,
		locals.GcpDnsRecord.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS record %s", locals.Name)
	}

	ctx.Export(OpFqdn, createdRecordSet.Name)
	ctx.Export(OpRecordType, createdRecordSet.Type)
	ctx.Export(OpManagedZone, createdRecordSet.ManagedZone)
	ctx.Export(OpProjectId, createdRecordSet.Project)
	ctx.Export(OpTtlSeconds, createdRecordSet.Ttl)

	return nil
}

// buildRoutingPolicy maps the spec's routing policy onto the provider's
// shape. The bridged SDK duplicates the health-checked-targets structs per
// arm, so each arm has its own small builder below.
func buildRoutingPolicy(policy *gcpdnsrecordv1alpha1.GcpDnsRecordRoutingPolicy) *dns.RecordSetRoutingPolicyArgs {
	args := &dns.RecordSetRoutingPolicyArgs{}

	// Geo fencing applies only to geolocation routing; the spec rejects it
	// for other styles, so passing it through unconditionally is safe.
	if policy.EnableGeoFencing {
		args.EnableGeoFencing = pulumi.BoolPtr(true)
	}

	if policy.HealthCheck.GetValue() != "" {
		args.HealthCheck = pulumi.StringPtr(policy.HealthCheck.GetValue())
	}

	if len(policy.Wrr) > 0 {
		wrrs := dns.RecordSetRoutingPolicyWrrArray{}
		for _, item := range policy.Wrr {
			wrrArgs := dns.RecordSetRoutingPolicyWrrArgs{
				Weight:  pulumi.Float64(item.GetWeight()),
				Rrdatas: pulumi.ToStringArray(stringValues(item.Values)),
			}
			if item.HealthCheckedTargets != nil {
				wrrArgs.HealthCheckedTargets = &dns.RecordSetRoutingPolicyWrrHealthCheckedTargetsArgs{
					ExternalEndpoints:     pulumi.ToStringArray(item.HealthCheckedTargets.ExternalEndpoints),
					InternalLoadBalancers: buildWrrInternalLoadBalancers(item.HealthCheckedTargets.InternalLoadBalancers),
				}
			}
			wrrs = append(wrrs, wrrArgs)
		}
		args.Wrrs = wrrs
	}

	if len(policy.Geo) > 0 {
		geos := dns.RecordSetRoutingPolicyGeoArray{}
		for _, item := range policy.Geo {
			geoArgs := dns.RecordSetRoutingPolicyGeoArgs{
				Location: pulumi.String(item.Location),
				Rrdatas:  pulumi.ToStringArray(stringValues(item.Values)),
			}
			if item.HealthCheckedTargets != nil {
				geoArgs.HealthCheckedTargets = &dns.RecordSetRoutingPolicyGeoHealthCheckedTargetsArgs{
					ExternalEndpoints:     pulumi.ToStringArray(item.HealthCheckedTargets.ExternalEndpoints),
					InternalLoadBalancers: buildGeoInternalLoadBalancers(item.HealthCheckedTargets.InternalLoadBalancers),
				}
			}
			geos = append(geos, geoArgs)
		}
		args.Geos = geos
	}

	if policy.PrimaryBackup != nil {
		pb := policy.PrimaryBackup

		primaryArgs := &dns.RecordSetRoutingPolicyPrimaryBackupPrimaryArgs{
			ExternalEndpoints:     pulumi.ToStringArray(pb.Primary.ExternalEndpoints),
			InternalLoadBalancers: buildPrimaryInternalLoadBalancers(pb.Primary.InternalLoadBalancers),
		}

		backupGeos := dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoArray{}
		for _, item := range pb.BackupGeo {
			backupGeoArgs := dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoArgs{
				Location: pulumi.String(item.Location),
				Rrdatas:  pulumi.ToStringArray(stringValues(item.Values)),
			}
			if item.HealthCheckedTargets != nil {
				backupGeoArgs.HealthCheckedTargets = &dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoHealthCheckedTargetsArgs{
					ExternalEndpoints:     pulumi.ToStringArray(item.HealthCheckedTargets.ExternalEndpoints),
					InternalLoadBalancers: buildBackupGeoInternalLoadBalancers(item.HealthCheckedTargets.InternalLoadBalancers),
				}
			}
			backupGeos = append(backupGeos, backupGeoArgs)
		}

		pbArgs := &dns.RecordSetRoutingPolicyPrimaryBackupArgs{
			Primary:    primaryArgs,
			BackupGeos: backupGeos,
		}
		if pb.TrickleRatio != nil {
			pbArgs.TrickleRatio = pulumi.Float64Ptr(pb.GetTrickleRatio())
		}
		if pb.EnableGeoFencingForBackups {
			pbArgs.EnableGeoFencingForBackups = pulumi.BoolPtr(true)
		}
		args.PrimaryBackup = pbArgs
	}

	return args
}

func buildWrrInternalLoadBalancers(targets []*gcpdnsrecordv1alpha1.GcpDnsRecordInternalLoadBalancerTarget) dns.RecordSetRoutingPolicyWrrHealthCheckedTargetsInternalLoadBalancerArray {
	result := dns.RecordSetRoutingPolicyWrrHealthCheckedTargetsInternalLoadBalancerArray{}
	for _, target := range targets {
		item := dns.RecordSetRoutingPolicyWrrHealthCheckedTargetsInternalLoadBalancerArgs{
			IpAddress:  pulumi.String(target.IpAddress.GetValue()),
			IpProtocol: pulumi.String(target.IpProtocol),
			NetworkUrl: pulumi.String(target.NetworkUrl.GetValue()),
			Port:       pulumi.String(target.Port),
			Project:    pulumi.String(target.Project.GetValue()),
		}
		if target.LoadBalancerType != "" {
			item.LoadBalancerType = pulumi.StringPtr(target.LoadBalancerType)
		}
		if target.Region != "" {
			item.Region = pulumi.StringPtr(target.Region)
		}
		result = append(result, item)
	}
	return result
}

func buildGeoInternalLoadBalancers(targets []*gcpdnsrecordv1alpha1.GcpDnsRecordInternalLoadBalancerTarget) dns.RecordSetRoutingPolicyGeoHealthCheckedTargetsInternalLoadBalancerArray {
	result := dns.RecordSetRoutingPolicyGeoHealthCheckedTargetsInternalLoadBalancerArray{}
	for _, target := range targets {
		item := dns.RecordSetRoutingPolicyGeoHealthCheckedTargetsInternalLoadBalancerArgs{
			IpAddress:  pulumi.String(target.IpAddress.GetValue()),
			IpProtocol: pulumi.String(target.IpProtocol),
			NetworkUrl: pulumi.String(target.NetworkUrl.GetValue()),
			Port:       pulumi.String(target.Port),
			Project:    pulumi.String(target.Project.GetValue()),
		}
		if target.LoadBalancerType != "" {
			item.LoadBalancerType = pulumi.StringPtr(target.LoadBalancerType)
		}
		if target.Region != "" {
			item.Region = pulumi.StringPtr(target.Region)
		}
		result = append(result, item)
	}
	return result
}

func buildPrimaryInternalLoadBalancers(targets []*gcpdnsrecordv1alpha1.GcpDnsRecordInternalLoadBalancerTarget) dns.RecordSetRoutingPolicyPrimaryBackupPrimaryInternalLoadBalancerArray {
	result := dns.RecordSetRoutingPolicyPrimaryBackupPrimaryInternalLoadBalancerArray{}
	for _, target := range targets {
		item := dns.RecordSetRoutingPolicyPrimaryBackupPrimaryInternalLoadBalancerArgs{
			IpAddress:  pulumi.String(target.IpAddress.GetValue()),
			IpProtocol: pulumi.String(target.IpProtocol),
			NetworkUrl: pulumi.String(target.NetworkUrl.GetValue()),
			Port:       pulumi.String(target.Port),
			Project:    pulumi.String(target.Project.GetValue()),
		}
		if target.LoadBalancerType != "" {
			item.LoadBalancerType = pulumi.StringPtr(target.LoadBalancerType)
		}
		if target.Region != "" {
			item.Region = pulumi.StringPtr(target.Region)
		}
		result = append(result, item)
	}
	return result
}

func buildBackupGeoInternalLoadBalancers(targets []*gcpdnsrecordv1alpha1.GcpDnsRecordInternalLoadBalancerTarget) dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoHealthCheckedTargetsInternalLoadBalancerArray {
	result := dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoHealthCheckedTargetsInternalLoadBalancerArray{}
	for _, target := range targets {
		item := dns.RecordSetRoutingPolicyPrimaryBackupBackupGeoHealthCheckedTargetsInternalLoadBalancerArgs{
			IpAddress:  pulumi.String(target.IpAddress.GetValue()),
			IpProtocol: pulumi.String(target.IpProtocol),
			NetworkUrl: pulumi.String(target.NetworkUrl.GetValue()),
			Port:       pulumi.String(target.Port),
			Project:    pulumi.String(target.Project.GetValue()),
		}
		if target.LoadBalancerType != "" {
			item.LoadBalancerType = pulumi.StringPtr(target.LoadBalancerType)
		}
		if target.Region != "" {
			item.Region = pulumi.StringPtr(target.Region)
		}
		result = append(result, item)
	}
	return result
}

// stringValues resolves each value-or-reference entry to the string the
// provider receives (references are resolved to values before the module runs).
func stringValues(entries []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		values = append(values, entry.GetValue())
	}
	return values
}
