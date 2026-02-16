package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/wafv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// logging creates a WAFv2 Web ACL Logging Configuration resource. This is a
// separate AWS resource that links logging to the Web ACL by ARN.
func logging(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, webAclResource *wafv2.WebAcl) error {
	loggingSpec := locals.WebAcl.Spec.Logging

	args := &wafv2.WebAclLoggingConfigurationArgs{
		ResourceArn:            webAclResource.Arn,
		LogDestinationConfigs: pulumi.StringArray{pulumi.String(loggingSpec.DestinationArn.GetValue())},
	}

	// Build redacted fields from the simplified spec fields.
	var redactedFields wafv2.WebAclLoggingConfigurationRedactedFieldArray

	for _, headerName := range loggingSpec.RedactedHeaderNames {
		redactedFields = append(redactedFields, &wafv2.WebAclLoggingConfigurationRedactedFieldArgs{
			SingleHeader: &wafv2.WebAclLoggingConfigurationRedactedFieldSingleHeaderArgs{
				Name: pulumi.String(headerName),
			},
		})
	}

	if loggingSpec.RedactUriPath {
		redactedFields = append(redactedFields, &wafv2.WebAclLoggingConfigurationRedactedFieldArgs{
			UriPath: &wafv2.WebAclLoggingConfigurationRedactedFieldUriPathArgs{},
		})
	}

	if loggingSpec.RedactQueryString {
		redactedFields = append(redactedFields, &wafv2.WebAclLoggingConfigurationRedactedFieldArgs{
			QueryString: &wafv2.WebAclLoggingConfigurationRedactedFieldQueryStringArgs{},
		})
	}

	if len(redactedFields) > 0 {
		args.RedactedFields = redactedFields
	}

	_, err := wafv2.NewWebAclLoggingConfiguration(
		ctx,
		locals.WebAcl.Metadata.Name+"-logging",
		args,
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{webAclResource}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create WAF logging configuration")
	}

	return nil
}
