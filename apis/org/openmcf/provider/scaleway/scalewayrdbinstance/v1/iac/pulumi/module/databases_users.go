package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scaleway "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/databases"
)

// databasesAndUsers creates the bundled sub-resources that depend on the
// RDB instance: databases, users, privileges, and ACL rules.
//
// Creation order ensures dependency correctness:
//  1. Databases (depend on instance only)
//  2. Users (depend on instance only)
//  3. Privileges (depend on instance + specific user + specific database)
//  4. ACL rules (depend on instance only)
//
// Pulumi resolves dependencies automatically through resource references,
// so the ordering here is logical rather than strictly required.
func databasesAndUsers(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
	instance *databases.Instance,
) error {
	spec := locals.ScalewayRdbInstance.Spec

	// ── 1. Create databases ─────────────────────────────────────────────

	createdDatabases := make(map[string]*databases.Database)

	for _, db := range spec.Databases {
		resourceName := fmt.Sprintf("db-%s", db.Name)

		createdDb, err := databases.NewDatabase(
			ctx,
			resourceName,
			&databases.DatabaseArgs{
				InstanceId: instance.ID(),
				Name:       pulumi.String(db.Name),
				Region:     pulumi.StringPtr(spec.Region),
			},
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{instance}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create database %q", db.Name)
		}

		createdDatabases[db.Name] = createdDb
	}

	// ── 2. Create users ─────────────────────────────────────────────────

	createdUsers := make(map[string]*databases.User)

	for _, user := range spec.Users {
		resourceName := fmt.Sprintf("user-%s", user.Name)

		createdUser, err := databases.NewUser(
			ctx,
			resourceName,
			&databases.UserArgs{
				InstanceId: instance.ID(),
				Name:       pulumi.String(user.Name),
				Password:   pulumi.StringPtr(user.Password),
				IsAdmin:    pulumi.BoolPtr(user.IsAdmin),
				Region:     pulumi.StringPtr(spec.Region),
			},
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{instance}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create user %q", user.Name)
		}

		createdUsers[user.Name] = createdUser
	}

	// ── 3. Create privileges (user × database grants) ───────────────────

	for _, user := range spec.Users {
		for _, priv := range user.Privileges {
			resourceName := fmt.Sprintf("priv-%s-%s", user.Name, priv.DatabaseName)

			// Build dependency list: always depend on the user. If the
			// database was created by this module, also depend on it.
			deps := []pulumi.Resource{createdUsers[user.Name]}
			if db, ok := createdDatabases[priv.DatabaseName]; ok {
				deps = append(deps, db)
			}

			_, err := databases.NewPrivilege(
				ctx,
				resourceName,
				&databases.PrivilegeArgs{
					InstanceId:   instance.ID(),
					UserName:     pulumi.String(user.Name),
					DatabaseName: pulumi.String(priv.DatabaseName),
					Permission:   pulumi.String(priv.Permission),
					Region:       pulumi.StringPtr(spec.Region),
				},
				pulumi.Provider(scalewayProvider),
				pulumi.DependsOn(deps),
			)
			if err != nil {
				return errors.Wrapf(err, "failed to create privilege for user %q on database %q", user.Name, priv.DatabaseName)
			}
		}
	}

	// ── 4. Create ACL rules (if specified) ──────────────────────────────

	if len(spec.AclRules) > 0 {
		aclRules := make(databases.AclAclRuleArray, 0, len(spec.AclRules))
		for _, rule := range spec.AclRules {
			aclRules = append(aclRules, &databases.AclAclRuleArgs{
				Ip:          pulumi.String(rule.Ip),
				Description: pulumi.StringPtr(rule.Description),
			})
		}

		_, err := databases.NewAcl(
			ctx,
			"acl",
			&databases.AclArgs{
				InstanceId: instance.ID(),
				AclRules:   aclRules,
				Region:     pulumi.StringPtr(spec.Region),
			},
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{instance}),
		)
		if err != nil {
			return errors.Wrap(err, "failed to create acl rules")
		}
	}

	return nil
}
