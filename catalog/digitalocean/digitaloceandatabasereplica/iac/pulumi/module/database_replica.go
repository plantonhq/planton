package module

import (
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tagsCombinedBudget is DigitalOcean's cap on a database replica's COMBINED
// tags -- the tag names joined by commas -- measured 2026-09-17 against
// `POST /v2/databases/{id}/replicas`: 255 characters pass, 256 fail with
// `422 combined tags cannot exceed 255 characters` (the same rule and the
// same honest validation error as the primary's create; nothing is created
// on the 422). The six Planton label tags carry metadata.name and
// metadata.id, so a long resource name spends the budget before any
// spec.tags entry does -- and because replica tags are create-only, the
// guard is the only place a customer learns the rule before a replacement.
// Checked before anything renders, with the same number and the same
// message as the Terraform module's precondition (its twin).
const tagsCombinedBudget = 255

// databaseReplica provisions the read-only replica and exports its
// outputs.
//
// Update semantics mirror the provider: only size and storage_size_mib
// change in place (a resize); every other argument change REPLACES the
// replica -- including tags, which are create-only upstream. region and
// size are required by the spec (explicit values kill the upstream
// omitted-value drift class that would otherwise schedule replacements).
func databaseReplica(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseReplica, error) {
	spec := locals.DigitalOceanDatabaseReplica.Spec

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags -- the exact set the Terraform module applies. Labels are added
	// in key order so the rendered list is deterministic on every apply;
	// that matters doubly here because tags are CREATE-ONLY upstream: a
	// change to the final set replaces the replica.
	tagSet := map[string]bool{}
	var tags []string
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, t)
		}
	}
	labelKeys := make([]string, 0, len(locals.DigitalOceanLabels))
	for k := range locals.DigitalOceanLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	var labelTags []string
	for _, k := range labelKeys {
		t := k + ":" + locals.DigitalOceanLabels[k]
		labelTags = append(labelTags, t)
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, t)
		}
	}

	// Fail loud on DigitalOcean's combined-tags budget before anything
	// renders (see tagsCombinedBudget; twin of the Terraform precondition).
	if combined := strings.Join(tags, ","); len(combined) > tagsCombinedBudget {
		return nil, errors.Errorf(
			"DigitalOcean caps a database replica's combined tags (joined by commas) at %d characters; this replica's %d tags join to %d characters. Shorten metadata.name or metadata.id, or remove entries from spec.tags -- the Planton label tags alone use %d characters here.",
			tagsCombinedBudget, len(tags), len(combined), len(strings.Join(labelTags, ",")))
	}

	tagInputs := make(pulumi.StringArray, 0, len(tags))
	for _, t := range tags {
		tagInputs = append(tagInputs, pulumi.String(t))
	}

	replicaArgs := &digitalocean.DatabaseReplicaArgs{
		// References are resolved to the literal cluster UUID before the
		// module runs. Enum value names are exactly the DigitalOcean
		// region slugs.
		ClusterId: pulumi.String(spec.Cluster.GetValue()),
		Name:      pulumi.String(spec.ReplicaName),
		Region:    pulumi.String(spec.Region.String()),
		Size:      pulumi.String(spec.Size),
		Tags:      tagInputs,
	}

	// Optional VPC placement for the replica's region (create-only;
	// Optional+Computed upstream, so omission is drift-safe).
	if spec.Vpc != nil && spec.Vpc.GetValue() != "" {
		replicaArgs.PrivateNetworkUuid = pulumi.StringPtr(spec.Vpc.GetValue())
	}

	// The provider's storage_size_mib is a string holding a bare MiB
	// count; the spec carries the number. Must stay >= the primary's
	// storage.
	if spec.StorageSizeMib > 0 {
		replicaArgs.StorageSizeMib = pulumi.StringPtr(strconv.FormatUint(spec.StorageSizeMib, 10))
	}

	createdReplica, err := digitalocean.NewDatabaseReplica(
		ctx,
		"replica",
		replicaArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean database replica")
	}

	// replica_id is the API UUID (the SDK's Uuid attribute) -- the
	// resource's own ID() is a legacy composite string, not the UUID.
	ctx.Export(OpReplicaId, createdReplica.Uuid)
	ctx.Export(OpClusterId, createdReplica.ClusterId)
	ctx.Export(OpReplicaName, createdReplica.Name)
	ctx.Export(OpHost, createdReplica.Host)
	ctx.Export(OpPrivateHost, createdReplica.PrivateHost)
	ctx.Export(OpPort, createdReplica.Port)
	ctx.Export(OpDatabase, createdReplica.Database)
	ctx.Export(OpUser, createdReplica.User)
	ctx.Export(OpPassword, createdReplica.Password)
	ctx.Export(OpUri, createdReplica.Uri)
	ctx.Export(OpPrivateUri, createdReplica.PrivateUri)

	return createdReplica, nil
}
