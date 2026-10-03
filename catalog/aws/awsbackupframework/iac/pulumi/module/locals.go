package module

import (
	"strconv"

	awsbackupframeworkv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbackupframework/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbackupframeworkv1alpha1.AwsBackupFramework
	Spec   *awsbackupframeworkv1alpha1.AwsBackupFrameworkSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbackupframeworkv1alpha1.AwsBackupFrameworkIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBackupFramework.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
