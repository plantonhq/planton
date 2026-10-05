package module

import (
	"strings"

	azureserviceplanv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureserviceplan/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServicePlan  *azureserviceplanv1alpha1.AzureServicePlan
	ResourceGroupName string
	AzureTags         map[string]string
	// OsType is Azure's wire value (Linux / Windows / WindowsContainer),
	// mapped from the spec enum; unset deploys Linux.
	OsType string
	// SkuName is Azure's wire spelling of the SKU (P1v3 style), mapped
	// from the spec enum (PREMIUM_P1V3 style).
	SkuName string
}

// osTypeStrings maps the spec's OS-type enum to Azure's wire values.
var osTypeStrings = map[azureserviceplanv1alpha1.AzureServicePlanOsType]string{
	azureserviceplanv1alpha1.AzureServicePlanOsType_LINUX:             "Linux",
	azureserviceplanv1alpha1.AzureServicePlanOsType_WINDOWS:           "Windows",
	azureserviceplanv1alpha1.AzureServicePlanOsType_WINDOWS_CONTAINER: "WindowsContainer",
}

// skuStrings maps the spec's SKU enum to Azure's wire spellings. Spelled
// out row by row so a vocabulary drift fails loudly at preview time
// instead of deploying a wrong SKU.
var skuStrings = map[azureserviceplanv1alpha1.AzureServicePlanSku]string{
	azureserviceplanv1alpha1.AzureServicePlanSku_FREE_F1:              "F1",
	azureserviceplanv1alpha1.AzureServicePlanSku_SHARED_D1:            "D1",
	azureserviceplanv1alpha1.AzureServicePlanSku_SHARED:               "SHARED",
	azureserviceplanv1alpha1.AzureServicePlanSku_BASIC_B1:             "B1",
	azureserviceplanv1alpha1.AzureServicePlanSku_BASIC_B2:             "B2",
	azureserviceplanv1alpha1.AzureServicePlanSku_BASIC_B3:             "B3",
	azureserviceplanv1alpha1.AzureServicePlanSku_STANDARD_S1:          "S1",
	azureserviceplanv1alpha1.AzureServicePlanSku_STANDARD_S2:          "S2",
	azureserviceplanv1alpha1.AzureServicePlanSku_STANDARD_S3:          "S3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P1V2:         "P1v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P2V2:         "P2v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P3V2:         "P3v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P0V3:         "P0v3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P1V3:         "P1v3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P2V3:         "P2v3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P3V3:         "P3v3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P1MV3:        "P1mv3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P2MV3:        "P2mv3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P3MV3:        "P3mv3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P4MV3:        "P4mv3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P5MV3:        "P5mv3",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P0V4:         "P0v4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P1V4:         "P1v4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P2V4:         "P2v4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P3V4:         "P3v4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P1MV4:        "P1mv4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P2MV4:        "P2mv4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P3MV4:        "P3mv4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P4MV4:        "P4mv4",
	azureserviceplanv1alpha1.AzureServicePlanSku_PREMIUM_P5MV4:        "P5mv4",
	azureserviceplanv1alpha1.AzureServicePlanSku_CONSUMPTION_Y1:       "Y1",
	azureserviceplanv1alpha1.AzureServicePlanSku_ELASTIC_PREMIUM_EP1:  "EP1",
	azureserviceplanv1alpha1.AzureServicePlanSku_ELASTIC_PREMIUM_EP2:  "EP2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ELASTIC_PREMIUM_EP3:  "EP3",
	azureserviceplanv1alpha1.AzureServicePlanSku_FLEX_CONSUMPTION_FC1: "FC1",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I1:          "I1",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I2:          "I2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I3:          "I3",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I1V2:        "I1v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I2V2:        "I2v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I3V2:        "I3v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I4V2:        "I4v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I5V2:        "I5v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I6V2:        "I6v2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I1MV2:       "I1mv2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I2MV2:       "I2mv2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I3MV2:       "I3mv2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I4MV2:       "I4mv2",
	azureserviceplanv1alpha1.AzureServicePlanSku_ISOLATED_I5MV2:       "I5mv2",
	azureserviceplanv1alpha1.AzureServicePlanSku_WORKFLOW_WS1:         "WS1",
	azureserviceplanv1alpha1.AzureServicePlanSku_WORKFLOW_WS2:         "WS2",
	azureserviceplanv1alpha1.AzureServicePlanSku_WORKFLOW_WS3:         "WS3",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureserviceplanv1alpha1.AzureServicePlanIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServicePlan = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Unset OS type deploys Linux -- the catalog's app kinds are
	// Linux-based.
	locals.OsType = "Linux"
	if target.Spec.OsType != azureserviceplanv1alpha1.AzureServicePlanOsType_azure_service_plan_os_type_unspecified {
		locals.OsType = osTypeStrings[target.Spec.OsType]
	}

	// The sku enum is spec-required (never unspecified), so the map hit
	// is guaranteed by validation.
	locals.SkuName = skuStrings[target.Spec.SkuName]

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureServicePlan.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
