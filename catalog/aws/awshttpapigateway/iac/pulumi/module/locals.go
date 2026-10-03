package module

import (
	"strconv"

	awshttpapigatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awshttpapigateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/proto"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target  *awshttpapigatewayv1alpha1.AwsHttpApiGateway
	Spec    *awshttpapigatewayv1alpha1.AwsHttpApiGatewaySpec
	AwsTags map[string]string
	ApiName string
}

// dedupIntegrations returns the distinct integration objects across all
// routes, in first-appearance order, plus a per-route index into that list.
// Routes whose integration blocks are IDENTICAL across every field share one
// API Gateway integration resource; any difference (even a single
// request-parameter mapping) yields a separate integration. Whole-object
// equality keeps the dedup honest as the integration surface grows -- a
// partial key (type:uri:payload) would silently merge integrations that
// differ in the newer fields. The Terraform module dedups with the same
// whole-object rule.
func dedupIntegrations(routes []*awshttpapigatewayv1alpha1.AwsHttpApiGatewayRoute) (distinct []*awshttpapigatewayv1alpha1.AwsHttpApiGatewayIntegration, routeToIndex []int) {
	routeToIndex = make([]int, len(routes))
	for i, route := range routes {
		found := -1
		for j, existing := range distinct {
			if proto.Equal(existing, route.Integration) {
				found = j
				break
			}
		}
		if found == -1 {
			distinct = append(distinct, route.Integration)
			found = len(distinct) - 1
		}
		routeToIndex[i] = found
	}
	return distinct, routeToIndex
}

func initializeLocals(ctx *pulumi.Context, in *awshttpapigatewayv1alpha1.AwsHttpApiGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.ApiName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsHttpApiGateway.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
