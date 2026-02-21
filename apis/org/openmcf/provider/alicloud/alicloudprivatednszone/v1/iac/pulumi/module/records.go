package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudprivatednszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudprivatednszone/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/pvtz"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func zoneRecord(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	zone *pvtz.Zone,
	zoneName string,
	index int,
	record *alicloudprivatednszonev1.AliCloudPrivateDnsZoneRecord,
) error {
	resourceName := fmt.Sprintf("%s-%s-%s-%d", zoneName, record.Rr, record.Type, index)

	args := &pvtz.ZoneRecordArgs{
		ZoneId: zone.ID(),
		Rr:     pulumi.String(record.Rr),
		Type:   pulumi.String(record.Type),
		Value:  pulumi.String(record.Value),
	}

	if record.Ttl > 0 {
		args.Ttl = pulumi.Int(int(record.Ttl))
	}

	if record.Priority > 0 {
		args.Priority = pulumi.Int(int(record.Priority))
	}

	if record.Remark != "" {
		args.Remark = pulumi.String(record.Remark)
	}

	_, err := pvtz.NewZoneRecord(ctx, resourceName, args,
		pulumi.Provider(provider),
		pulumi.Parent(zone),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create private zone record %s (type %s) in zone %s", record.Rr, record.Type, zoneName)
	}

	return nil
}
