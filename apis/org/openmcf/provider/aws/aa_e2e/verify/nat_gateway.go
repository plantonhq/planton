package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// natGatewayVerifier verifies an AwsNatGateway via DescribeNatGateways. AWS does
// not delete NAT gateway records immediately; a destroyed gateway lingers in the
// "deleted" state for a while and only later returns the typed
// NatGatewayNotFound error. Both mean "absent" for verification purposes.
type natGatewayVerifier struct{}

func (*natGatewayVerifier) IDOutputKey() string { return "nat_gateway_id" }

func (*natGatewayVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := natGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsnatgateway verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsnatgateway %q not found after deploy", id)
	}
	return nil
}

func (*natGatewayVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := natGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsnatgateway verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsnatgateway %q still exists after destroy", id)
	}
	return nil
}

// natGatewayExists reports whether the gateway is present and not in a
// deleting/deleted state. A NatGatewayNotFound error is treated as absent.
func natGatewayExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{NatGatewayIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NatGatewayNotFound" {
			return false, nil
		}
		return false, err
	}
	for _, ngw := range out.NatGateways {
		switch ngw.State {
		case awstypes.NatGatewayStateDeleting, awstypes.NatGatewayStateDeleted:
			continue
		default:
			return true, nil
		}
	}
	return false, nil
}
