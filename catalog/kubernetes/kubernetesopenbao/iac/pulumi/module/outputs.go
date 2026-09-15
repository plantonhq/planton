package module

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesOpenBaoStackOutputs field.
const (
	OpNamespace          = "namespace"
	OpService            = "service"
	OpInternalService    = "internal_service"
	OpActiveService      = "active_service"
	OpUiService          = "ui_service"
	OpApiEndpoint        = "api_endpoint"
	OpPort               = "port"
	OpServiceAccountName = "service_account_name"
	OpPortForwardCommand = "port_forward_command"

	OpBackupServiceAccountName = "backup_service_account_name"
	OpBackupPolicyName         = "backup_policy_name"
	OpBackupAuthRole           = "backup_auth_role"
	OpBackupCronJobName        = "backup_cron_job_name"
	OpRestoreJobName           = "restore_job_name"
)

// exportOutputs publishes the composition handles. All names derive from
// the fullnameOverride pin (= metadata.name). Root tokens and unseal
// keys are deliberately NOT outputs — `bao operator init` produces them
// at runtime and no deployment surface ever holds them.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.ReleaseName))
	ctx.Export(OpInternalService, pulumi.String(locals.ReleaseName+"-internal"))

	// The active-leader Service exists only in HA mode (selected on the
	// openbao-active label the server itself maintains through
	// service_registration).
	activeService := ""
	if locals.Mode == modeHa {
		activeService = locals.ReleaseName + "-active"
	}
	ctx.Export(OpActiveService, pulumi.String(activeService))

	uiService := ""
	uiEnabled := true
	if locals.Spec.UiEnabled != nil {
		uiEnabled = locals.Spec.GetUiEnabled()
	}
	if uiEnabled {
		uiService = locals.ReleaseName + "-ui"
	}
	ctx.Export(OpUiService, pulumi.String(uiService))

	ctx.Export(OpApiEndpoint, pulumi.String(locals.apiEndpoint()))
	ctx.Export(OpPort, pulumi.String(fmt.Sprintf("%d", vars.ApiPort)))
	// serviceAccount.name is unset in the spec surface, so the chart
	// falls back to the fullname — which the module pins to
	// metadata.name.
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpPortForwardCommand, pulumi.String(fmt.Sprintf(
		"kubectl port-forward -n %s svc/%s %d:%d",
		locals.Namespace, locals.ReleaseName, vars.ApiPort, vars.ApiPort)))

	// The backup handles: one name (`<name>-backup`) serves as the job's
	// ServiceAccount, its OpenBao policy, and the CronJob, so the login
	// recipe an operator copies from these outputs is one noun. Empty
	// strings when the block is not declared.
	backupName, backupRole := "", ""
	if locals.BackupEnabled {
		backupName, backupRole = locals.BackupName, locals.BackupAuthRole
	}
	ctx.Export(OpBackupServiceAccountName, pulumi.String(backupName))
	ctx.Export(OpBackupPolicyName, pulumi.String(backupName))
	ctx.Export(OpBackupAuthRole, pulumi.String(backupRole))
	ctx.Export(OpBackupCronJobName, pulumi.String(backupName))
	ctx.Export(OpRestoreJobName, pulumi.String(locals.RestoreJobName))
}
