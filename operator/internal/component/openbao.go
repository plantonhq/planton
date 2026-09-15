package component

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/bootstrap"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// OpenBAO deploys and monitors OpenBAO (open-source Vault fork), the bundled
// secrets manager. Deployed by default (spec.vault): it backs the credential
// store for pasted connection secrets, the default envelope-encryption key,
// and the OIDC issuer's signing key -- integral the way the database is.
// spec.vault.enabled: false is the deliberate opt-out.
//
// The vault keeps its data in the platform's PostgreSQL -- its own database,
// its own least-privilege role, both declared by the database component --
// so it has no volume, the platform's one archive carries every secret with
// the records, and a restored database brings back a vault that finds itself
// initialized. What opens it is the seal the declaration chose at creation:
// the built-in key shares, which the operator holds in the init Secret and
// unseals with after every restart and every restore; or a key in the
// adopter's cloud (spec.vault.autoUnseal), which the server opens itself
// with and the operator never touches. The init Secret is the adopter's own
// when spec.vault.initSecretName names one (written once, never deleted by
// the operator) and the operator's otherwise. There is exactly one
// initialization path -- the platform's database is not initialized by hand
// either -- and the component waits for the vault's role and database to
// exist before the server is rendered (the backend gives up after one
// failed connect).
type OpenBAO struct{ Base }

func (o *OpenBAO) Name() string { return "openbao" }

// Dependencies: the vault's storage IS the platform's database, so it waits
// for the database component to report Ready (the Cluster up, its credentials
// present) before it renders at all.
func (o *OpenBAO) Dependencies(_ *v1.PlantonPlatform) []string { return []string{"postgresql"} }

// IsEnabled defaults to true: absence of spec.vault means deploy the bundled
// secrets manager. Must agree with isVaultEnabled (control_plane.go) and
// isOpenBAOEnabled (status package), or the reconciler, the control-plane
// wiring, and the status slot disagree about the component's existence.
func (o *OpenBAO) IsEnabled(planton *v1.PlantonPlatform) bool {
	return planton.Spec.Vault == nil || planton.Spec.Vault.Enabled == nil ||
		*planton.Spec.Vault.Enabled
}

// sealOpenWait is how long a pass waits, after initializing a vault under a
// cloud seal, for the server's own loop to open it: the server retries the
// stored key every five seconds, so one interval plus margin covers the
// common case in the same pass; a vault still sealed after that is reported
// and the next pass finds it open.
const sealOpenWait = 8 * time.Second

func (o *OpenBAO) Reconcile(ctx context.Context, c client.Client, _ *runtime.Scheme, planton *v1.PlantonPlatform) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", o.Name())

	// The vault's role and database are the database component's to declare
	// and CloudNativePG's to reconcile; this component starts the server only
	// once both exist, because the storage backend gives up after one failed
	// connect and a crash-looping pod would hide the real wait behind a log.
	ready, waiting, err := o.vaultDatabaseReady(ctx, c, planton)
	if err != nil {
		return Result{}, err
	}
	if !ready {
		return Result{
			Ready:   false,
			Reason:  v1.ComponentReasonDeploying,
			Object:  &v1.ComponentObjectReference{Kind: "Database", Name: resources.PostgreSQLVaultDatabaseObjectName(planton.Name)},
			Message: waiting,
		}, nil
	}

	seal := sealOptionsFrom(planton)
	refused, initSecret, err := o.preflightSeal(ctx, c, planton, seal)
	if err != nil {
		return Result{}, err
	}
	if refused != nil {
		return *refused, nil
	}

	chartData := resources.LoadOpenBAOChart()
	values := resources.OpenBAOHelmValues(resources.OpenBAOHelmOptions{
		CRName:                    planton.Name,
		Namespace:                 planton.Namespace,
		StoragePasswordSecretName: resources.PostgreSQLVaultRoleSecretName(planton.Name),
		Seal:                      seal,
		ServiceAccountAnnotations: vaultServiceAccountAnnotations(planton),
	})

	releaseName := fmt.Sprintf("%s-openbao", planton.Name)
	rendered, err := resources.RenderHelmChart(chartData, releaseName, planton.Namespace, values)
	if err != nil {
		return Result{}, fmt.Errorf("rendering OpenBAO chart: %w", err)
	}

	if err := o.ApplyManifests(ctx, c, planton, rendered); err != nil {
		return Result{}, fmt.Errorf("applying OpenBAO manifests: %w", err)
	}

	// OpenBAO readiness probes fail when sealed/uninitialized, so checking
	// StatefulSet readiness would deadlock. Instead, check if the pod is in
	// Running phase (container started, API accessible) before proceeding
	// to initialization. The chart's fullnameOverride makes the StatefulSet's
	// name the release name.
	podRunning, err := o.IsPodRunning(ctx, c, releaseName, planton.Namespace)
	if err != nil {
		return Result{}, fmt.Errorf("checking OpenBAO pod status: %w", err)
	}
	if !podRunning {
		log.Info("OpenBAO pod not yet running")
		return o.notReadyWithSealHint(ctx, c, planton, seal, releaseName, "Waiting for OpenBAO pod"), nil
	}

	return o.ensureInitialized(ctx, c, planton, seal, initSecret, resources.OpenBAOAPIAddr(planton.Name, planton.Namespace), http.DefaultClient)
}

// preflightSeal is everything checked BEFORE the chart renders, so that a
// declaration the vault cannot honor is a sentence and never a pod that
// cannot start: the credentials Secret a seal arm names must exist with its
// key, and the declared seal must be the one the vault was initialized under
// (the init Secret records it; the server would refuse to start against its
// own storage under another). Returns the refusal, or the init Secret as
// read (nil when it does not exist) for the arms that follow.
func (o *OpenBAO) preflightSeal(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions) (*Result, *corev1.Secret, error) {
	initSecretName := vaultInitSecretName(planton)

	if msg, err := o.preflightSealCredentialsSecret(ctx, c, planton.Namespace, seal); err != nil {
		return nil, nil, err
	} else if msg != "" {
		return &Result{
			Ready:   false,
			Reason:  v1.ComponentReasonConfigurationRefused,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: seal.CredentialsSecretName()},
			Message: msg,
		}, nil, nil
	}

	initSecret, err := readInitSecret(ctx, c, initSecretName, planton.Namespace)
	if err != nil {
		return nil, nil, err
	}
	if initSecret != nil {
		if recorded := initSecret.Annotations[resources.OpenBAOSealAnnotation]; recorded != "" && recorded != seal.Fingerprint() {
			return &Result{
				Ready:   false,
				Reason:  v1.ComponentReasonConfigurationRefused,
				Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
				Message: sealChangedMessage(recorded, seal.Fingerprint(), initSecretName),
			}, initSecret, nil
		}
	}
	return nil, initSecret, nil
}

// notReadyWithSealHint classifies the vault's workload through the shared
// explainer and, when the verdict is a crash-loop under a cloud seal, adds
// the one clause the classifier cannot know: the server configures the seal
// before anything else and exits when the wrapper's first call fails.
func (o *OpenBAO) notReadyWithSealHint(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, releaseName, waiting string) Result {
	res := o.NotReady(ctx, c, planton.Namespace, StatefulSetRef(releaseName), waiting)
	if res.Reason == v1.ComponentReasonCrashLooping && seal != nil {
		res.Message += " " + sealStartHint(seal)
	}
	return res
}

// ensureInitialized drives the vault from whatever state its storage is in
// to open and usable: a fresh vault is initialized in the shape its seal
// requires and its keys written to the init Secret; a sealed vault under the
// built-in seal is unsealed from that Secret; a sealed vault under a cloud
// seal is the server's to open, and the sentence says whether it is opening
// or blocked; an open vault has its engines ensured and its Secret's note
// re-asserted. apiAddr and httpClient are parameters so the flow is proven
// against a fake server.
func (o *OpenBAO) ensureInitialized(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, initSecret *corev1.Secret, apiAddr string, httpClient *http.Client) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", o.Name())
	releaseName := fmt.Sprintf("%s-openbao", planton.Name)

	health, err := bootstrap.CheckOpenBAOHealth(ctx, httpClient, apiAddr)
	if err != nil {
		// A pod that runs but does not answer is, above all, a server that
		// keeps exiting -- and a crash-looping pod still reports phase
		// Running. The classifier names the pod and its exit; the seal
		// hint names the check that most likely failed.
		log.Info("OpenBAO health check failed (may not be ready yet)", "error", err.Error())
		return o.notReadyWithSealHint(ctx, c, planton, seal, releaseName, "Waiting for OpenBAO health endpoint"), nil
	}

	switch {
	case health.Initialized && !health.Sealed:
		return o.reconcileOpenVault(ctx, c, planton, seal, initSecret, apiAddr, httpClient)

	case !health.Initialized:
		return o.initializeVault(ctx, c, planton, seal, initSecret, apiAddr, httpClient)

	default: // initialized and sealed
		if seal != nil {
			// The operator never unseals a cloud-sealed vault; the server
			// does, retrying its key every five seconds. Young pod: it is
			// opening. Old pod: the identity cannot decrypt with the key.
			age := o.vaultPodAge(ctx, c, releaseName, planton.Namespace, time.Now())
			return Result{
				Ready:   false,
				Reason:  v1.ComponentReasonDeploying,
				Object:  &v1.ComponentObjectReference{Kind: "StatefulSet", Name: releaseName},
				Message: sealNotOpenedMessage(seal, age),
			}, nil
		}
		return o.unsealFromInitSecret(ctx, c, planton, initSecret, apiAddr, httpClient)
	}
}

// initializeVault runs the one initialization path: /sys/init in the shape
// the seal requires, the keys and root token written to the init Secret,
// then the vault opened -- by the operator with the shares under the
// built-in seal, by the server itself under a cloud seal -- and the engines
// mounted. A Secret that already holds keys is refused first: writing new
// keys over another vault's is the one thing this path must never do.
func (o *OpenBAO) initializeVault(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, initSecret *corev1.Secret, apiAddr string, httpClient *http.Client) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", o.Name())
	initSecretName := vaultInitSecretName(planton)

	if initSecretHoldsKeys(initSecret) {
		return Result{
			Ready:   false,
			Reason:  v1.ComponentReasonConfigurationRefused,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
			Message: initSecretHoldsForeignKeysMessage(planton, initSecretName),
		}, nil
	}

	log.Info("OpenBAO not initialized, initializing", "seal", seal.Word())
	initOpts := bootstrap.ShamirInit(resources.OpenBAOSecretShares, resources.OpenBAOSecretThreshold)
	if seal != nil {
		initOpts = bootstrap.RecoveryInit(resources.OpenBAOSecretShares, resources.OpenBAOSecretThreshold)
	}
	initResult, err := bootstrap.InitializeOpenBAO(ctx, httpClient, apiAddr, initOpts)
	if err != nil {
		return Result{}, fmt.Errorf("initializing OpenBAO: %w", err)
	}

	// From here until the write lands the keys exist only in this pass.
	if err := o.writeInitSecret(ctx, c, planton, seal, initResult); err != nil {
		log.Error(err, "OpenBAO initialized but its keys could not be written", "secret", initSecretName)
		return Result{
			Ready:   false,
			Reason:  v1.ComponentReasonReconcileFailed,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
			Message: initSecretWriteFailedMessage(seal, initSecretName, err),
		}, nil
	}
	log.Info("OpenBAO initialized, init secret written", "secret", initSecretName)

	if seal == nil {
		if err := bootstrap.UnsealOpenBAO(ctx, httpClient, apiAddr, initResult.UnsealKeys); err != nil {
			return Result{}, fmt.Errorf("unsealing OpenBAO: %w", err)
		}
	} else {
		open, err := bootstrap.WaitUntilUnsealed(ctx, httpClient, apiAddr, sealOpenWait)
		if err != nil {
			return Result{}, fmt.Errorf("waiting for the seal to open OpenBAO: %w", err)
		}
		if !open {
			return Result{
				Ready:   false,
				Reason:  v1.ComponentReasonDeploying,
				Object:  &v1.ComponentObjectReference{Kind: "StatefulSet", Name: fmt.Sprintf("%s-openbao", planton.Name)},
				Message: sealNotOpenedMessage(seal, 0),
			}, nil
		}
	}

	if err := bootstrap.EnsureOpenBAOMounts(ctx, httpClient, apiAddr, initResult.RootToken); err != nil {
		return Result{}, fmt.Errorf("ensuring OpenBAO secrets engines: %w", err)
	}

	log.Info("OpenBAO initialized and open")
	return Result{Ready: true, Message: "OpenBAO healthy (initialized)"}, nil
}

// unsealFromInitSecret opens a built-in-seal vault that is initialized and
// sealed -- a restarted pod, or a restored vault -- with the shares from the
// init Secret. No Secret is the restored case above all (the data came back
// from the archive, the keys did not) and is refused with the Secret named.
func (o *OpenBAO) unsealFromInitSecret(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, initSecret *corev1.Secret, apiAddr string, httpClient *http.Client) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", o.Name())
	initSecretName := vaultInitSecretName(planton)

	if initSecret == nil {
		return Result{
			Ready:   false,
			Reason:  v1.ComponentReasonConfigurationRefused,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
			Message: sealedWithoutKeysMessage(planton, initSecretName),
		}, nil
	}

	var unsealKeys []string
	if err := json.Unmarshal(initSecret.Data[resources.OpenBAOInitSecretUnsealKeysKey], &unsealKeys); err != nil {
		return Result{}, fmt.Errorf("parsing unseal keys from secret %s: %w", initSecretName, err)
	}
	if err := bootstrap.UnsealOpenBAO(ctx, httpClient, apiAddr, unsealKeys); err != nil {
		return Result{}, fmt.Errorf("unsealing OpenBAO: %w", err)
	}
	if err := bootstrap.EnsureOpenBAOMounts(ctx, httpClient, apiAddr, string(initSecret.Data[resources.OpenBAOInitSecretRootTokenKey])); err != nil {
		return Result{}, fmt.Errorf("ensuring OpenBAO secrets engines: %w", err)
	}

	log.Info("OpenBAO unsealed from stored keys")
	return Result{Ready: true, Message: "OpenBAO healthy (unsealed)"}, nil
}

// reconcileOpenVault is the steady state. The engine mounts are ensured every
// pass (idempotent GET + enable-if-missing) with the root token from the init
// Secret, and the Secret's note and fingerprint are re-asserted (annotations
// only). A vault that is open while its Secret does not exist is refused
// with the Secret named: the operator and the control plane sign in with the
// token it holds.
func (o *OpenBAO) reconcileOpenVault(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, initSecret *corev1.Secret, apiAddr string, httpClient *http.Client) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", o.Name())
	initSecretName := vaultInitSecretName(planton)

	if initSecret == nil {
		return Result{
			Ready:   false,
			Reason:  v1.ComponentReasonConfigurationRefused,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
			Message: openVaultWithoutSecretMessage(planton, seal, initSecretName),
		}, nil
	}

	if err := bootstrap.EnsureOpenBAOMounts(ctx, httpClient, apiAddr, string(initSecret.Data[resources.OpenBAOInitSecretRootTokenKey])); err != nil {
		return Result{}, fmt.Errorf("ensuring OpenBAO secrets engines: %w", err)
	}
	if err := o.ensureInitSecretAnnotations(ctx, c, planton, seal, initSecret); err != nil {
		return Result{}, err
	}

	log.Info("OpenBAO ready (initialized and unsealed)")
	return Result{Ready: true, Message: "OpenBAO healthy"}, nil
}
