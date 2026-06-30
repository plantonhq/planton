package module

import (
	"strings"

	alicloudcdndomainv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudcdndomain/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudCdnDomain *alicloudcdndomainv1.AliCloudCdnDomain
	Tags              map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudcdndomainv1.AliCloudCdnDomainStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudCdnDomain = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudCdnDomain.String()),
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
