package component

import (
	"context"
	"fmt"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/keycloak"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// Re-establishing the master admin on a restored realm.
//
// A platform whose database was restored from an archive (status.backup.
// restoredFrom, read by the database component from the live Cluster) comes
// back with the identity realm its source platform had -- every user, every
// client, and the master admin with the SOURCE's password. This install
// generated a new bootstrap admin Secret, and Keycloak reads
// KC_BOOTSTRAP_ADMIN_* into an empty master realm only, so the operator
// would sign in as admin forever refused and never reconcile the realm.
//
// Keycloak's own answer is its recovery command (`kc.sh bootstrap-admin
// user`), which creates a TEMPORARY admin -- and requires every server node
// to be stopped while it runs. So the recovery is two halves around the
// server's start: before the identity Deployment exists, a one-shot Job runs
// the command and creates the recovery admin; once the server answers, the
// operator signs in as the recovery admin, gives the real admin this
// install's password, deletes the recovery admin, and records completion in
// a ConfigMap so this runs exactly once per platform lifetime.

// identityRecoveryPending reports whether this platform's realm was restored
// and its master admin not yet re-established.
func (id *Identity) identityRecoveryPending(ctx context.Context, c client.Client, planton *v1.PlantonPlatform) (bool, error) {
	if planton.Status.Backup == nil || planton.Status.Backup.RestoredFrom == "" {
		return false, nil
	}
	var marker corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Name: resources.IdentityRecoveryMarkerName(planton.Name), Namespace: planton.Namespace}, &marker)
	if err == nil {
		return false, nil
	}
	if !apierrors.IsNotFound(err) {
		return false, fmt.Errorf("checking the identity recovery marker: %w", err)
	}
	return true, nil
}

// recoverAdminBeforeServer is the first half: with the identity server not
// running, create the recovery admin through Keycloak's recovery command. It
// returns a not-ready Result while the Job runs or when it failed, and nil
// once the Job succeeded -- the caller then starts the server. A server
// already running (an operator upgraded onto a restored platform whose admin
// it could never reach) is stopped first; the command refuses to run beside
// a live node.
func (id *Identity) recoverAdminBeforeServer(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, cfg resources.IdentityConfig, ownerRef *metav1.OwnerReference) (*Result, error) {
	log := logf.FromContext(ctx).WithValues("component", id.Name(), "step", "recover-admin")

	// The Job's state is read before anything is stopped. This half runs on
	// every reconcile until the second half writes the marker, and the pass
	// after the Job succeeds is exactly the one that starts the server: a
	// server stopped again here would be started and killed every pass, and
	// the second half -- which needs it answering -- would never run.
	name := resources.IdentityRecoveryAdminName(planton.Name)
	var job batchv1.Job
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: planton.Namespace}, &job)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("reading the identity recovery admin Job: %w", err)
	}
	if err == nil && job.Status.Succeeded > 0 {
		return nil, nil
	}

	var deploy appsv1.Deployment
	err = c.Get(ctx, types.NamespacedName{Name: resources.IdentityDeploymentName(planton.Name), Namespace: planton.Namespace}, &deploy)
	switch {
	case err == nil:
		log.Info("Stopping the identity server so the recovery command can run against the restored realm")
		if err := c.Delete(ctx, &deploy, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("stopping the identity server for admin recovery: %w", err)
		}
		return &Result{Ready: false, Reason: v1.ComponentReasonDeploying,
			Message: "Restarting the identity server to re-establish its admin credential on the restored realm"}, nil
	case !apierrors.IsNotFound(err):
		return nil, fmt.Errorf("checking the identity server before admin recovery: %w", err)
	}

	if _, err := id.EnsureAndReadCredential(ctx, c, name, planton.Namespace,
		map[string]string{resources.IdentityBootstrapAdminPasswordKey: ""}, ownerRef); err != nil {
		return nil, fmt.Errorf("ensuring the identity recovery admin Secret: %w", err)
	}

	err = c.Get(ctx, types.NamespacedName{Name: name, Namespace: planton.Namespace}, &job)
	if apierrors.IsNotFound(err) {
		log.Info("Creating the recovery admin on the restored realm", "job", name)
		if err := id.ApplyTypedObject(ctx, c, resources.IdentityRecoveryAdminJob(cfg)); err != nil {
			return nil, fmt.Errorf("applying the identity recovery admin Job: %w", err)
		}
		return &Result{Ready: false, Reason: v1.ComponentReasonDeploying,
			Message: fmt.Sprintf("Re-establishing the identity server's admin credential on the restored realm (job %s) before the server starts", name)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the identity recovery admin Job: %w", err)
	}

	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
			return &Result{Ready: false, Reason: v1.ComponentReasonJobFailed,
				Object: &v1.ComponentObjectReference{Kind: "Job", Name: job.Name},
				Message: fmt.Sprintf(
					"the identity admin recovery job %s failed (%s: %s) -- read its log with kubectl logs -n %s job/%s; "+
						"the identity server does not start until the restored realm's admin is re-established",
					job.Name, cond.Reason, oneLine(cond.Message), job.Namespace, job.Name),
			}, nil
		}
	}
	if job.Status.Succeeded == 0 {
		return &Result{Ready: false, Reason: v1.ComponentReasonDeploying,
			Object:  &v1.ComponentObjectReference{Kind: "Job", Name: job.Name},
			Message: fmt.Sprintf("Re-establishing the identity server's admin credential on the restored realm (job %s) before the server starts", name)}, nil
	}
	return nil, nil
}

// recoverAdminAfterServer is the second half: with the identity server
// answering, make the real admin the one this install knows. The real
// admin's own sign-in is the probe: if it works (a partial earlier run, or a
// realm that never needed recovery), only the record is written. It returns
// a not-ready Result when the recovery cannot proceed, and nil when the
// admin is re-established or the failure is the server's to report.
func (id *Identity) recoverAdminAfterServer(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, serverRoot, adminPassword string, httpClient *http.Client) (*Result, error) {
	log := logf.FromContext(ctx).WithValues("component", id.Name(), "step", "recover-admin")

	probe := keycloak.NewAdminClient(httpClient, serverRoot)
	err := probe.Authenticate(ctx, resources.IdentityBootstrapAdminUsername, adminPassword)
	if err == nil {
		log.Info("The master admin already authenticates on the restored realm; recording the recovery as complete")
		return nil, id.finishAdminRecovery(ctx, c, planton)
	}
	if !keycloak.IsCredentialRefused(err) {
		// The server is not answering the token endpoint at all; the realm
		// reconciler reports that in its own words on this pass.
		return nil, nil
	}

	name := resources.IdentityRecoveryAdminName(planton.Name)
	var recoverySecret corev1.Secret
	err = c.Get(ctx, types.NamespacedName{Name: name, Namespace: planton.Namespace}, &recoverySecret)
	if apierrors.IsNotFound(err) {
		// A running server refusing the admin with no recovery credential
		// in play: this server started before the operator knew to recover
		// (an operator upgrade). Stop it; the next pass runs the first half.
		log.Info("The restored realm refuses the admin and no recovery is in flight; restarting the identity server to recover")
		deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: resources.IdentityDeploymentName(planton.Name), Namespace: planton.Namespace}}
		if err := c.Delete(ctx, deploy, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("stopping the identity server for admin recovery: %w", err)
		}
		return &Result{Ready: false, Reason: v1.ComponentReasonDeploying,
			Message: "Restarting the identity server to re-establish its admin credential on the restored realm"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the identity recovery admin Secret: %w", err)
	}

	err = keycloak.ReestablishAdmin(ctx, keycloak.ReestablishAdminInput{
		HTTPClient:       httpClient,
		ServerRoot:       serverRoot,
		RecoveryUsername: resources.IdentityRecoveryAdminUsername,
		RecoveryPassword: string(recoverySecret.Data[resources.IdentityBootstrapAdminPasswordKey]),
		AdminUsername:    resources.IdentityBootstrapAdminUsername,
		AdminPassword:    adminPassword,
	})
	switch {
	case err == nil:
		log.Info("Re-established the master admin on the restored realm")
		return nil, id.finishAdminRecovery(ctx, c, planton)
	case keycloak.IsAdminNotFound(err):
		return &Result{Ready: false, Reason: v1.ComponentReasonConfigurationRefused, Message: fmt.Sprintf(
			"%s; the operator will not invent one -- sign in to the identity server's admin console as %q with the password in Secret %s and create the admin, then delete that account",
			err.Error(), resources.IdentityRecoveryAdminUsername, name)}, nil
	case keycloak.IsCredentialRefused(err):
		return &Result{Ready: false, Reason: v1.ComponentReasonReconcileFailed, Message: fmt.Sprintf(
			"the restored realm refused the recovery admin %q as well as the master admin; the recovery job %s reported success -- read its log with kubectl logs -n %s job/%s, then delete the job and Secret %s so the operator recovers again",
			resources.IdentityRecoveryAdminUsername, name, planton.Namespace, name, name)}, nil
	default:
		log.Error(err, "Re-establishing the master admin failed; will retry on next reconcile")
		return &Result{Ready: false, Reason: v1.ComponentReasonReconcileFailed, Message: fmt.Sprintf(
			"the identity server's admin credential could not be re-established on the restored realm: %v", err)}, nil
	}
}

// finishAdminRecovery records completion and removes the recovery artifacts:
// the marker ConfigMap first (the fact that matters), then the Job and the
// recovery Secret, which are inert once the recovery admin is gone.
func (id *Identity) finishAdminRecovery(ctx context.Context, c client.Client, planton *v1.PlantonPlatform) error {
	marker := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.IdentityRecoveryMarkerName(planton.Name),
			Namespace: planton.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": resources.ManagedByLabel,
			},
			OwnerReferences: []metav1.OwnerReference{*id.OwnerReferenceFor(planton)},
		},
		Data: map[string]string{
			resources.IdentityRecoveryMarkerKey: time.Now().UTC().Format(time.RFC3339),
		},
	}
	if err := c.Create(ctx, marker); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("recording the identity admin recovery: %w", err)
	}

	name := resources.IdentityRecoveryAdminName(planton.Name)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: planton.Namespace}}
	if err := c.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("removing the identity recovery admin Job: %w", err)
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: planton.Namespace}}
	if err := c.Delete(ctx, secret); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("removing the identity recovery admin Secret: %w", err)
	}
	return nil
}
