package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpcloudsqluserv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudsqluser/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/sql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// user provisions a database user on a Cloud SQL instance. Users are
// first-class nodes: one per application/service with its own credential,
// instead of sharing the instance's admin user.
//
// No API enablement here: the instance this user lives on cannot exist
// without sqladmin.googleapis.com already enabled (its own module enables
// it).
//
// BUILT_IN users authenticate with the spec password (rotatable in place —
// updating the field updates the credential without recreating the user).
// CLOUD_IAM_* users authenticate through IAM and carry no password at all;
// on PostgreSQL the instance must first set the database flag
// "cloudsql.iam_authentication" = "on".
func user(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudSqlUser.Spec

	args := &sql.UserArgs{
		Name:     pulumi.String(databaseUserName(spec)),
		Instance: pulumi.String(spec.Instance.GetValue()),
		Type:     pulumi.StringPtr(spec.GetType()),
	}

	// An empty project falls back to the provider's default project — the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Marked secret so the credential is encrypted in Pulumi state; never
	// exported in outputs.
	if spec.Password != "" {
		args.Password = pulumi.ToSecret(pulumi.String(spec.Password)).(pulumi.StringOutput)
	}

	// MySQL-only user@host scoping; omitted on other engines.
	if spec.Host != "" {
		args.Host = pulumi.StringPtr(spec.Host)
	}

	// MySQL 8+ / PostgreSQL: roles granted at creation (custom roles must
	// already exist in the database).
	if len(spec.DatabaseRoles) > 0 {
		args.DatabaseRoles = pulumi.ToStringArray(spec.DatabaseRoles)
	}

	// DELETE (default) drops the user; ABANDON removes it from IaC
	// management — the documented workaround when owned objects block a
	// PostgreSQL drop; PREVENT fails destroying previews.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if pp := spec.PasswordPolicy; pp != nil {
		ppArgs := &sql.UserPasswordPolicyArgs{
			EnableFailedAttemptsCheck:  pulumi.BoolPtr(pp.EnableFailedAttemptsCheck),
			EnablePasswordVerification: pulumi.BoolPtr(pp.EnablePasswordVerification),
		}
		if pp.AllowedFailedAttempts != nil {
			ppArgs.AllowedFailedAttempts = pulumi.IntPtr(int(*pp.AllowedFailedAttempts))
		}
		if pp.PasswordExpirationDuration != "" {
			ppArgs.PasswordExpirationDuration = pulumi.StringPtr(pp.PasswordExpirationDuration)
		}
		args.PasswordPolicy = ppArgs
	}

	createdUser, err := sql.NewUser(ctx, "user", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	ctx.Export(OpUserName, createdUser.Name)
	ctx.Export(OpInstanceName, createdUser.Instance)

	return nil
}

// serviceAccountEmailSuffix is the part of a service account's email Cloud
// SQL does not store in an IAM database username.
const serviceAccountEmailSuffix = ".gserviceaccount.com"

// databaseUserName derives the username Cloud SQL expects. A service-account
// user (wired through service_account, or a literal user_name with the
// CLOUD_IAM_SERVICE_ACCOUNT type) is the account's email with the
// ".gserviceaccount.com" suffix dropped: PostgreSQL stores exactly that
// form, and MySQL keeps only the part before "@" -- one rule serves both
// engines (identical in the Terraform module). Every other type passes
// user_name through untouched.
func databaseUserName(spec *gcpcloudsqluserv1alpha1.GcpCloudSqlUserSpec) string {
	name := spec.UserName
	if spec.ServiceAccount.GetValue() != "" {
		name = spec.ServiceAccount.GetValue()
	}
	if spec.GetType() == "CLOUD_IAM_SERVICE_ACCOUNT" {
		return strings.TrimSuffix(name, serviceAccountEmailSuffix)
	}
	return name
}
