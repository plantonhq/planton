package module

import (
	"strings"

	azuremssqlelasticpoolv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlelasticpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMssqlElasticPool *azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPool
	// ResourceGroupName and ServerName are derived from the parent
	// server's ARM id -- the spec carries ONE authoritative parent
	// reference:
	// /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Sql/servers/{name}
	ResourceGroupName string
	ServerName        string
	AzureTags         map[string]string
}

// skuTierStrings derives the service tier ARM wants alongside the SKU
// name -- a pure function of it, making a name/tier mismatch
// unrepresentable. Exhaustive over the spec's closed sku_name vocabulary.
var skuTierStrings = map[string]string{
	"BasicPool":    "Basic",
	"StandardPool": "Standard",
	"PremiumPool":  "Premium",
	"GP_Gen5":      "GeneralPurpose",
	"GP_Fsv2":      "GeneralPurpose",
	"GP_DC":        "GeneralPurpose",
	"BC_Gen5":      "BusinessCritical",
	"BC_DC":        "BusinessCritical",
	"HS_Gen5":      "Hyperscale",
	"HS_PRMS":      "Hyperscale",
	"HS_MOPRMS":    "Hyperscale",
}

// skuFamilyStrings derives the hardware family for vCore pools (the
// name's suffix); DTU pools carry no family.
var skuFamilyStrings = map[string]string{
	"GP_Gen5":   "Gen5",
	"GP_Fsv2":   "Fsv2",
	"GP_DC":     "DC",
	"BC_Gen5":   "Gen5",
	"BC_DC":     "DC",
	"HS_Gen5":   "Gen5",
	"HS_PRMS":   "PRMS",
	"HS_MOPRMS": "MOPRMS",
}

// enclaveTypeStrings maps the confidential-computing enclave enum to ARM's
// values.
var enclaveTypeStrings = map[azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolEnclaveType]string{
	azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolEnclaveType_VBS:             "VBS",
	azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolEnclaveType_DEFAULT_ENCLAVE: "Default",
}

// licenseTypeStrings maps the Azure Hybrid Benefit enum to ARM's values.
var licenseTypeStrings = map[azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolLicenseType]string{
	azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolLicenseType_BASE_PRICE:       "BasePrice",
	azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolLicenseType_LICENSE_INCLUDED: "LicenseIncluded",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMssqlElasticPool = iacInput.Target
	target := iacInput.Target

	serverIdParts := strings.Split(target.Spec.ServerId.GetValue(), "/")
	if len(serverIdParts) >= 9 {
		locals.ResourceGroupName = serverIdParts[4]
		locals.ServerName = serverIdParts[8]
	}

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMssqlElasticPool.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	// The user's spec tags merge over the metadata-derived tags -- user
	// tags deliberately win so an org's governance conventions can
	// override the derived values where they collide.
	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
