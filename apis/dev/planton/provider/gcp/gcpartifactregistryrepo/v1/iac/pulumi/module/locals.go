package module

import (
	"strconv"
	"strings"

	gcpartifactregistryrepov1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpartifactregistryrepo/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpArtifactRegistryRepo *gcpartifactregistryrepov1.GcpArtifactRegistryRepo
	GcpLabels               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcpartifactregistryrepov1.GcpArtifactRegistryRepoStackInput) *Locals {
	locals := &Locals{}

	locals.GcpArtifactRegistryRepo = stackInput.Target

	target := stackInput.Target

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpArtifactRegistryRepo.String()),
	}

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	return locals
}
