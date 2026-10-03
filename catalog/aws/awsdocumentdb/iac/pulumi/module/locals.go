package module

import (
	"strconv"
	"strings"

	awsdocumentdbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsdocumentdb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsDocumentDb *awsdocumentdbv1alpha1.AwsDocumentDb

	// ClusterIdentifier is metadata.name -- create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on either.
	ClusterIdentifier string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsdocumentdbv1alpha1.AwsDocumentDbIacInput) *Locals {
	locals := &Locals{}
	locals.AwsDocumentDb = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.ClusterIdentifier = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsDocumentDb.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}

// engineFamily derives the cluster parameter-group family from the pinned
// engine_version (inline parameters require a pinned version,
// CEL-enforced, so this never sees an empty version): "5.0.0" -> docdb5.0.
// DocumentDB families are keyed by major.minor -- AWS's own family
// naming, not a convention of ours.
func engineFamily(engineVersion string) string {
	parts := strings.Split(engineVersion, ".")
	if len(parts) >= 2 {
		return "docdb" + parts[0] + "." + parts[1]
	}
	return "docdb" + parts[0]
}
