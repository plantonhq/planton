package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// elasticIpVerifier verifies an AwsElasticIp via DescribeAddresses. A released
// allocation returns the typed InvalidAllocationID.NotFound error, which is the
// "absent" signal; any other error is a genuine failure and must surface. This
// verifier exists so an AwsElasticIp can serve as a prerequisite of an
// AwsNatGateway scenario (a public NAT gateway needs an Elastic IP).
type elasticIpVerifier struct{}

func (*elasticIpVerifier) IDOutputKey() string { return "allocation_id" }

func (*elasticIpVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := elasticIpExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awselasticip verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awselasticip %q not found after deploy", id)
	}
	return nil
}

func (*elasticIpVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := elasticIpExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awselasticip verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awselasticip %q still exists after destroy", id)
	}
	return nil
}

func elasticIpExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{AllocationIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidAllocationID.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.Addresses) > 0, nil
}
