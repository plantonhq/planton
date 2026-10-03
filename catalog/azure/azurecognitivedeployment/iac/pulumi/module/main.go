package module

import (
	"github.com/pkg/errors"
	azurecognitivedeploymentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecognitivedeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cognitive"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecognitivedeploymentv1alpha1.AzureCognitiveDeploymentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCognitiveDeployment.Spec

	// The model being deployed. Format and name are ForceNew -- a
	// different model is a new deployment; version updates in place
	// (omitted to track the model's current default version).
	modelArgs := &cognitive.DeploymentModelArgs{
		Format: pulumi.String(spec.Model.Format),
		Name:   pulumi.String(spec.Model.Name),
	}
	if spec.Model.Version != "" {
		modelArgs.Version = pulumi.String(spec.Model.Version)
	}

	// The throughput class and capacity. The pay-per-token SKUs carry
	// no idle cost (capacity is a rate limit); the ProvisionedManaged
	// SKUs bill their PTU capacity continuously.
	skuArgs := &cognitive.DeploymentSkuArgs{
		// Already a wire value in the spec.
		Name: pulumi.String(spec.Sku.Name),
	}
	// Enum name -> wire value; unspecified omits the property so ARM
	// derives the tier from the SKU name.
	if tier, ok := skuTierWire[spec.Sku.Tier]; ok {
		skuArgs.Tier = pulumi.String(tier)
	}
	if spec.Sku.Size != "" {
		skuArgs.Size = pulumi.String(spec.Sku.Size)
	}
	if spec.Sku.Family != "" {
		skuArgs.Family = pulumi.String(spec.Sku.Family)
	}
	// Unset applies the provider default of 1. The in-place scale knob.
	if spec.Sku.Capacity != nil {
		skuArgs.Capacity = pulumi.Int(int(*spec.Sku.Capacity))
	}

	// Create the model deployment on its Azure AI services account.
	// Which models exist at which capacities differs per region and per
	// subscription quota -- ARM rejects what the region or quota cannot
	// host, so quota errors surface here, never as silent degradation.
	deploymentArgs := &cognitive.DeploymentArgs{
		Name: pulumi.String(spec.Name),
		// The parent account (kind "OpenAI" or "AIServices"). ForceNew.
		CognitiveAccountId:       pulumi.String(locals.CognitiveAccountId),
		Model:                    modelArgs,
		Sku:                      skuArgs,
		DynamicThrottlingEnabled: pulumi.Bool(spec.DynamicThrottlingEnabled),
	}

	// Optional+Computed on the provider: omit when the spec leaves it
	// empty so ARM assigns its default policy and reads don't drift.
	if spec.RaiPolicyName != "" {
		deploymentArgs.RaiPolicyName = pulumi.String(spec.RaiPolicyName)
	}

	// Enum name -> wire value; unspecified lets the provider apply its
	// default, "OnceNewDefaultVersionAvailable".
	if option, ok := versionUpgradeOptionWire[spec.VersionUpgradeOption]; ok {
		deploymentArgs.VersionUpgradeOption = pulumi.String(option)
	}

	createdDeployment, err := cognitive.NewDeployment(ctx,
		spec.Name,
		deploymentArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cognitive deployment %s", spec.Name)
	}

	ctx.Export(OpDeploymentId, createdDeployment.ID())
	ctx.Export(OpDeploymentName, createdDeployment.Name)
	// Optional+Computed: ARM resolves the version when the spec left it
	// unset -- export the resolved value.
	ctx.Export(OpModelVersion, createdDeployment.Model.Version())

	return nil
}
