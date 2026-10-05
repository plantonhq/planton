package module

import (
	"strconv"
	"strings"

	awsmanagedprometheusscraperv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmanagedprometheusscraper/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsmanagedprometheusscraperv1alpha1.AwsManagedPrometheusScraper
	Spec   *awsmanagedprometheusscraperv1alpha1.AwsManagedPrometheusScraperSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsmanagedprometheusscraperv1alpha1.AwsManagedPrometheusScraperIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key
	// (the scraper is taggable; its logging-configuration satellite is
	// not).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsManagedPrometheusScraper.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}

// wildcardLogGroupArn appends the ":*" suffix AWS requires on the
// scraper's logging destination when absent. The log group resource
// (and the AwsCloudwatchLogGroup kind's output) exports the bare ARN -
// the module owns that quirk so specs wire the natural output.
func wildcardLogGroupArn(arn string) string {
	if strings.HasSuffix(arn, ":*") {
		return arn
	}
	return arn + ":*"
}
