package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/tags"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tagBinding attaches the tag value to the resource.
//
// Every input is immutable, so any change replaces the binding. Two
// provider resources sit behind the kind: the global binding for
// organizations, folders, projects, and other global resources, and the
// location-scoped binding when spec.location names the region or zone of
// a regional or zonal resource -- the spec's location is the selector.
//
// Google requires a project's NUMBER in the parent. A GcpProject reference
// resolves to the number and a numeric literal is used as is (both
// rendered in locals); a project ID literal, or an empty parent meaning
// the provider's default project, is resolved through one read of the
// project at apply time -- the same guarded lookup the Terraform module
// count-gates on data.google_project, so neither engine performs a live
// read when the number is already known.
func tagBinding(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpTagBinding.Spec

	parent := locals.Parent
	if parent == "" {
		args := &organizations.LookupProjectArgs{}
		if locals.ProjectIdToResolve != "" {
			args.ProjectId = pulumi.StringRef(locals.ProjectIdToResolve)
		}
		project, err := organizations.LookupProject(ctx, args, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the project number for the tag binding's parent")
		}
		if project.Number == "" {
			return errors.New("the binding names no parent and the provider has no default project -- set spec.parent or configure a project")
		}
		parent = resourceManagerPrefix + "projects/" + project.Number
	}

	var deletionPolicy pulumi.StringPtrInput
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	var name, boundParent, tagValue pulumi.StringOutput
	if locals.IsLocationScoped {
		created, err := tags.NewLocationTagBinding(ctx, locals.GcpTagBinding.Metadata.Name, &tags.LocationTagBindingArgs{
			Parent:         pulumi.String(parent),
			TagValue:       pulumi.String(spec.TagValue.GetValue()),
			Location:       pulumi.StringPtr(spec.Location),
			DeletionPolicy: deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create location-scoped tag binding")
		}
		name, boundParent, tagValue = created.Name, created.Parent, created.TagValue
	} else {
		created, err := tags.NewTagBinding(ctx, locals.GcpTagBinding.Metadata.Name, &tags.TagBindingArgs{
			Parent:         pulumi.String(parent),
			TagValue:       pulumi.String(spec.TagValue.GetValue()),
			DeletionPolicy: deletionPolicy,
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create tag binding")
		}
		name, boundParent, tagValue = created.Name, created.Parent, created.TagValue
	}

	ctx.Export(OpName, name)
	ctx.Export(OpParent, boundParent)
	ctx.Export(OpTagValue, tagValue)

	return nil
}
