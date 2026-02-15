package module

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func outputs(ctx *pulumi.Context, stream *kinesis.FirehoseDeliveryStream) error {
	ctx.Export("delivery_stream_arn", stream.Arn)
	ctx.Export("delivery_stream_name", stream.Name)
	return nil
}
