package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestorageobjectreplicationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageobjectreplication/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestorageobjectreplicationv1alpha1.AzureStorageObjectReplicationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageObjectReplication.Spec

	// The source account name, parsed from the resolved ARM id, keys the
	// Pulumi resource name (the account pair is what a policy IS). The
	// id must END with /storageAccounts/{name} (matching the Terraform
	// module's anchored regex), so a malformed id fails loudly here.
	sourceIdParts := strings.Split(locals.SourceStorageAccountId, "/storageAccounts/")
	if len(sourceIdParts) != 2 || sourceIdParts[1] == "" || strings.Contains(sourceIdParts[1], "/") {
		return errors.Errorf("source_storage_account_id %q is not a storage-account ARM id", locals.SourceStorageAccountId)
	}
	destinationIdParts := strings.Split(locals.DestinationStorageAccountId, "/storageAccounts/")
	if len(destinationIdParts) != 2 || destinationIdParts[1] == "" || strings.Contains(destinationIdParts[1], "/") {
		return errors.Errorf("destination_storage_account_id %q is not a storage-account ARM id", locals.DestinationStorageAccountId)
	}

	// Container-to-container mappings. The provider handles the
	// two-sided materialization (destination first -- which assigns rule
	// IDs -- then the source mirror), so one resource IS the pair.
	rules := make(storage.ObjectReplicationRuleArray, 0, len(spec.Rules))
	for _, rule := range spec.Rules {
		ruleArgs := storage.ObjectReplicationRuleArgs{
			SourceContainerName:      pulumi.String(rule.SourceContainerName.GetValue()),
			DestinationContainerName: pulumi.String(rule.DestinationContainerName.GetValue()),
		}
		// Unset lets the provider default (OnlyNewObjects -- no backfill)
		// apply; Everything backfills the whole container; an RFC 3339
		// instant backfills blobs created after that moment. Presence-
		// guarded because IaC inputs do not materialize proto defaults.
		if rule.GetCopyBlobsCreatedAfter() != "" {
			ruleArgs.CopyBlobsCreatedAfter = pulumi.String(rule.GetCopyBlobsCreatedAfter())
		}
		// The spec names this prefix_match after ARM's own INCLUDE
		// semantics (prefixMatch); the provider input's "filterOut" name
		// is historical and means the same include-only-these-prefixes
		// behavior.
		if len(rule.PrefixMatch) > 0 {
			ruleArgs.FilterOutBlobsWithPrefixes = pulumi.ToStringArray(rule.PrefixMatch)
		}
		rules = append(rules, ruleArgs)
	}

	// Apply-time prerequisites the ACCOUNTS must carry (deliberately not
	// mirrored as spec validation -- the accounts arrive as references):
	// blob versioning + change feed on the source, blob versioning on
	// the destination. The policy carries no Azure tags (ARM does not
	// support tags on objectReplicationPolicies).
	createdPolicy, err := storage.NewObjectReplication(ctx,
		fmt.Sprintf("%s-to-%s", sourceIdParts[1], destinationIdParts[1]),
		&storage.ObjectReplicationArgs{
			SourceStorageAccountId:      pulumi.String(locals.SourceStorageAccountId),
			DestinationStorageAccountId: pulumi.String(locals.DestinationStorageAccountId),
			Rules:                       rules,
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create object replication policy from %s to %s",
			sourceIdParts[1], destinationIdParts[1])
	}

	// Export outputs. Azure materializes the one logical policy on
	// BOTH accounts under one server-assigned GUID -- the GUID (parsed
	// from the destination-side id, the authoritative copy, identically
	// to the Terraform module) is what `az storage account or-policy`
	// and the monitoring surfaces key on.
	ctx.Export(OpSourceObjectReplicationId, createdPolicy.SourceObjectReplicationId)
	ctx.Export(OpDestinationObjectReplicationId, createdPolicy.DestinationObjectReplicationId)
	ctx.Export(OpPolicyId, createdPolicy.DestinationObjectReplicationId.ApplyT(func(id string) string {
		segments := strings.Split(id, "/")
		return segments[len(segments)-1]
	}).(pulumi.StringOutput))

	return nil
}
