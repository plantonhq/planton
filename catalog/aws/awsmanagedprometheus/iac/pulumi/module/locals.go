package module

import (
	"strconv"
	"strings"

	awsmanagedprometheusv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmanagedprometheus/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsmanagedprometheusv1alpha1.AwsManagedPrometheus
	Spec   *awsmanagedprometheusv1alpha1.AwsManagedPrometheusSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsmanagedprometheusv1alpha1.AwsManagedPrometheusIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key
	// (applied to the workspace, rule group namespaces, and anomaly
	// detectors; the configuration/alert-manager/query-logging/
	// resource-policy satellites are untaggable at AWS).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsManagedPrometheus.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}

// wildcardLogGroupArn appends the ":*" suffix AWS requires on AMP's
// log-group ARN fields when absent. The log group resource (and the
// AwsCloudwatchLogGroup kind's output) exports the bare ARN - the
// module owns that quirk so specs wire the natural output.
func wildcardLogGroupArn(arn string) string {
	if strings.HasSuffix(arn, ":*") {
		return arn
	}
	return arn + ":*"
}
