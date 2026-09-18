package module

import (
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// project provisions the DigitalOcean project and exports its outputs.
func project(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Project, error) {
	spec := locals.DigitalOceanProject.Spec

	projectArgs := &digitalocean.ProjectArgs{
		Name: pulumi.String(spec.ProjectName),

		// Sent unconditionally, matching the Terraform module: false is
		// the provider default.
		IsDefault: pulumi.Bool(spec.IsDefault),
	}

	if spec.Description != "" {
		projectArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Unset defers to the provider's default purpose ("Web Application").
	// DigitalOcean stores non-standard purposes as "Other: <text>" and
	// strips the prefix on read, so free text round-trips cleanly; values
	// starting with "Other:" are unrepresentable (spec validation).
	if spec.Purpose != "" {
		projectArgs.Purpose = pulumi.StringPtr(spec.Purpose)
	}

	// Lowercase canonical (spec validation); DigitalOcean accepts it
	// case-insensitively and reports it back capitalized, which the
	// provider diff-suppresses.
	if spec.Environment != "" {
		projectArgs.Environment = pulumi.StringPtr(spec.Environment)
	}

	// Membership is managed only when declared: an empty list stays unset
	// so out-of-band assignments (and the resources' own project
	// selections) are left untouched -- the attribute is Optional+Computed
	// upstream, so omitting it adopts whatever the API reports without
	// drift.
	if len(spec.Resources) > 0 {
		// References are resolved to literal URNs before the module runs.
		var urns pulumi.StringArray
		for _, ref := range spec.Resources {
			urns = append(urns, pulumi.String(ref.GetValue()))
		}
		projectArgs.Resources = urns
	}

	createdProject, err := digitalocean.NewProject(
		ctx,
		"project",
		projectArgs,
		pulumi.Provider(digitalOceanProvider),
		// The relocation of members is asynchronous on DigitalOcean's side
		// and the provider retries the DELETE through "412 cannot delete a
		// project with resources" only until this timeout. Its 3-minute
		// default was exceeded live: a one-member project stayed non-empty
		// for over 180 seconds after the relocation was accepted, the
		// destroy failed, and a second destroy of the by-then-empty project
		// succeeded. Ten minutes covers the measured lag with room; a retry
		// that ends earlier costs nothing. Twin of the Terraform module's
		// timeouts block.
		pulumi.Timeouts(&pulumi.CustomTimeouts{Delete: "10m"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean project")
	}

	ctx.Export(OpProjectId, createdProject.ID())
	ctx.Export(OpOwnerUuid, createdProject.OwnerUuid)
	// The SDK surfaces the owner id as an integer; the outputs contract is
	// a string on both provisioners.
	ctx.Export(OpOwnerId, createdProject.OwnerId.ApplyT(func(id int) string {
		return strconv.Itoa(id)
	}).(pulumi.StringOutput))
	// Membership as DigitalOcean reports it after apply (the provider reads
	// the set back whether or not the spec manages it); sorted so both
	// provisioners export identical lists from the API's unordered set.
	ctx.Export(OpResourceUrns, createdProject.Resources.ApplyT(func(urns []string) []string {
		sorted := append([]string(nil), urns...)
		sort.Strings(sorted)
		return sorted
	}).(pulumi.StringArrayOutput))

	return createdProject, nil
}
