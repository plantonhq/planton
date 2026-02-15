package module

import (
	"fmt"
	"strconv"

	awshttpapigatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awshttpapigateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awshttpapigatewayv1.AwsHttpApiGateway
	Spec    *awshttpapigatewayv1.AwsHttpApiGatewaySpec
	AwsTags map[string]string
	ApiName string
}

// integrationKey generates a deduplication key for an integration.
// Routes with the same key share one underlying API Gateway integration resource.
func integrationKey(integration *awshttpapigatewayv1.AwsHttpApiGatewayIntegration) string {
	return fmt.Sprintf("%s:%s:%s",
		integration.IntegrationType,
		integration.IntegrationUri.GetValue(),
		integration.PayloadFormatVersion,
	)
}

func initializeLocals(ctx *pulumi.Context, in *awshttpapigatewayv1.AwsHttpApiGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.ApiName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsHttpApiGateway.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
