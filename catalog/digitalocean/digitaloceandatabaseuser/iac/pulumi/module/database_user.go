package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// databaseUser provisions the database user and exports its outputs.
func databaseUser(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseUser, error) {
	spec := locals.DigitalOceanDatabaseUser.Spec

	userArgs := &digitalocean.DatabaseUserArgs{
		// References are resolved to the literal cluster UUID before the
		// module runs.
		ClusterId: pulumi.String(spec.Cluster.GetValue()),
		Name:      pulumi.String(spec.UserName),
	}

	// MySQL clusters only (API-enforced). Unset defers to DigitalOcean's
	// caching_sha2_password default; updates apply through a
	// password-preserving auth reset.
	if spec.MysqlAuthPlugin != "" {
		userArgs.MysqlAuthPlugin = pulumi.StringPtr(spec.MysqlAuthPlugin)
	}

	// Engine-specific ACLs. The provider records `settings` only from the
	// CREATE response and never refreshes it from the API, so this
	// configuration is the source of truth. The SDK models settings as a
	// list; a user has one settings object, so the spec's single message
	// wraps into a one-element array (mirroring the Terraform module's
	// single dynamic block).
	//
	// The element is sent exactly when the spec sets `settings` -- an EMPTY
	// message counts -- because what DigitalOcean stores is engine-specific
	// (measured 2026-09-17): every PostgreSQL user comes back with a
	// settings object (`pg_allow_replication: false`), which the provider
	// keeps as one empty element, so a PostgreSQL manifest declares
	// `settings: {}` to match it; a MySQL user never carries one and the API
	// REFUSES a settings update (`422 operation is not supported for this
	// cluster type`), so a MySQL manifest leaves it out. The module cannot
	// know the engine from a cluster UUID, so the manifest carries that
	// knowledge. Same shape as the Terraform module.
	if spec.Settings != nil {
		settingArgs := digitalocean.DatabaseUserSettingArgs{}

		if len(spec.Settings.KafkaAcls) > 0 {
			var acls digitalocean.DatabaseUserSettingAclArray
			for _, acl := range spec.Settings.KafkaAcls {
				acls = append(acls, digitalocean.DatabaseUserSettingAclArgs{
					Topic:      pulumi.String(acl.Topic),
					Permission: pulumi.String(acl.Permission),
				})
			}
			settingArgs.Acls = acls
		}

		if len(spec.Settings.OpensearchAcls) > 0 {
			var acls digitalocean.DatabaseUserSettingOpensearchAclArray
			for _, acl := range spec.Settings.OpensearchAcls {
				acls = append(acls, digitalocean.DatabaseUserSettingOpensearchAclArgs{
					Index:      pulumi.String(acl.Index),
					Permission: pulumi.String(acl.Permission),
				})
			}
			settingArgs.OpensearchAcls = acls
		}

		userArgs.Settings = digitalocean.DatabaseUserSettingArray{settingArgs}
	}

	createdUser, err := digitalocean.NewDatabaseUser(
		ctx,
		"user",
		userArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean database user")
	}

	ctx.Export(OpClusterId, createdUser.ClusterId)
	ctx.Export(OpUserName, createdUser.Name)
	ctx.Export(OpRole, createdUser.Role)
	ctx.Export(OpPassword, createdUser.Password)
	ctx.Export(OpAccessCert, createdUser.AccessCert)
	ctx.Export(OpAccessKey, createdUser.AccessKey)

	return createdUser, nil
}
