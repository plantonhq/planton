package module

import (
	"strconv"

	awsopensearchserverlesscollectionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsopensearchserverlesscollection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsopensearchserverlesscollectionv1alpha1.AwsOpenSearchServerlessCollection
	Spec   *awsopensearchserverlesscollectionv1alpha1.AwsOpenSearchServerlessCollectionSpec

	// CollectionName is metadata.name -- the AWS collection name is
	// create-time immutable (3-32 chars, ^[a-z][0-9a-z-]+$) and is the
	// naming basis both engines share. The collection-scoped policies the
	// module renders (encryption, network, data access, retention) are all
	// named after it too, and their rules match exactly
	// "collection/<name>" / "index/<name>/..." -- one manifest owns one
	// collection and everything that makes it usable.
	CollectionName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsopensearchserverlesscollectionv1alpha1.AwsOpenSearchServerlessCollectionIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.CollectionName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsOpenSearchServerlessCollection.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
