package module

import (
	"fmt"

	awskinesisfirehose "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisfirehose/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deliveryStream creates the Kinesis Firehose delivery stream with the
// configured source, encryption, and destination.
func deliveryStream(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*kinesis.FirehoseDeliveryStream, error) {
	spec := locals.Spec

	args := &kinesis.FirehoseDeliveryStreamArgs{
		Name: pulumi.String(locals.DeliveryStreamName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// ---------------------------------------------------------------------------
	// Source configuration
	// ---------------------------------------------------------------------------

	if src := spec.KinesisStreamSource; src != nil {
		args.KinesisSourceConfiguration = &kinesis.FirehoseDeliveryStreamKinesisSourceConfigurationArgs{
			KinesisStreamArn: pulumi.String(src.StreamArn.GetValue()),
			RoleArn:          pulumi.String(src.RoleArn.GetValue()),
		}
	}

	// ---------------------------------------------------------------------------
	// Server-side encryption (Direct PUT only)
	// ---------------------------------------------------------------------------

	if spec.SseEnabled {
		sseArgs := &kinesis.FirehoseDeliveryStreamServerSideEncryptionArgs{
			Enabled: pulumi.Bool(true),
		}
		if spec.SseKmsKeyArn != nil {
			sseArgs.KeyType = pulumi.StringPtr("CUSTOMER_MANAGED_CMK")
			sseArgs.KeyArn = pulumi.StringPtr(spec.SseKmsKeyArn.GetValue())
		} else {
			sseArgs.KeyType = pulumi.StringPtr("AWS_OWNED_CMK")
		}
		args.ServerSideEncryption = sseArgs
	}

	// ---------------------------------------------------------------------------
	// Destination dispatch
	// ---------------------------------------------------------------------------

	switch dest := spec.DestinationConfig.(type) {
	case *awskinesisfirehose.AwsKinesisFirehoseSpec_ExtendedS3:
		args.Destination = pulumi.String("extended_s3")
		extS3Args, err := buildExtendedS3Args(dest.ExtendedS3, locals)
		if err != nil {
			return nil, fmt.Errorf("extended_s3 destination: %w", err)
		}
		args.ExtendedS3Configuration = extS3Args

	case *awskinesisfirehose.AwsKinesisFirehoseSpec_Opensearch:
		args.Destination = pulumi.String("opensearch")
		osArgs, err := buildOpenSearchArgs(dest.Opensearch, locals)
		if err != nil {
			return nil, fmt.Errorf("opensearch destination: %w", err)
		}
		args.OpensearchConfiguration = osArgs

	case *awskinesisfirehose.AwsKinesisFirehoseSpec_HttpEndpoint:
		args.Destination = pulumi.String("http_endpoint")
		httpArgs, err := buildHttpEndpointArgs(dest.HttpEndpoint, locals)
		if err != nil {
			return nil, fmt.Errorf("http_endpoint destination: %w", err)
		}
		args.HttpEndpointConfiguration = httpArgs

	case *awskinesisfirehose.AwsKinesisFirehoseSpec_Redshift:
		args.Destination = pulumi.String("redshift")
		rsArgs, err := buildRedshiftArgs(dest.Redshift, locals)
		if err != nil {
			return nil, fmt.Errorf("redshift destination: %w", err)
		}
		args.RedshiftConfiguration = rsArgs

	default:
		return nil, fmt.Errorf("unsupported destination type: %T", spec.DestinationConfig)
	}

	stream, err := kinesis.NewFirehoseDeliveryStream(ctx, locals.DeliveryStreamName, args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, fmt.Errorf("creating delivery stream: %w", err)
	}

	return stream, nil
}
