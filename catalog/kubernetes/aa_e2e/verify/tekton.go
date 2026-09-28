package verify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// tektonOperatorNamespace is the release manifest's fixed installation
// namespace (baked into its own cross-references — not configurable).
const tektonOperatorNamespace = "tekton-operator"

// tektonConfigCrd is the CRD the KubernetesTekton declaration renders
// against; it installs and deletes WITH the operator (the manifest-bundle
// posture).
const tektonConfigCrd = "tektonconfigs.operator.tekton.dev"

// TektonOperatorVerifier checks a Tekton Operator install to the point a
// KubernetesTekton declaration could be applied against it: the operator
// serving the moment its deploy reports done (asked FIRST and without any
// wait, so this verifier never supplies the readiness the module owes),
// the operator and webhook Deployments rolled out, the operator.tekton.dev
// CRDs established — and THE DESIGN INVARIANT proven on every lane: NO
// TektonConfig exists after install. The module patches the release's
// AUTOINSTALL_COMPONENTS to "false" so the KubernetesTekton declaration
// is the single owner of the cluster's Tekton configuration; an
// auto-created TektonConfig here means that patch regressed (the SSA
// field-manager fight the two-kind grain exists to prevent).
type TektonOperatorVerifier struct {
	Name string
}

func (v *TektonOperatorVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] tekton-operator %q in namespace %q\n", v.Name, tektonOperatorNamespace)

	if err := tektonOperatorServes(ctx, kubeconfig); err != nil {
		return err
	}
	fmt.Printf("  [verify] READY MEANS SERVING: a TektonConfig came back defaulted the moment the deploy reported done\n")

	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-operator", tektonOperatorNamespace, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the operator deployment never rolled out")
	}
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-operator-webhook", tektonOperatorNamespace, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the operator webhook deployment never rolled out")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "crd", tektonConfigCrd, ""); err != nil {
		return errors.Wrap(err, "the TektonConfig CRD was not installed")
	}

	// THE DESIGN INVARIANT: auto-install disabled, so installing the
	// operator alone deploys no Tekton components and creates no
	// TektonConfig. Give the operator a settle window first — the
	// auto-install path (were it regressed) runs at controller startup.
	time.Sleep(15 * time.Second)
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "tektonconfigs.operator.tekton.dev", "-o", "name").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "listing TektonConfigs: %s", firstLines(string(out), 3))
	}
	if strings.TrimSpace(string(out)) != "" {
		return errors.Errorf("a TektonConfig exists after installing the operator alone — the AUTOINSTALL_COMPONENTS=false patch regressed (found: %s)", strings.TrimSpace(string(out)))
	}
	fmt.Printf("  [verify] INVARIANT: no TektonConfig after install — auto-install is disabled, the declaration kind owns the configuration\n")
	return nil
}

// tektonReadinessProbe is a TektonConfig declaring only its target
// namespace: the operator's defaulting webhook fills in spec.profile, and
// nothing else would.
const tektonReadinessProbe = `apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: config
spec:
  targetNamespace: tekton-pipelines
`

// tektonOperatorServes asks, once and with no wait, whether the API server
// admits a TektonConfig with the operator's defaults filled in -- a
// server-side dry run, so nothing is written and the no-TektonConfig
// invariant below still holds. The module's readiness gate owes exactly
// this before the deploy reports done; a deploy that returned while the
// operator was still starting fails here, where the verifier's own
// rollout waits would otherwise have carried it (the cold-node-pool
// scenario holds the pods off a node to make that window real).
func tektonOperatorServes(ctx context.Context, kubeconfig string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"create", "--dry-run=server", "-o", "jsonpath={.spec.profile}", "-f", "-")
	cmd.Stdin = strings.NewReader(tektonReadinessProbe)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return errors.Errorf("the deploy reported done but the operator cannot serve yet: a TektonConfig applied now would be refused (the readiness gate did not hold): %s", firstLines(stderr.String(), 3))
	}
	if strings.TrimSpace(string(out)) == "" {
		return errors.New("the deploy reported done but the operator's webhook is not filling in TektonConfig defaults yet: a KubernetesTekton applied now would be refused (the readiness gate did not hold)")
	}
	return nil
}

func (v *TektonOperatorVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", "tekton-operator", tektonOperatorNamespace); err != nil {
		return err
	}
	// The manifest-bundle posture: every document deletes with the
	// resource, INCLUDING the CRDs. The tektoninstallersets CRD is
	// checked with a bounded wait: its deletion DRAINS the operator's
	// runtime InstallerSet CRs, whose finalizer the (deliberately
	// still-alive, destroy-ordered) operator processes asynchronously —
	// a one-shot check would race the tail of that drain.
	deadlineCrd := time.Now().Add(2 * time.Minute)
	for {
		err := KubectlResourceAbsent(ctx, kubeconfig, "crd", "tektoninstallersets.operator.tekton.dev", "")
		if err == nil {
			break
		}
		if time.Now().After(deadlineCrd) {
			return errors.Wrap(err, "the tektoninstallersets CRD never finished draining after destroy — the operator-processed finalizer ordering is broken")
		}
		time.Sleep(5 * time.Second)
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "crd", tektonConfigCrd, ""); err != nil {
		return errors.Wrap(err, "the operator.tekton.dev CRDs must delete with the operator (the manifest-bundle posture)")
	}
	// The FIXED namespace must be FULLY GONE, not merely Terminating —
	// namespace deletion is asynchronous, the name is baked into the
	// manifest, and a lingering Terminating namespace makes the next
	// install fail with "unable to create new content in namespace ...
	// because it is being terminated" (verified live: back-to-back
	// lanes raced exactly this window). Zero-orphan for a
	// fixed-namespace kind includes the namespace itself.
	deadline := time.Now().Add(3 * time.Minute)
	for {
		err := KubectlResourceAbsent(ctx, kubeconfig, "namespace", tektonOperatorNamespace, "")
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return errors.Wrap(err, "the fixed tekton-operator namespace never finished terminating after destroy")
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("  [verify] DESTROY: operator workloads, CRDs and the fixed namespace gone (the manifest-bundle posture)\n")
	return nil
}

// TektonVerifier checks a Tekton installation (the TektonConfig
// declaration the operator reconciles) to the point a customer could run
// pipelines on it: TektonConfig Ready, the per-profile component
// Deployments rolled out (dashboard on `all` only — profile honesty is
// asserted both ways), the pruner CronJob present when declared — and
// THE ENGINE PROOF on every lane: a verifier-owned TaskRun runs to
// Succeeded (a pipeline engine that cannot run a task is not a pipeline
// engine).
//
// Destroy proves the posture the two-kind grain exists for: the
// TektonConfig deletion completed while the operator was still running
// (the module blocks on the operator-processed finalizer), so no
// TektonInstallerSet is left behind.
type TektonVerifier struct {
	Name string
	// Profile is the resolved profile (lite/basic/all).
	Profile string
	// TargetNamespace is where the operator installs the components.
	TargetNamespace string
	// PrunerDeclared gates the pruner CronJob assertion.
	PrunerDeclared bool
}

// tektonProfile reads spec.profile (default "all" — the proto/operator
// default).
func tektonProfile(spec map[string]interface{}) string {
	if raw, ok := spec["profile"].(string); ok && raw != "" {
		return raw
	}
	return "all"
}

// tektonTargetNamespace reads spec.target_namespace (default
// "tekton-pipelines" — the upstream default; both manifest key forms
// tolerated).
func tektonTargetNamespace(spec map[string]interface{}) string {
	for _, key := range []string{"target_namespace", "targetNamespace"} {
		if raw, ok := spec[key].(string); ok && raw != "" {
			return raw
		}
	}
	return "tekton-pipelines"
}

// tektonPrunerDeclared reports whether spec.pruner is present.
func tektonPrunerDeclared(spec map[string]interface{}) bool {
	pruner, ok := spec["pruner"].(map[string]interface{})
	return ok && len(pruner) > 0
}

func (v *TektonVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] tekton %q — profile %q in namespace %q\n", v.Name, v.Profile, v.TargetNamespace)

	// TektonConfig Ready: the operator reconciles the declaration into
	// running components through TektonInstallerSets; Ready=True is its
	// own all-components-installed signal. Cold clusters pull every
	// component image here — the budget is generous.
	if err := v.awaitTektonConfigReady(ctx, kubeconfig, 15*time.Minute); err != nil {
		return err
	}

	// Pipelines runs on every profile.
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-pipelines-controller", v.TargetNamespace, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the pipelines controller never rolled out")
	}
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-pipelines-webhook", v.TargetNamespace, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the pipelines webhook never rolled out")
	}

	// Profile honesty — asserted BOTH ways.
	if v.Profile == "basic" || v.Profile == "all" {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-triggers-controller", v.TargetNamespace, 5*time.Minute); err != nil {
			return errors.Wrapf(err, "profile %q installs Triggers but its controller never rolled out", v.Profile)
		}
	}
	if v.Profile == "all" {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/tekton-dashboard", v.TargetNamespace, 5*time.Minute); err != nil {
			return errors.Wrap(err, "profile \"all\" installs the Dashboard but it never rolled out")
		}
	} else {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", "tekton-dashboard", v.TargetNamespace); err != nil {
			return errors.Wrapf(err, "profile %q must NOT install the Dashboard", v.Profile)
		}
		fmt.Printf("  [verify] PROFILE: dashboard correctly ABSENT on profile %q\n", v.Profile)
	}

	if v.PrunerDeclared {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "cronjobs", "-n", v.TargetNamespace, "-o", "name").CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "listing cronjobs: %s", firstLines(string(out), 3))
		}
		// The operator generates the pruner CronJob with the fixed
		// tekton-resource-pruner name prefix (operator source at the pin).
		if !strings.Contains(string(out), "tekton-resource-pruner") {
			return errors.Errorf("the pruner was declared but no tekton-resource-pruner CronJob exists in %s (found: %s)", v.TargetNamespace, firstLines(string(out), 3))
		}
		fmt.Printf("  [verify] PRUNER: tekton-resource-pruner CronJob present\n")
	}

	// THE ENGINE PROOF — on every lane.
	return v.proveTaskRun(ctx, kubeconfig)
}

func (v *TektonVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	// The TektonConfig is gone (its deletion blocked on the
	// operator-processed finalizer completing the component teardown) —
	// and with it every TektonInstallerSet. A leftover InstallerSet is
	// exactly the stranded-finalizer class the two-kind destroy ordering
	// exists to prevent.
	if err := KubectlResourceAbsent(ctx, kubeconfig, "tektonconfigs.operator.tekton.dev", "config", ""); err != nil {
		return err
	}
	// The InstallerSet teardown is an ASYNC cascade (the TektonConfig
	// finalizer removes component CRs, whose own finalizers remove their
	// InstallerSets) that keeps converging for a short window after the
	// TektonConfig object is gone — so this asserts CONVERGENCE within a
	// bounded budget, never a one-shot snapshot (verified live: an
	// instant check reported five in-flight InstallerSets as stranded).
	// A genuinely stranded finalizer never converges and still fails.
	//
	// The operator's OWN webhook InstallerSet is excluded by its type
	// label: it belongs to the OPERATOR's lifetime, not the TektonConfig
	// teardown — the webhook maintains it from startup even with zero
	// TektonConfigs (verified live and in the operator source), and it
	// leaves with the KubernetesTektonOperator resource.
	deadlineSets := time.Now().Add(3 * time.Minute)
	for {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "tektoninstallersets.operator.tekton.dev",
			"-l", "operator.tekton.dev/type!=operatorValidatingDefaultingWebhook",
			"-o", "name").CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) == "" {
			break
		}
		if time.Now().After(deadlineSets) {
			return errors.Errorf("TektonInstallerSets survived the TektonConfig teardown (the stranded-finalizer class): %s", firstLines(string(out), 5))
		}
		time.Sleep(5 * time.Second)
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", "tekton-pipelines-controller", v.TargetNamespace); err != nil {
		return err
	}
	// The operator DELETES its target namespace with the TektonConfig
	// teardown (live-observed — offline-unverifiable before the first
	// lane). Wait for it to be FULLY gone: a namespace lingering in
	// Terminating (e.g. on an unprocessed content finalizer) wedges the
	// next lane that targets the same fixed name.
	deadlineNs := time.Now().Add(5 * time.Minute)
	for {
		err := KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.TargetNamespace, "")
		if err == nil {
			break
		}
		if time.Now().After(deadlineNs) {
			return errors.Wrapf(err, "the operator-created target namespace %q never finished terminating after the TektonConfig teardown", v.TargetNamespace)
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("  [verify] DESTROY: TektonConfig, InstallerSets, component workloads and the target namespace gone — the operator-alive teardown completed cleanly\n")
	return nil
}

// awaitTektonConfigReady polls the cluster singleton TektonConfig
// (`config` — the operator-required fixed name) for the Ready condition.
func (v *TektonVerifier) awaitTektonConfigReady(ctx context.Context, kubeconfig string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	lastState := "(no status yet)"
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "tektonconfig", "config",
			"-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`).CombinedOutput()
		if err == nil {
			state := strings.TrimSpace(string(out))
			if state != "" {
				lastState = state
			}
			if state == "True" {
				fmt.Printf("  [verify] READY: TektonConfig reconciled to Ready\n")
				return nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("TektonConfig never reached Ready within %s (last Ready status %q)", timeout, lastState)
}

// proveTaskRun applies a verifier-owned TaskRun (tekton.dev/v1, busybox
// echo) in the target namespace and waits for the Succeeded condition.
func (v *TektonVerifier) proveTaskRun(ctx context.Context, kubeconfig string) error {
	const runName = "e2e-proof-taskrun"
	manifest := fmt.Sprintf(`apiVersion: tekton.dev/v1
kind: TaskRun
metadata:
  name: %s
  namespace: %s
spec:
  taskSpec:
    steps:
      - name: main
        image: busybox:1.37
        command: ["sh", "-c", "echo e2e-proof-ok"]
`, runName, v.TargetNamespace)

	path, err := writeVerifierTempManifest("tekton-taskrun-proof", manifest)
	if err != nil {
		return errors.Wrap(err, "writing the proof TaskRun manifest")
	}
	defer func() { _ = os.Remove(path) }()

	if err := kubectlApplyFile(ctx, kubeconfig, path); err != nil {
		return errors.Wrap(err, "submitting the proof TaskRun")
	}
	fmt.Printf("  [verify] SUBMIT: TaskRun %q submitted\n", runName)

	// Zero-orphan duty: the TaskRun (and the pod it owns) leave with the
	// verifier — CONFIRMED gone, not fire-and-forget. Tekton Results
	// guards TaskRuns with its `results.tekton.dev/taskrun` archival
	// finalizer, and a proof TaskRun still pending that finalizer when
	// the destroy phase kills the results watcher can never finish
	// deleting — it wedges the operator-created target namespace in
	// Terminating and poisons every later lane (verified live). The
	// deletion therefore completes HERE, while the stack serving the
	// finalizer is alive; if the finalizer still never clears within the
	// budget, it is stripped — the archival guarantee is meaningless for
	// a verifier-owned proof object on its way out.
	defer func() {
		delCmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"delete", "taskrun", runName, "-n", v.TargetNamespace, "--ignore-not-found", "--wait=false")
		_, _ = delCmd.CombinedOutput()
		goneDeadline := time.Now().Add(2 * time.Minute)
		for time.Now().Before(goneDeadline) {
			if KubectlResourceAbsent(ctx, kubeconfig, "taskrun", runName, v.TargetNamespace) == nil {
				return
			}
			time.Sleep(5 * time.Second)
		}
		patchCmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"patch", "taskrun", runName, "-n", v.TargetNamespace,
			"-p", `{"metadata":{"finalizers":[]}}`, "--type=merge")
		_, _ = patchCmd.CombinedOutput()
		fmt.Printf("  [verify] CLEANUP: proof TaskRun finalizers stripped after the deletion budget — a lingering archival finalizer must never outlive the lane\n")
	}()

	deadline := time.Now().Add(6 * time.Minute)
	lastStatus := "(no status yet)"
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "taskrun", runName, "-n", v.TargetNamespace,
			"-o", `jsonpath={.status.conditions[?(@.type=="Succeeded")].status},{.status.conditions[?(@.type=="Succeeded")].reason}`).CombinedOutput()
		if err == nil {
			parts := strings.SplitN(strings.TrimSpace(string(out)), ",", 2)
			status := parts[0]
			if status != "" {
				lastStatus = strings.TrimSpace(string(out))
			}
			switch status {
			case "True":
				fmt.Printf("  [verify] RUN: TaskRun %q ran to Succeeded — the engine executes tasks\n", runName)
				return nil
			case "False":
				return errors.Errorf("the proof TaskRun %q failed: %s", runName, lastStatus)
			}
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("the proof TaskRun %q never completed (last condition %q)", runName, lastStatus)
}
