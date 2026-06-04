package module

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrun"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dns attaches a custom domain (if enabled) and creates the TXT record that
// Google Cloud Run expects for ownership verification.
func customDns(ctx *pulumi.Context,
	locals *Locals,
	createdService *cloudrunv2.Service,
	gcpProvider *gcp.Provider) error {

	// Skip entirely when dns.enabled is false or not set.
	if locals.GcpCloudRun.Spec.Dns == nil || !locals.GcpCloudRun.Spec.Dns.Enabled {
		return nil
	}
	if len(locals.GcpCloudRun.Spec.Dns.Hostnames) == 0 {
		return errors.New("dns.enabled is true but hostnames list is empty")
	}

	hostname := locals.GcpCloudRun.Spec.Dns.Hostnames[0]

	// Create the DomainMapping that binds <hostname> to the Cloud Run service.
	createdDomainMapping, err := cloudrun.NewDomainMapping(ctx,
		"domain-mapping",
		&cloudrun.DomainMappingArgs{
			Location: createdService.Location,
			Name:     pulumi.String(hostname),
			Metadata: &cloudrun.DomainMappingMetadataArgs{
				Labels: pulumi.ToStringMap(locals.GcpLabels),
			},
			Spec: &cloudrun.DomainMappingSpecArgs{
				RouteName: createdService.Name,
			},
		},
		pulumi.Parent(createdService),
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create domain mapping")
	}

	// Google returns a TXT record for domain verification at:
	//   statuses[0].resourceRecords[0].rrdata
	// Convert StringPtrOutput -> StringOutput with .Elem() so it satisfies StringInput.
	txtValue := createdDomainMapping.Statuses.
		Index(pulumi.Int(0)).
		ResourceRecords().
		Index(pulumi.Int(0)).
		Rrdata().
		Elem()

	// Provision the TXT record in Cloud DNS.
	_, err = dns.NewRecordSet(ctx,
		"domain-verification",
		&dns.RecordSetArgs{
			ManagedZone: pulumi.String(locals.GcpCloudRun.Spec.Dns.ManagedZone.GetValue()),
			Name:        pulumi.String(fmt.Sprintf("%s.", hostname)),
			Ttl:         pulumi.Int(300),
			Type:        pulumi.String("TXT"),
			Rrdatas: pulumi.StringArray{
				txtValue, // now a StringOutput, which implements pulumi.StringInput
			},
		},
		pulumi.Parent(createdDomainMapping),
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to add TXT verification record")
	}

	return nil
}
