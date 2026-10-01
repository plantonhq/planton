package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/gkehub"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// scope creates the team scope and, folded into it, its fleet namespaces,
// role bindings, and cluster bindings -- each keyed by its own declared
// ID, so adding or removing one never renames or recreates its siblings.
func scope(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGkeFleetScope.Spec
	resourceName := locals.GcpGkeFleetScope.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Fleet API (GKE Hub). DisableOnDestroy stays false: tearing down
	// one team's scope must never disable the API for the rest of the
	// fleet.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("gkehub.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdGkehubApi, err := projects.NewService(ctx, "gcpflsc-gkehub.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable gkehub.googleapis.com")
	}

	// The scope ID defaults to metadata.name -- identical to the Terraform
	// module.
	scopeId := spec.ScopeId
	if scopeId == "" {
		scopeId = resourceName
	}

	var projectArg pulumi.StringPtrInput
	if project != "" {
		projectArg = pulumi.StringPtr(project)
	}
	var deletionPolicy pulumi.StringPtrInput
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	scopeArgs := &gkehub.ScopeArgs{
		Project:        projectArg,
		ScopeId:        pulumi.String(scopeId),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		DeletionPolicy: deletionPolicy,
	}
	if len(spec.NamespaceLabels) > 0 {
		scopeArgs.NamespaceLabels = pulumi.ToStringMap(spec.NamespaceLabels)
	}
	createdScope, err := gkehub.NewScope(ctx, resourceName, scopeArgs,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdGkehubApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create fleet scope")
	}

	// The scope's fleet namespaces. Google takes the scope twice: its short
	// ID in the URL and its full name in the body.
	for _, namespace := range spec.Namespaces {
		namespaceArgs := &gkehub.NamespaceArgs{
			Project:          projectArg,
			ScopeId:          createdScope.ScopeId,
			Scope:            createdScope.Name,
			ScopeNamespaceId: pulumi.String(namespace.ScopeNamespaceId),
			Labels:           pulumi.ToStringMap(mergeLabels(namespace.Labels, locals.AttributionLabels)),
			DeletionPolicy:   deletionPolicy,
		}
		if len(namespace.NamespaceLabels) > 0 {
			namespaceArgs.NamespaceLabels = pulumi.ToStringMap(namespace.NamespaceLabels)
		}
		if _, err := gkehub.NewNamespace(ctx, fmt.Sprintf("%s-namespace-%s", resourceName, namespace.ScopeNamespaceId),
			namespaceArgs, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrapf(err, "failed to create fleet namespace %s", namespace.ScopeNamespaceId)
		}
	}

	// Who gets which access in the scope's namespaces: exactly one of user
	// or group, and exactly one of a predefined or a custom role.
	for _, binding := range spec.RbacRoleBindings {
		role := &gkehub.ScopeRbacRoleBindingRoleArgs{}
		if binding.Role.PredefinedRole != "" {
			role.PredefinedRole = pulumi.StringPtr(binding.Role.PredefinedRole)
		}
		if binding.Role.CustomRole != "" {
			role.CustomRole = pulumi.StringPtr(binding.Role.CustomRole)
		}
		bindingArgs := &gkehub.ScopeRbacRoleBindingArgs{
			Project:                projectArg,
			ScopeId:                createdScope.ScopeId,
			ScopeRbacRoleBindingId: pulumi.String(binding.ScopeRbacRoleBindingId),
			Role:                   role,
			Labels:                 pulumi.ToStringMap(mergeLabels(binding.Labels, locals.AttributionLabels)),
			DeletionPolicy:         deletionPolicy,
		}
		if user := binding.GetUser().GetValue(); user != "" {
			bindingArgs.User = pulumi.StringPtr(user)
		}
		if group := binding.GetGroup().GetValue(); group != "" {
			bindingArgs.Group = pulumi.StringPtr(group)
		}
		if _, err := gkehub.NewScopeRbacRoleBinding(ctx, fmt.Sprintf("%s-rbac-%s", resourceName, binding.ScopeRbacRoleBindingId),
			bindingArgs, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrapf(err, "failed to create scope role binding %s", binding.ScopeRbacRoleBindingId)
		}
	}

	// The clusters the team may use. A binding addresses its membership by
	// location and ID, parsed from the membership's full name; it lives in
	// the scope's fleet project, which Google requires the membership to
	// share.
	for _, binding := range spec.MembershipBindings {
		location, membershipId, err := parseMembershipName(binding.GetMembership().GetValue())
		if err != nil {
			return errors.Wrapf(err, "membership binding %s", binding.MembershipBindingId)
		}
		if _, err := gkehub.NewMembershipBinding(ctx, fmt.Sprintf("%s-binding-%s", resourceName, binding.MembershipBindingId),
			&gkehub.MembershipBindingArgs{
				Project:             projectArg,
				Location:            pulumi.String(location),
				MembershipId:        pulumi.String(membershipId),
				MembershipBindingId: pulumi.String(binding.MembershipBindingId),
				Scope:               createdScope.Name,
				Labels:              pulumi.ToStringMap(mergeLabels(binding.Labels, locals.AttributionLabels)),
				DeletionPolicy:      deletionPolicy,
			}, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrapf(err, "failed to create membership binding %s", binding.MembershipBindingId)
		}
	}

	ctx.Export(OpName, createdScope.Name)
	ctx.Export(OpScopeId, createdScope.ScopeId)
	ctx.Export(OpUid, createdScope.Uid)
	return nil
}

// parseMembershipName returns the location and ID of a membership's full
// name, "projects/{p}/locations/{l}/memberships/{id}" -- the Terraform
// module's split() twin.
func parseMembershipName(name string) (location, membershipId string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) != 6 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "memberships" {
		return "", "", errors.Errorf("membership %q is not projects/{project}/locations/{location}/memberships/{id}", name)
	}
	return parts[3], parts[5], nil
}
