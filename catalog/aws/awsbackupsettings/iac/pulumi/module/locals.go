package module

import (
	"strconv"

	awsbackupsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbackupsettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbackupsettingsv1alpha1.AwsBackupSettings
	Spec   *awsbackupsettingsv1alpha1.AwsBackupSettingsSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbackupsettingsv1alpha1.AwsBackupSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// Neither settings resource supports tags at the provider - the
	// map exists so the module keeps the catalog-wide shape if AWS
	// ever adds them.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBackupSettings.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
