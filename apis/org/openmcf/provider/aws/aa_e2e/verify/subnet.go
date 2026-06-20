package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// subnetVerifier verifies an AwsSubnet via DescribeSubnets. A deleted subnet
// returns the typed InvalidSubnetID.NotFound error, which is the "absent" signal;
// any other error is a genuine failure and must surface.
type subnetVerifier struct{}

func (*subnetVerifier) IDOutputKey() string { return "subnet_id" }

func (*subnetVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := subnetExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awssubnet verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awssubnet %q not found after deploy", id)
	}
	return nil
}

func (*subnetVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := subnetExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awssubnet verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awssubnet %q still exists after destroy", id)
	}
	return nil
}

func subnetExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{SubnetIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidSubnetID.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.Subnets) > 0, nil
}
