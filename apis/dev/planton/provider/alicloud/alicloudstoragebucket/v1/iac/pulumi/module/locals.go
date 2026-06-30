package module

import (
	"strings"

	alicloudstoragebucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudstoragebucket/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudStorageBucket *alicloudstoragebucketv1.AliCloudStorageBucket
	Tags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudstoragebucketv1.AliCloudStorageBucketStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudStorageBucket = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudStorageBucket.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}
