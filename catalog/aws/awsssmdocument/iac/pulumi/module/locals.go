package module

import (
	"strconv"

	awsssmdocumentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsssmdocument/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsssmdocumentv1alpha1.AwsSsmDocument
	Spec   *awsssmdocumentv1alpha1.AwsSsmDocumentSpec

	// DocumentName is metadata.name on both engines (document names
	// allow letters, digits, underscores, hyphens, and periods -
	// hyphenated names fit). Changing it forces replacement.
	DocumentName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsssmdocumentv1alpha1.AwsSsmDocumentIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	locals.DocumentName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSsmDocument.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
