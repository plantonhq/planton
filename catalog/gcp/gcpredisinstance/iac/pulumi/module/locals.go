package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpredisinstancev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpredisinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpRedisInstance  *gcpredisinstancev1alpha1.GcpRedisInstance
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpredisinstancev1alpha1.GcpRedisInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.GcpRedisInstance = iacInput.Target
	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpRedisInstance.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpRedisInstance.Spec.InstanceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpRedisInstance.String())

	if locals.GcpRedisInstance.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpRedisInstance.Metadata.Org
	}
	if locals.GcpRedisInstance.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpRedisInstance.Metadata.Env
	}
	if locals.GcpRedisInstance.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpRedisInstance.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
