package module

import (
	"github.com/pkg/errors"
	azuremssqlelasticpoolv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlelasticpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMssqlElasticPool.Spec

	// The server is addressed by name + resource group (azurerm's
	// contract), both derived from the spec's server ARM id; the region
	// must match the server's (ARM rejects a mismatch).
	poolArgs := &mssql.ElasticPoolArgs{
		Name:              pulumi.String(spec.PoolName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServerName:        pulumi.String(locals.ServerName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The tier and family are derived from the sku name (pure functions
	// of it), so a mismatched combination is unrepresentable. DTU pools
	// (BasicPool/StandardPool/PremiumPool) carry no family.
	skuArgs := &mssql.ElasticPoolSkuArgs{
		Name:     pulumi.String(spec.SkuName),
		Tier:     pulumi.String(skuTierStrings[spec.SkuName]),
		Capacity: pulumi.Int(int(spec.Capacity)),
	}
	if family, ok := skuFamilyStrings[spec.SkuName]; ok {
		skuArgs.Family = pulumi.String(family)
	}
	poolArgs.Sku = skuArgs

	// What any ONE member database may consume: min is guaranteed
	// (reserved even while idle), max caps noisy neighbors.
	poolArgs.PerDatabaseSettings = &mssql.ElasticPoolPerDatabaseSettingsArgs{
		MinCapacity: pulumi.Float64(spec.PerDatabaseSettings.MinCapacity),
		MaxCapacity: pulumi.Float64(spec.PerDatabaseSettings.MaxCapacity),
	}

	// The pool's total storage cap -- gigabytes XOR bytes (spec-
	// validated); neither set lets ARM apply the SKU default.
	if spec.MaxSizeGb != nil {
		poolArgs.MaxSizeGb = pulumi.Float64(spec.GetMaxSizeGb())
	}
	if spec.MaxSizeBytes != nil {
		poolArgs.MaxSizeBytes = pulumi.Int(int(spec.GetMaxSizeBytes()))
	}

	poolArgs.ZoneRedundant = pulumi.Bool(spec.ZoneRedundant)

	// Every database in the pool must share the enclave type, so it
	// lives at the pool level. Changing it is disruptive -- plan
	// accordingly.
	if spec.EnclaveType != azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolEnclaveType_azure_mssql_elastic_pool_enclave_type_unspecified {
		poolArgs.EnclaveType = pulumi.String(enclaveTypeStrings[spec.EnclaveType])
	}

	if spec.LicenseType != azuremssqlelasticpoolv1alpha1.AzureMssqlElasticPoolLicenseType_azure_mssql_elastic_pool_license_type_unspecified {
		poolArgs.LicenseType = pulumi.String(licenseTypeStrings[spec.LicenseType])
	}

	// Hyperscale pools only: readable HA replicas per member database.
	if spec.HighAvailabilityReplicaCount != nil {
		poolArgs.HighAvailabilityReplicaCount = pulumi.Int(int(spec.GetHighAvailabilityReplicaCount()))
	}

	// Member databases inherit this window (and must not set their own).
	// Presence-guarded to the spec default -- IaC inputs built from a
	// manifest do NOT materialize proto defaults.
	if spec.MaintenanceConfigurationName != nil {
		poolArgs.MaintenanceConfigurationName = pulumi.String(spec.GetMaintenanceConfigurationName())
	} else {
		poolArgs.MaintenanceConfigurationName = pulumi.String("SQL_Default")
	}

	pool, err := mssql.NewElasticPool(ctx,
		spec.PoolName,
		poolArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mssql elastic pool %s", spec.PoolName)
	}

	// Export outputs from the created resource.
	ctx.Export(OpElasticPoolId, pool.ID())
	ctx.Export(OpElasticPoolName, pool.Name)

	return nil
}
