package module

import (
	"github.com/pkg/errors"
	awslambdav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awslambda/v1"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type createdFunction struct {
	Function     *awslambda.Function
	LogGroupName pulumi.StringInput
}

func lambdaFunction(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*createdFunction, error) {
	spec := locals.AwsLambda.Spec

	// Resolve basics
	functionName := locals.AwsLambda.Metadata.Name

	// Determine architecture
	var arch pulumi.StringInput = pulumi.String("x86_64")
	if spec.Architecture == awslambdav1.Architecture_ARM64 {
		arch = pulumi.String("arm64")
	}

	// Build environment variables map
	var envVars pulumi.StringMap
	if len(spec.Environment) > 0 {
		envVars = pulumi.ToStringMap(spec.Environment)
	}

	// Attach layers (literal values only for now)
	layerArns := valuefrom.ToStringArray(spec.LayerArns)

	// VPC config (subnets/security groups literals only for now)
	subnetIds := valuefrom.ToStringArray(spec.Subnets)
	sgIds := valuefrom.ToStringArray(spec.SecurityGroups)

	// Optional log group for consistency in exports and retention
	logGroupName := pulumi.String("/aws/lambda/" + functionName)
	var logGroup *cloudwatch.LogGroup
	var err error
	logGroup, err = cloudwatch.NewLogGroup(ctx, "log-group", &cloudwatch.LogGroupArgs{
		Name:            logGroupName,
		RetentionInDays: pulumi.Int(30),
		Tags:            pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CloudWatch log group")
	}

	// Common args
	args := &awslambda.FunctionArgs{
		Description: pulumi.String(spec.Description),
		Role:        pulumi.String(spec.RoleArn.GetValue()),
		Architectures: pulumi.StringArray{
			arch,
		},
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Runtime/handler for S3 code
	if spec.CodeSourceType == awslambdav1.CodeSourceType_CODE_SOURCE_TYPE_S3 {
		if spec.Runtime != "" {
			args.Runtime = pulumi.String(spec.Runtime)
		}
		if spec.Handler != "" {
			args.Handler = pulumi.String(spec.Handler)
		}
		if spec.S3.Bucket != "" && spec.S3.Key != "" {
			args.S3Bucket = pulumi.String(spec.S3.Bucket)
			args.S3Key = pulumi.String(spec.S3.Key)
			if spec.S3.ObjectVersion != "" {
				args.S3ObjectVersion = pulumi.String(spec.S3.ObjectVersion)
			}
		}
	}

	// Image code
	if spec.CodeSourceType == awslambdav1.CodeSourceType_CODE_SOURCE_TYPE_IMAGE {
		if spec.ImageUri != "" {
			args.ImageUri = pulumi.String(spec.ImageUri)
			// PackageType defaults to Zip; override to Image when using image
			args.PackageType = pulumi.String("Image")
		}
	}

	// Environment
	if envVars != nil {
		args.Environment = awslambda.FunctionEnvironmentArgs{
			Variables: envVars,
		}
	}

	// Memory/timeout
	if spec.MemoryMb != 0 {
		args.MemorySize = pulumi.Int(int(spec.MemoryMb))
	}
	if spec.TimeoutSeconds != 0 {
		args.Timeout = pulumi.Int(int(spec.TimeoutSeconds))
	}

	// KMS key for env encryption
	if spec.KmsKeyArn.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(spec.KmsKeyArn.GetValue())
	}

	// Layers
	if len(layerArns) > 0 {
		args.Layers = pulumi.ToStringArray(layerArns)
	}

	// VPC config when any of the sets is non-empty
	if len(subnetIds) > 0 || len(sgIds) > 0 {
		args.VpcConfig = &awslambda.FunctionVpcConfigArgs{
			SubnetIds:        pulumi.ToStringArray(subnetIds),
			SecurityGroupIds: pulumi.ToStringArray(sgIds),
		}
	}

	fn, err := awslambda.NewFunction(ctx, functionName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create lambda function")
	}

	// Export selected outputs
	ctx.Export(OpFunctionArn, fn.Arn)
	ctx.Export(OpFunctionName, fn.Name)
	ctx.Export(OpLogGroupName, logGroup.Name)
	ctx.Export(OpRoleArn, pulumi.String(spec.RoleArn.GetValue()))

	// layer arns as array
	if len(layerArns) > 0 {
		ctx.Export(OpLayerArns, pulumi.ToStringArray(layerArns))
	}

	return &createdFunction{
		Function:     fn,
		LogGroupName: logGroup.Name,
	}, nil
}
