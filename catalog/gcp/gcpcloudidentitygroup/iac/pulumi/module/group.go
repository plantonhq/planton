package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudidentity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// group builds the Google Group and one membership per member. The group
// key, parent, namespace, initial configuration, and labels are immutable;
// the display name, description, and each membership's roles change in
// place. A different member is a different membership.
func group(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudIdentityGroup.Spec

	groupKey := &cloudidentity.GroupGroupKeyArgs{Id: pulumi.String(spec.GroupEmail)}
	if spec.GroupNamespace != "" {
		groupKey.Namespace = pulumi.StringPtr(spec.GroupNamespace)
	}

	args := &cloudidentity.GroupArgs{
		Parent:      pulumi.String(spec.CustomerId),
		DisplayName: pulumi.StringPtr(locals.DisplayName),
		GroupKey:    groupKey,
		Labels:      pulumi.ToStringMap(locals.GroupLabels),
		// The proto default (EMPTY) is applied by the manifest loader before
		// the module runs; stated explicitly so a hand-written input agrees.
		InitialGroupConfig: pulumi.StringPtr(orDefault(spec.GetInitialGroupConfig(), "EMPTY")),
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := cloudidentity.NewGroup(ctx, "group", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create group")
	}

	for _, m := range spec.Memberships {
		email := m.Member.GetValue()
		memberKey := &cloudidentity.GroupMembershipPreferredMemberKeyArgs{Id: pulumi.String(email)}
		if m.MemberNamespace != "" {
			memberKey.Namespace = pulumi.StringPtr(m.MemberNamespace)
		}

		// Unset roles mean MEMBER only; an expiry rides its role.
		var roles cloudidentity.GroupMembershipRoleArray
		if len(m.Roles) == 0 {
			roles = append(roles, &cloudidentity.GroupMembershipRoleArgs{Name: pulumi.String("MEMBER")})
		}
		for _, r := range m.Roles {
			role := &cloudidentity.GroupMembershipRoleArgs{Name: pulumi.String(r.Name)}
			if r.ExpireTime != "" {
				role.ExpiryDetail = &cloudidentity.GroupMembershipRoleExpiryDetailArgs{ExpireTime: pulumi.String(r.ExpireTime)}
			}
			roles = append(roles, role)
		}

		mArgs := &cloudidentity.GroupMembershipArgs{
			Group:              created.ID(),
			PreferredMemberKey: memberKey,
			Roles:              roles,
		}
		// Adopt an existing membership instead of failing; sent only when
		// true (the provider default is false).
		if m.CreateIgnoreAlreadyExists {
			mArgs.CreateIgnoreAlreadyExists = pulumi.BoolPtr(true)
		}
		// The memberships share the group's fate on destroy.
		if spec.DeletionPolicy != "" {
			mArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if _, err := cloudidentity.NewGroupMembership(ctx, "membership-"+email, mArgs,
			pulumi.Provider(gcpProvider), pulumi.Parent(created)); err != nil {
			return errors.Wrapf(err, "failed to create membership for %s", email)
		}
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpGroupEmail, pulumi.String(spec.GroupEmail))
	ctx.Export(OpMembershipCount, pulumi.String(strconv.Itoa(len(spec.Memberships))))
	return nil
}

// orDefault returns v, or def when v is empty -- the module-side twin of
// the proto default the manifest loader applies.
func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
