package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/managedkafka"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// acl creates the access rules for one resource pattern on the cluster.
func acl(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpManagedKafkaAcl.Spec
	resourceName := locals.GcpManagedKafkaAcl.Metadata.Name

	// permission_type and host are sent only when set: the provider
	// defaults them to ALLOW and "*", the only host Google accepts -- the
	// Terraform module's rule.
	entries := managedkafka.AclAclEntryArray{}
	for _, entry := range spec.AclEntries {
		entryArgs := &managedkafka.AclAclEntryArgs{
			Principal: pulumi.String(entry.Principal),
			Operation: pulumi.String(entry.Operation),
		}
		if entry.PermissionType != "" {
			entryArgs.PermissionType = pulumi.String(entry.PermissionType)
		}
		if entry.Host != "" {
			entryArgs.Host = pulumi.String(entry.Host)
		}
		entries = append(entries, entryArgs)
	}

	args := &managedkafka.AclArgs{
		Location:   pulumi.String(spec.Location),
		Cluster:    pulumi.String(locals.ClusterId),
		AclId:      pulumi.String(spec.AclId),
		AclEntries: entries,
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := managedkafka.NewAcl(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create managed kafka acl")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpResourceType, created.ResourceType)
	ctx.Export(OpResourceName, created.ResourceName)
	ctx.Export(OpPatternType, created.PatternType)
	return nil
}
