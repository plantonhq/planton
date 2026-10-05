package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpartifactregistryrepov1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpartifactregistryrepo/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpArtifactRegistryRepo *gcpartifactregistryrepov1alpha1.GcpArtifactRegistryRepo
	GcpLabels               map[string]string
	RepositoryId            string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpartifactregistryrepov1alpha1.GcpArtifactRegistryRepoIacInput) *Locals {
	locals := &Locals{}
	locals.GcpArtifactRegistryRepo = iacInput.Target

	// The repository ID falls back to metadata.name — one honest identity,
	// never a format-derived suffix. Identical to the Terraform module.
	locals.RepositoryId = locals.GcpArtifactRegistryRepo.Spec.RepositoryId
	if locals.RepositoryId == "" {
		locals.RepositoryId = locals.GcpArtifactRegistryRepo.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpArtifactRegistryRepo.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.RepositoryId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpArtifactRegistryRepo.String())

	if locals.GcpArtifactRegistryRepo.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpArtifactRegistryRepo.Metadata.Org
	}
	if locals.GcpArtifactRegistryRepo.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpArtifactRegistryRepo.Metadata.Env
	}
	if locals.GcpArtifactRegistryRepo.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpArtifactRegistryRepo.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
