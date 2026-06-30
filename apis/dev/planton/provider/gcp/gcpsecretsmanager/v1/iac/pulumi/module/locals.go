package module

import (
	"strconv"
	"strings"

	gcpsecretsmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpsecretsmanager/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpSecretsManager *gcpsecretsmanagerv1.GcpSecretsManager
	GcpLabels         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcpsecretsmanagerv1.GcpSecretsManagerStackInput) *Locals {
	locals := &Locals{}

	locals.GcpSecretsManager = stackInput.Target

	target := stackInput.Target

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpSecretsManager.String()),
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
