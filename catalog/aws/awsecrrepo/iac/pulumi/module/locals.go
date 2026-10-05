package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsecrrepov1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecrrepo/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the AWS ECR Repo resource definition from the IaC input
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsEcrRepo *awsecrrepov1alpha1.AwsEcrRepo
	AwsTags    map[string]string
}

// initializeLocals is similar to Terraform "locals" usage. It reads
// values from AwsEcrRepoIacInput to build a Locals instance.
func initializeLocals(ctx *pulumi.Context, iacInput *awsecrrepov1alpha1.AwsEcrRepoIacInput) *Locals {
	locals := &Locals{}

	locals.AwsEcrRepo = iacInput.Target

	// No Name tag: the repository has a real name of its own
	// (spec.repository_name, a slash-namespaced registry path) — tagging
	// Name with the graph node's metadata.name would show a DIFFERENT value
	// than the repository's actual name in every console tag view. Kinds
	// whose AWS resource carries its own name omit the Name tag (the Global
	// Accelerator / SageMaker convention; the Terraform module matches).
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEcrRepo.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEcrRepo.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEcrRepo.String(),
		awstagkeys.ResourceId:   locals.AwsEcrRepo.Metadata.Id,
	}

	return locals
}
