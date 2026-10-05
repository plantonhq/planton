package module

import (
	"strconv"
	"strings"

	gcpsecretmanagersecretv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsecretmanagersecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpSecretManagerSecret *gcpsecretmanagersecretv1alpha1.GcpSecretManagerSecret

	// The cloud-side secret ID defaults to metadata.name when the spec
	// leaves secret_id empty — the same naming basis every kind uses.
	SecretId string

	// Merged labels: spec labels first so platform attribution labels win
	// on key conflicts — identical merge order to the Terraform module.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpsecretmanagersecretv1alpha1.GcpSecretManagerSecretIacInput) *Locals {
	target := iacInput.Target

	secretId := target.Spec.SecretId
	if secretId == "" {
		secretId = target.Metadata.Name
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpSecretManagerSecret.String())

	if target.Metadata.Org != "" {
		gcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		gcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		gcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return &Locals{
		GcpSecretManagerSecret: target,
		SecretId:               secretId,
		GcpLabels:              gcpLabels,
	}
}
