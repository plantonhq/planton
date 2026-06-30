package module

import (
	ociobjectstoragebucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociobjectstoragebucket/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciObjectStorageBucket *ociobjectstoragebucketv1.OciObjectStorageBucket
	BucketName             string
	FreeformTags           map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ociobjectstoragebucketv1.OciObjectStorageBucketStackInput) *Locals {
	locals := &Locals{}
	locals.OciObjectStorageBucket = stackInput.Target

	locals.BucketName = stackInput.Target.Spec.Name

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciObjectStorageBucket.String(),
		"resource_id":   stackInput.Target.Metadata.Id,
	}
	if stackInput.Target.Metadata.Org != "" {
		locals.FreeformTags["organization"] = stackInput.Target.Metadata.Org
	}
	if stackInput.Target.Metadata.Env != "" {
		locals.FreeformTags["environment"] = stackInput.Target.Metadata.Env
	}
	for k, v := range stackInput.Target.Metadata.Labels {
		locals.FreeformTags[k] = v
	}

	return locals
}
