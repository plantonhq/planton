package module

import (
	"strconv"
	"strings"

	gcpcertificatemapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertificatemap/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpCertificateMap *gcpcertificatemapv1alpha1.GcpCertificateMap

	// The map name defaults to metadata.name when the spec leaves
	// map_name empty — the same naming basis every kind uses.
	MapName string

	// Merged labels: spec labels first so platform attribution labels win
	// on key conflicts — identical merge order to the Terraform module.
	// Applied to the map and every entry (per-entry spec labels merged
	// underneath per entry).
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcertificatemapv1alpha1.GcpCertificateMapIacInput) *Locals {
	target := iacInput.Target

	mapName := target.Spec.MapName
	if mapName == "" {
		mapName = target.Metadata.Name
	}

	gcpLabels := map[string]string{}
	for key, value := range target.Spec.Labels {
		gcpLabels[key] = value
	}
	gcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	gcpLabels[gcplabelkeys.ResourceName] = target.Metadata.Name
	gcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpCertificateMap.String())
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
		GcpCertificateMap: target,
		MapName:           mapName,
		GcpLabels:         gcpLabels,
	}
}

// entryLabels layers the shared map-level set (user labels + platform
// attribution keys) over an entry's own labels, so platform attribution
// can never be shadowed — identical merge order to the Terraform module.
func entryLabels(shared map[string]string, own map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range own {
		merged[key] = value
	}
	for key, value := range shared {
		merged[key] = value
	}
	return merged
}
