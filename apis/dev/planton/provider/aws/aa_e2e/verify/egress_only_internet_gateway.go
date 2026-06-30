package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// egressOnlyInternetGatewayVerifier verifies an AwsEgressOnlyInternetGateway via
// DescribeEgressOnlyInternetGateways. Unlike DescribeInternetGateways (which
// returns a typed InvalidInternetGatewayID.NotFound error for a missing id),
// DescribeEgressOnlyInternetGateways signals absence by returning an empty result
// set, so "absent" is treated as either an empty slice OR a typed not-found error.
type egressOnlyInternetGatewayVerifier struct{}

func (*egressOnlyInternetGatewayVerifier) IDOutputKey() string {
	return "egress_only_internet_gateway_id"
}

func (*egressOnlyInternetGatewayVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := egressOnlyInternetGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsegressonlyinternetgateway verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsegressonlyinternetgateway %q not found after deploy", id)
	}
	return nil
}

func (*egressOnlyInternetGatewayVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := egressOnlyInternetGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsegressonlyinternetgateway verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsegressonlyinternetgateway %q still exists after destroy", id)
	}
	return nil
}

func egressOnlyInternetGatewayExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeEgressOnlyInternetGateways(ctx, &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: []string{id},
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidEgressOnlyInternetGatewayId.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.EgressOnlyInternetGateways) > 0, nil
}
