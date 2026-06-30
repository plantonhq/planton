package module

import (
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayobjectbucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayobjectbucket/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module.
//
// NOTE: Scaleway Object Storage uses key-value map tags (map[string]string)
// unlike other Scaleway resources that use flat string tags ([]string).
// This is because Object Storage uses the S3-compatible API which supports
// structured tags natively.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayObjectBucket   *scalewayobjectbucketv1.ScalewayObjectBucket

	// ScalewayTags is a key-value map for Object Storage tags.
	// Unlike other Scaleway resources which use flat "key=value" string
	// tags, Object Storage supports structured key-value tags via the
	// S3-compatible API.
	ScalewayTags map[string]string
}

// initializeLocals copies stack-input fields into the Locals struct and
// builds a reusable tag map. Tags use key-value format (not flat strings)
// because Scaleway Object Storage's S3-compatible API supports structured
// tags natively.
func initializeLocals(_ *pulumi.Context, stackInput *scalewayobjectbucketv1.ScalewayObjectBucketStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayObjectBucket = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Standard labels applied as Scaleway Object Storage tags.
	// Object Storage uses map[string]string tags (S3-compatible),
	// NOT flat "key=value" strings like other Scaleway resources.
	locals.ScalewayTags = map[string]string{
		scalewaylabelkeys.Resource:     strconv.FormatBool(true),
		scalewaylabelkeys.ResourceName: locals.ScalewayObjectBucket.Metadata.Name,
		scalewaylabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_ScalewayObjectBucket.String(),
	}

	if locals.ScalewayObjectBucket.Metadata.Org != "" {
		locals.ScalewayTags[scalewaylabelkeys.Organization] = locals.ScalewayObjectBucket.Metadata.Org
	}

	if locals.ScalewayObjectBucket.Metadata.Env != "" {
		locals.ScalewayTags[scalewaylabelkeys.Environment] = locals.ScalewayObjectBucket.Metadata.Env
	}

	if locals.ScalewayObjectBucket.Metadata.Id != "" {
		locals.ScalewayTags[scalewaylabelkeys.ResourceId] = locals.ScalewayObjectBucket.Metadata.Id
	}

	return locals
}
