package module

import (
	azuremonitordatacollectionruleassociationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitordatacollectionruleassociation/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMonitorDataCollectionRuleAssociation *azuremonitordatacollectionruleassociationv1alpha1.AzureMonitorDataCollectionRuleAssociation

	// The three reference fields are StringValueOrRef; the platform
	// middleware resolves valueFrom references before IaC modules run,
	// so GetValue() always returns the resolved literals.
	TargetResourceId         string
	DataCollectionRuleId     string
	DataCollectionEndpointId string
}

// The association carries NO tags argument on the provider (ARM
// extension resources are untagged), so this module derives no tag map.

func initializeLocals(ctx *pulumi.Context, iacInput *azuremonitordatacollectionruleassociationv1alpha1.AzureMonitorDataCollectionRuleAssociationIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMonitorDataCollectionRuleAssociation = iacInput.Target
	target := iacInput.Target

	locals.TargetResourceId = target.Spec.TargetResourceId.GetValue()
	locals.DataCollectionRuleId = target.Spec.DataCollectionRuleId.GetValue()
	locals.DataCollectionEndpointId = target.Spec.DataCollectionEndpointId.GetValue()

	return locals
}
