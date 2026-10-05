package module

import (
	"github.com/pkg/errors"
	awslambdav1alpha1 "github.com/plantonhq/planton/catalog/aws/awslambda/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Lambda function and its function-scoped
// satellites (aliases, provisioned concurrency, function URL, invoke
// permissions, async-invocation config, recursion and runtime-update
// management). The function composes onto its neighbors instead of
// embedding them: the IAM execution role, VPC subnets and security
// groups, KMS keys, the dead-letter queue, the EFS access point, and
// the CloudWatch log group all attach by reference -- this module never
// creates or mutates a resource that deserves to be its own node, and
// event sources attach through the separate event-source-mapping kind.
func Resources(ctx *pulumi.Context, iacInput *awslambdav1alpha1.AwsLambdaIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared
	// builder, which resolves the right credential mechanism (static
	// keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsLambda.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdFunction, err := function(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Lambda function")
	}

	aliasArns, createdFunctionUrl, err := satellites(ctx, locals, provider, createdFunction)
	if err != nil {
		return errors.Wrap(err, "failed to create Lambda function satellites")
	}

	ctx.Export(OpFunctionArn, createdFunction.Arn)
	ctx.Export(OpFunctionName, createdFunction.Name)
	ctx.Export(OpInvokeArn, createdFunction.InvokeArn)

	// Version outputs only carry values when publishing is on -- an
	// unpublished function has no version to reference, and exporting
	// the provider's placeholder would differ between engines.
	if locals.AwsLambda.Spec.Publish {
		ctx.Export(OpQualifiedArn, createdFunction.QualifiedArn)
		ctx.Export(OpVersion, createdFunction.Version)
	} else {
		ctx.Export(OpQualifiedArn, pulumi.String(""))
		ctx.Export(OpVersion, pulumi.String(""))
	}

	if createdFunctionUrl != nil {
		ctx.Export(OpFunctionUrl, createdFunctionUrl.FunctionUrl)
	} else {
		ctx.Export(OpFunctionUrl, pulumi.String(""))
	}

	ctx.Export(OpAliasArns, aliasArns)
	ctx.Export(OpLogGroupName, pulumi.String(locals.LogGroupName))

	return nil
}
