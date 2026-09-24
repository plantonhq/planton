package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// plantonPlatformCrd is the CRD the KubernetesPlantonPlatform declaration
// renders against. The operator chart owns it as a release resource and
// keeps it on uninstall by default (crds.keep_on_uninstall), so destroy
// asserts it SURVIVES unless the scenario turned keeping off.
const plantonPlatformCrd = "plantonplatforms.planton.ai"

// plantonIdentityProviderCrd is the operator chart's second definition; it
// follows the same keep dial.
const plantonIdentityProviderCrd = "plantonidentityproviders.planton.ai"

// plantonOperatorDeployment is the chart fullname with the module's fixed
// release name "planton-operator" (release name == chart name, so the
// fullname collapses to the release name).
const plantonOperatorDeployment = "planton-operator"

// PlantonOperatorInstallVerifier checks a Planton operator installation to
// the point a KubernetesPlantonPlatform declaration could be applied
// against it: the manager Deployment Available and the PlantonPlatform CRD
// Established — and THE DESIGN INVARIANT proven on every lane: NO
// PlantonPlatform exists after installing the operator alone. The operator
// never auto-creates a platform; every platform is a deliberate
// declaration, and an auto-created one here would mean the two-kind grain
// regressed into an SSA field-manager fight.
type PlantonOperatorInstallVerifier struct {
	Namespace string
	// The manifest's chart version and crds dials (both proto-JSON key
	// cases), and the shared three-part refusal check: destroy asserts the
	// definitions SURVIVE when keep_on_uninstall is true (the default) and
	// are GONE when false, so the keep, reinstall, and cleanup lanes share
	// one verifier, and a refused deploy is pinned to its class the same
	// way every CRD-carrying kind pins it.
	helmCRDLifecycle
}

// plantonOperatorDefaultChartVersion mirrors the module's default pin; the
// refusal check names the version a scenario pinned, so only a scenario
// that leaves it unset relies on this.
const plantonOperatorDefaultChartVersion = "0.8.1"

// newPlantonOperatorInstallVerifier reads the scenario manifest's chart
// version and crds dials, defaulting to the chart's own defaults when a dial
// is untouched.
func newPlantonOperatorInstallVerifier(namespace, manifestPath string) *PlantonOperatorInstallVerifier {
	return &PlantonOperatorInstallVerifier{
		Namespace: namespace,
		helmCRDLifecycle: readHelmCRDLifecycle(manifestPath, plantonOperatorDefaultChartVersion, "planton-operator",
			[]string{plantonPlatformCrd, plantonIdentityProviderCrd}),
	}
}

// VerifyExpectedDeployFailure pins a refused deploy (an unpublished chart
// version) to the shared three-part refusal.
func (v *PlantonOperatorInstallVerifier) VerifyExpectedDeployFailure(ctx context.Context, kubeconfig, expectation string, deployErr error) error {
	return v.verifyRefusal(ctx, kubeconfig, expectation, deployErr)
}

func (v *PlantonOperatorInstallVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] planton-operator in namespace %q\n", v.Namespace)

	if err := KubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for planton-operator", v.Namespace)
	}
	if err := kubectlWait(ctx, kubeconfig, "deployment", plantonOperatorDeployment, v.Namespace,
		"condition=Available", 3*time.Minute); err != nil {
		return errors.Wrap(err, "planton-operator deployment not available (a sibling operator on the cluster makes the startup guard refuse — read the pod log)")
	}
	// Available is not enough: during a chart upgrade the previous manager
	// stays available while the new one starts. The chart and the operator
	// image share one version line (chart X.Y.Z runs operator vX.Y.Z), so the
	// manager must have finished rolling out at the pinned version's tag —
	// on the first act that pins the install, on an upgrade act it is what
	// makes the upgrade an upgrade.
	operatorTag := "v" + v.ChartVersion
	if err := waitForDeploymentRolledOut(ctx, kubeconfig, v.Namespace, plantonOperatorDeployment, operatorTag, 3*time.Minute); err != nil {
		return errors.Wrapf(err, "planton-operator did not roll out at image tag %s (chart %s)", operatorTag, v.ChartVersion)
	}
	for _, crd := range []string{plantonPlatformCrd, plantonIdentityProviderCrd} {
		if err := kubectlWait(ctx, kubeconfig, "crd", crd, "",
			"condition=Established", 2*time.Minute); err != nil {
			return errors.Wrapf(err, "CRD %s not established (the chart renders it as a release resource behind crds.enabled)", crd)
		}
	}
	if err := v.verifyDefinitionsPredateManager(ctx, kubeconfig); err != nil {
		return err
	}

	// THE DESIGN INVARIANT: installing the operator alone deploys no
	// platform. Give the manager a settle window first, so a regression
	// toward startup auto-creation cannot slip under the check.
	time.Sleep(15 * time.Second)
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", plantonPlatformCrd, "-A", "-o", "name").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "listing PlantonPlatforms: %s", firstLines(string(out), 3))
	}
	if strings.TrimSpace(string(out)) != "" {
		return errors.Errorf("a PlantonPlatform exists after installing the operator alone — the operator must never auto-create a platform (found: %s)", strings.TrimSpace(string(out)))
	}
	fmt.Printf("  [verify] INVARIANT: no PlantonPlatform after install — platforms are always deliberate declarations\n")
	return nil
}

// verifyDefinitionsPredateManager proves the definitions moved IN PLACE: each
// CRD is owned by the chart's release and is no younger than the manager
// Deployment, the release's install epoch. A chart upgrade patches the
// Deployment (its creationTimestamp stays the install's) and rolls its pods;
// a definition the upgrade recreated would be created minutes after that
// epoch, and every PlantonPlatform on the cluster would have gone with it.
// On a fresh install the definitions and the Deployment land in the same
// apply (Helm orders definitions first), so "no younger" holds there too.
// The pod's start time is deliberately NOT the anchor: it is stamped in the
// same second as the definitions on a fresh install, and it moves on every
// rollout, so it cannot tell a survived definition from a recreated one. No
// state is carried across acts; the same check tells the truth on both.
func (v *PlantonOperatorInstallVerifier) verifyDefinitionsPredateManager(ctx context.Context, kubeconfig string) error {
	installed, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", plantonOperatorDeployment, v.Namespace, `{.metadata.creationTimestamp}`)
	if err != nil {
		return errors.Wrap(err, "reading the planton-operator Deployment's creation time")
	}
	epoch, err := time.Parse(time.RFC3339, strings.TrimSpace(installed))
	if err != nil {
		return errors.Wrapf(err, "planton-operator Deployment creationTimestamp %q", installed)
	}
	for _, crd := range []string{plantonPlatformCrd, plantonIdentityProviderCrd} {
		out, err := kubectlGetJSONPath(ctx, kubeconfig, "crd", crd, "",
			`{.metadata.creationTimestamp} {.metadata.annotations.meta\.helm\.sh/release-name}`)
		if err != nil {
			return errors.Wrapf(err, "reading CRD %s", crd)
		}
		fields := strings.Fields(out)
		if len(fields) != 2 {
			return errors.Errorf("CRD %s carries no Helm release ownership (got %q) -- the chart must own its definitions as release resources", crd, strings.TrimSpace(out))
		}
		created, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			return errors.Wrapf(err, "CRD %s creationTimestamp %q", crd, fields[0])
		}
		if fields[1] != v.SourceLabel {
			return errors.Errorf("CRD %s is owned by Helm release %q, not %q", crd, fields[1], v.SourceLabel)
		}
		if created.After(epoch) {
			return errors.Errorf("CRD %s was created at %s, after the operator was installed at %s -- the definition was recreated under the operator instead of upgraded in place", crd, created.Format(time.RFC3339), epoch.Format(time.RFC3339))
		}
	}
	fmt.Printf("  [verify] IN PLACE: both definitions are owned by release %q and are no younger than the operator's install\n", v.SourceLabel)
	return nil
}

func (v *PlantonOperatorInstallVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", plantonOperatorDeployment, v.Namespace); err != nil {
		return err
	}
	crds := v.CRDs
	if v.KeepOnUninstall {
		// THE KEEP POSTURE, asserted positively: the chart stamps
		// helm.sh/resource-policy: keep on its definitions while
		// crds.keep_on_uninstall is true, so removing the operator can never
		// cascade-delete platform declarations. A missing CRD here means the
		// keep semantics regressed.
		for _, crd := range crds {
			if err := KubectlResourceExists(ctx, kubeconfig, "crd", crd, ""); err != nil {
				return errors.Wrapf(err, "the %s CRD must SURVIVE the operator's destroy (crds.keep_on_uninstall is true) — its absence means the chart's keep annotation regressed", crd)
			}
		}
		fmt.Printf("  [verify] KEEP POSTURE: both definitions survived the operator's destroy\n")
		return nil
	}
	// THE CLEANUP POSTURE: keeping was explicitly disabled, so the
	// definitions leave with the release and the shared cluster is left as
	// it started.
	for _, crd := range crds {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "crd", crd, ""); err != nil {
			return errors.Wrapf(err, "the %s CRD must be GONE after a destroy with crds.keep_on_uninstall false — its presence means the keep dial was not honored", crd)
		}
	}
	fmt.Printf("  [verify] CLEANUP POSTURE: both definitions left with the release\n")
	return nil
}

// PlantonPlatformVerifier checks a declared platform to the point a person
// could sign in: the PlantonPlatform reaches phase Ready (the operator's
// own per-component gates — databases, identity, control plane, console,
// vault, runner — all pass inside it), and the two first-visit handles the
// module exports actually exist (the gateway Service and the setup-code
// Secret).
//
// Destroy relies on Kubernetes GARBAGE COLLECTION (every operator-created
// object is owner-referenced to the CR — the operator has no finalizers),
// so absence is polled: the CR's deletion returns quickly and the children
// drain asynchronously.
type PlantonPlatformVerifier struct {
	Namespace string
	Name      string
	// Version is the release the manifest declares (`spec.version`, required
	// on this kind). After the boot the platform must run exactly it: on the
	// first act that pins the install, on an upgrade act it is what makes the
	// upgrade an upgrade.
	Version string
	// VaultKeysSecret is the Secret the manifest names in
	// `spec.vault.initSecretName` for the vault's keys, or "" when the
	// operator keeps its own. When named, a Ready platform must have filled
	// it and must not own it -- the contract a team relies on to bring the
	// vault back after the platform is gone.
	VaultKeysSecret string
}

// newPlantonPlatformVerifier reads the declared version and the vault's keys
// Secret off the scenario manifest (or the upgrade manifest, on the second
// act).
func newPlantonPlatformVerifier(namespace, name, manifestPath string) *PlantonPlatformVerifier {
	version, _ := manifestSpecString(manifestPath, "version")
	keysSecret := manifestNestedSpecString(manifestPath, "vault", "initSecretName")
	if keysSecret == "" {
		keysSecret = manifestNestedSpecString(manifestPath, "vault", "init_secret_name")
	}
	return &PlantonPlatformVerifier{Namespace: namespace, Name: name, Version: version, VaultKeysSecret: keysSecret}
}

// versionedDeployments are the components whose images carry the platform's
// version; their rollouts are what a version change IS.
var versionedDeployments = []string{"control-plane", "console", "runner"}

func (v *PlantonPlatformVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] planton platform %q in namespace %q (declared version %s)\n", v.Name, v.Namespace, v.Version)

	if err := KubectlResourceExists(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "the PlantonPlatform declaration was not created")
	}

	// The operator mirrors the declared version into status once it accepts
	// it (a refused version is recorded terminally instead; waitForBoot reads
	// that). Waiting for the mirror first matters on an upgrade act: the
	// phase may still read Ready from before the declaration changed.
	if v.Version != "" {
		if err := v.waitForDeclaredVersion(ctx, kubeconfig, 3*time.Minute); err != nil {
			return err
		}
	}

	// The whole platform boots inside this one wait — databases, identity
	// server, control plane, console, secrets manager, runner. On an
	// emulated-amd64 kind cluster the full boot runs 10-15 minutes; the
	// budget leaves headroom for cold image pulls.
	fmt.Printf("  [verify] waiting for phase Ready (a full platform boot — expect ~10-15 minutes on a kind cluster)\n")
	if err := v.waitForBoot(ctx, kubeconfig, 30*time.Minute); err != nil {
		return err
	}

	// The outside-vantage half of "runs this version": every versioned
	// Deployment carries an image tagged with the declared release and has
	// finished rolling (spec observed, every replica current, none stale,
	// all available -- `kubectl rollout status`'s test). Read directly
	// rather than trusted from the phase, because an operator that judged
	// readiness by availability alone would report Ready mid-rollout while
	// the previous release still served.
	if v.Version != "" {
		for _, component := range versionedDeployments {
			if err := waitForDeploymentRolledOut(ctx, kubeconfig, v.Namespace, v.Name+"-"+component, v.Version, 15*time.Minute); err != nil {
				return err
			}
		}
		// The data layer predates the control plane that serves it: the
		// PostgreSQL cluster the operator created, and the declaration
		// itself, were created before the control-plane pod now running
		// started. On a fresh boot that is the dependency order; after an
		// upgrade it is the proof that nothing holding data was recreated.
		if err := v.verifyDataLayerPredatesControlPlane(ctx, kubeconfig); err != nil {
			return err
		}
	}

	// The first-visit handles the module exports — a Ready platform must
	// actually serve them.
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name+"-gateway", v.Namespace); err != nil {
		return errors.Wrap(err, "the front-door gateway Service is missing on a Ready platform")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "secret", v.Name+"-identity-setup-code", v.Namespace); err != nil {
		return errors.Wrap(err, "the first-run setup-code Secret is missing on a Ready platform")
	}
	fmt.Printf("  [verify] platform Ready — gateway Service and setup-code Secret present\n")

	if v.VaultKeysSecret != "" {
		if err := v.verifyVaultKeysSecret(ctx, kubeconfig); err != nil {
			return err
		}
	}
	return nil
}

// verifyVaultKeysSecret proves the keys Secret an adopter named is theirs
// the way the contract says: filled by the operator (a root token and the
// unseal keys of the built-in seal, or the recovery keys of a cloud seal),
// and carrying no owner reference, so deleting the platform leaves it
// standing for the restore that needs it.
func (v *PlantonPlatformVerifier) verifyVaultKeysSecret(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceExists(ctx, kubeconfig, "secret", v.VaultKeysSecret, v.Namespace); err != nil {
		return errors.Wrapf(err, "the vault keys Secret %s the manifest named is missing on a Ready platform", v.VaultKeysSecret)
	}
	keys, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", v.VaultKeysSecret, v.Namespace, `{range $k, $v := .data}{$k}{" "}{end}`)
	if err != nil {
		return errors.Wrapf(err, "reading the vault keys Secret %s", v.VaultKeysSecret)
	}
	held := strings.Fields(keys)
	hasRoot, hasShares := false, false
	for _, key := range held {
		switch key {
		case "root-token":
			hasRoot = true
		case "unseal-keys", "recovery-keys":
			hasShares = true
		}
	}
	if !hasRoot || !hasShares {
		return errors.Errorf("the vault keys Secret %s must hold root-token and unseal-keys (or recovery-keys under a cloud seal); it holds %v -- the operator did not write the vault's keys into the Secret the manifest named", v.VaultKeysSecret, held)
	}
	owners, _ := kubectlGetJSONPath(ctx, kubeconfig, "secret", v.VaultKeysSecret, v.Namespace, `{.metadata.ownerReferences}`)
	if strings.TrimSpace(owners) != "" {
		return errors.Errorf("the vault keys Secret %s carries an owner reference (%s) -- an adopter-owned Secret must outlive the platform, so the operator must never own it", v.VaultKeysSecret, strings.TrimSpace(owners))
	}
	fmt.Printf("  [verify] VAULT KEYS: Secret %s holds %v and is owned by nobody\n", v.VaultKeysSecret, held)
	return nil
}

// waitForDeclaredVersion waits until status.version mirrors the manifest's
// version, stopping at once if the operator refuses it.
func (v *PlantonPlatformVerifier) waitForDeclaredVersion(ctx context.Context, kubeconfig string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	for {
		status, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, `{.status.version}`)
		if strings.TrimSpace(status) == v.Version {
			fmt.Printf("  [verify] the operator accepted version %s (status.version)\n", v.Version)
			return nil
		}
		supported, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace,
			`{.status.conditions[?(@.type=="VersionSupported")].status}`)
		if strings.TrimSpace(supported) == "False" {
			msg, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace,
				`{.status.conditions[?(@.type=="VersionSupported")].message}`)
			return errors.Errorf("the operator refuses version %s (VersionSupported=False): %s", v.Version, strings.TrimSpace(msg))
		}
		if time.Now().After(deadline) {
			return errors.Errorf("status.version never mirrored the declared %s within %s (last %q)", v.Version, budget, strings.TrimSpace(status))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// deploymentRollout is the subset of a Deployment a rollout verdict reads.
type deploymentRollout struct {
	Metadata struct {
		Generation int64 `json:"generation"`
	} `json:"metadata"`
	Spec struct {
		Replicas *int64 `json:"replicas"`
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		ObservedGeneration int64 `json:"observedGeneration"`
		Replicas           int64 `json:"replicas"`
		UpdatedReplicas    int64 `json:"updatedReplicas"`
		AvailableReplicas  int64 `json:"availableReplicas"`
	} `json:"status"`
}

// deploymentRolledOutAt reports whether the Deployment (as `kubectl get -o
// json` renders it) has finished rolling to an image tagged `tag`: the spec
// observed, every desired replica on the current template, none from an
// older one, all available. spec.replicas defaults to 1 as the API server
// does.
func deploymentRolledOutAt(deployJSON []byte, tag string) (bool, error) {
	var d deploymentRollout
	if err := json.Unmarshal(deployJSON, &d); err != nil {
		return false, errors.Wrap(err, "parsing the Deployment")
	}
	if len(d.Spec.Template.Spec.Containers) == 0 || !strings.HasSuffix(d.Spec.Template.Spec.Containers[0].Image, ":"+tag) {
		return false, nil
	}
	desired := int64(1)
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}
	s := d.Status
	return desired > 0 && s.ObservedGeneration >= d.Metadata.Generation &&
		s.UpdatedReplicas == desired && s.Replicas == desired && s.AvailableReplicas == desired, nil
}

// waitForDeploymentRolledOut polls one Deployment until deploymentRolledOutAt
// holds, naming the image it last saw when the budget is spent.
func waitForDeploymentRolledOut(ctx context.Context, kubeconfig, namespace, name, tag string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastImage string
	for {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "deployment", name, "-n", namespace, "-o", "json").Output()
		if err == nil {
			done, verdictErr := deploymentRolledOutAt(out, tag)
			if verdictErr != nil {
				return errors.Wrapf(verdictErr, "deployment %s", name)
			}
			if done {
				fmt.Printf("  [verify] deployment %s rolled out at :%s\n", name, tag)
				return nil
			}
			image, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", name, namespace, `{.spec.template.spec.containers[0].image}`)
			lastImage = strings.TrimSpace(image)
		}
		if time.Now().After(deadline) {
			return errors.Errorf("deployment %s did not finish rolling to an image tagged %s within %s (last image %q)", name, tag, budget, lastImage)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}

// verifyDataLayerPredatesControlPlane asserts that the objects holding the
// platform's state -- the PostgreSQL cluster and the declaration itself --
// were created before the control-plane pod now running started.
func (v *PlantonPlatformVerifier) verifyDataLayerPredatesControlPlane(ctx context.Context, kubeconfig string) error {
	started, err := kubectlGetJSONPathList(ctx, kubeconfig, "pods", v.Namespace,
		"app.kubernetes.io/name=control-plane", `{range .items[*]}{.status.startTime}{"\n"}{end}`)
	if err != nil || len(started) == 0 {
		return errors.Wrap(err, "reading the control-plane pod's start time")
	}
	newest := time.Time{}
	for _, s := range started {
		if t, err := time.Parse(time.RFC3339, s); err == nil && t.After(newest) {
			newest = t
		}
	}
	for _, holder := range []struct{ kind, name string }{
		{"clusters.postgresql.cnpg.io", v.Name + "-postgres"},
		{plantonPlatformCrd, v.Name},
	} {
		created, err := kubectlGetJSONPath(ctx, kubeconfig, holder.kind, holder.name, v.Namespace, `{.metadata.creationTimestamp}`)
		if err != nil {
			return errors.Wrapf(err, "reading %s %s", holder.kind, holder.name)
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(created))
		if err != nil {
			return errors.Wrapf(err, "%s %s creationTimestamp %q", holder.kind, holder.name, created)
		}
		if !t.Before(newest) {
			return errors.Errorf("%s %s was created at %s, after the control-plane pod started at %s -- the data layer was recreated under the platform", holder.kind, holder.name, t.Format(time.RFC3339), newest.Format(time.RFC3339))
		}
	}
	fmt.Printf("  [verify] DATA LAYER: the PostgreSQL cluster and the declaration predate the control plane that serves them\n")
	return nil
}

// kubectlGetJSONPathList reads a jsonpath over a label-selected list and
// returns its non-empty lines.
func kubectlGetJSONPathList(ctx context.Context, kubeconfig, kind, namespace, selector, jsonPath string) ([]string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", kind, "-n", namespace, "-l", selector, "-o", "jsonpath="+jsonPath).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "kubectl get %s -l %s", kind, selector)
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// bootVerdict is what one reading of the platform's status means for the wait.
type bootVerdict int

const (
	// bootWaiting: the operator is still working (or retrying); keep polling.
	bootWaiting bootVerdict = iota
	// bootReady: every enabled component is Ready.
	bootReady
	// bootRefused: the operator will never run this declaration; stop now.
	bootRefused
)

// platformBootVerdict reads the operator's contract off one status sample.
//
// The overall phase is NOT a stop signal on its own: the operator sets phase
// Error whenever any component's reconcile returned an error in that cycle
// and requeues thirty seconds later, so a database not yet accepting
// connections or a slow image pull shows as Error for one cycle and
// recovers. Stopping on the phase would fail boots that were about to
// succeed. The one terminal refusal is the platform version floor: the
// operator sets VersionSupported=False, writes the reason on the Ready
// condition too, and returns without requeueing. That condition alone ends
// the wait early.
func platformBootVerdict(phase, versionSupported string) bootVerdict {
	switch {
	case versionSupported == "False":
		return bootRefused
	case phase == "Ready":
		return bootReady
	default:
		return bootWaiting
	}
}

// componentPhaseSummary renders the platform's per-component status (the
// JSON object under .status.components) as a stable, name-sorted line such
// as "controlPlane=Deploying identity=Ready", so a trace of the boot reads
// as progress rather than as the map's changing iteration order. Anything
// that is not that object is returned as is.
func componentPhaseSummary(rawComponents string) string {
	var components map[string]struct {
		Phase string `json:"phase"`
	}
	if err := json.Unmarshal([]byte(rawComponents), &components); err != nil || len(components) == 0 {
		return strings.TrimSpace(rawComponents)
	}
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+"="+components[name].Phase)
	}
	return strings.Join(parts, " ")
}

// waitForBoot polls the platform until it is Ready, the operator refuses the
// declared version, or the budget is spent. Component phases are printed on
// every change so a fifteen-minute boot is legible, and the operator's own
// messages travel into the returned error so the lane's failure names the
// culprit in the operator's words.
func (v *PlantonPlatformVerifier) waitForBoot(ctx context.Context, kubeconfig string, budget time.Duration) error {
	const (
		versionSupportedStatus  = `{.status.conditions[?(@.type=="VersionSupported")].status}`
		versionSupportedMessage = `{.status.conditions[?(@.type=="VersionSupported")].message}`
		readyMessage            = `{.status.conditions[?(@.type=="Ready")].message}`
		componentStatuses       = `{.status.components}`
	)
	deadline := time.Now().Add(budget)
	var lastPhase, lastComponents string
	for {
		phase, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, `{.status.phase}`)
		supported, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, versionSupportedStatus)
		phase, supported = strings.TrimSpace(phase), strings.TrimSpace(supported)

		rawComponents, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, componentStatuses)
		components := componentPhaseSummary(rawComponents)
		if phase != lastPhase || components != lastComponents {
			fmt.Printf("  [verify] phase=%s components=[%s]\n", phase, components)
			lastPhase, lastComponents = phase, components
		}

		switch platformBootVerdict(phase, supported) {
		case bootReady:
			return nil
		case bootRefused:
			msg, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, versionSupportedMessage)
			return errors.Errorf("the operator refuses this platform declaration (VersionSupported=False): %s", strings.TrimSpace(msg))
		}

		if time.Now().After(deadline) {
			msg, _ := kubectlGetJSONPath(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace, readyMessage)
			out, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
				"get", plantonPlatformCrd, v.Name, "-n", v.Namespace,
				"-o", "jsonpath={.status.components}").CombinedOutput()
			return errors.Errorf("the platform never reached Ready within %s (phase %s: %s); component status: %s",
				budget, phase, strings.TrimSpace(msg), firstLines(string(out), 6))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
}

func (v *PlantonPlatformVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	// The CR's own deletion returns quickly (no operator finalizer);
	// the owner-referenced children drain through garbage collection
	// asynchronously — poll both with a bounded budget.
	deadline := time.Now().Add(5 * time.Minute)
	for {
		crErr := KubectlResourceAbsent(ctx, kubeconfig, plantonPlatformCrd, v.Name, v.Namespace)
		gwErr := KubectlResourceAbsent(ctx, kubeconfig, "service", v.Name+"-gateway", v.Namespace)
		cpErr := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name+"-control-plane", v.Namespace)
		if crErr == nil && gwErr == nil && cpErr == nil {
			fmt.Printf("  [verify] platform gone — CR deleted, children garbage-collected\n")
			return nil
		}
		if time.Now().After(deadline) {
			for _, err := range []error{crErr, gwErr, cpErr} {
				if err != nil {
					return errors.Wrap(err, "platform teardown did not complete within the garbage-collection budget")
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
