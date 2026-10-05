package module

import (
	"strconv"

	awsredshiftserverlessnamespacev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftserverlessnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRedshiftServerlessNamespace *awsredshiftserverlessnamespacev1alpha1.AwsRedshiftServerlessNamespace

	// NamespaceName is metadata.name -- create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on
	// either.
	NamespaceName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsredshiftserverlessnamespacev1alpha1.AwsRedshiftServerlessNamespaceIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRedshiftServerlessNamespace = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.NamespaceName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRedshiftServerlessNamespace.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
