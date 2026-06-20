package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// internetGatewayVerifier verifies an AwsInternetGateway via DescribeInternetGateways.
// A deleted gateway returns the typed InvalidInternetGatewayID.NotFound error,
// which is the "absent" signal; any other error is a genuine failure and must surface.
type internetGatewayVerifier struct{}

func (*internetGatewayVerifier) IDOutputKey() string { return "internet_gateway_id" }

func (*internetGatewayVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := internetGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsinternetgateway verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awsinternetgateway %q not found after deploy", id)
	}
	return nil
}

func (*internetGatewayVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := internetGatewayExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awsinternetgateway verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awsinternetgateway %q still exists after destroy", id)
	}
	return nil
}

func internetGatewayExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	client := ec2.NewFromConfig(cfg, func(o *ec2.Options) {
		if region != "" {
			o.Region = region
		}
	})
	out, err := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{InternetGatewayIds: []string{id}})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidInternetGatewayID.NotFound" {
			return false, nil
		}
		return false, err
	}
	return len(out.InternetGateways) > 0, nil
}
