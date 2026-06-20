package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// vpcVerifier verifies an AwsVpc via DescribeVpcs. It exists so an AwsVpc can be
// used as a deployed E2E prerequisite (e.g. for AwsSubnet) and confirmed live. A
// deleted VPC returns the typed InvalidVpcID.NotFound error (the "absent" signal).
type vpcVerifier struct{}

func (*vpcVerifier) IDOutputKey() string { return "vpc_id" }

func (*vpcVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := vpcExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsvpc verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsvpc %q not found after deploy", id)
	}
	return nil
}

func (*vpcVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := vpcExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsvpc verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsvpc %q still exists after destroy", id)
	}
	return nil
}

func vpcExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{VpcIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidVpcID.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.Vpcs) > 0, nil
}
