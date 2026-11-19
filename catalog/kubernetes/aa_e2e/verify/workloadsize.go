package verify

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// The replica-ownership lanes prove the promise behind a workload's
// `availability.replicas`: while an autoscaler owns the live count (the inline
// HPA, or an external scaler when replicas is unset), an apply never moves it.
// Each lane has two acts. Act one deploys, then changes the size the way an
// autoscaler does: a write to the /scale subresource under another field
// manager (kubectl's), the exact write an HPA or KEDA makes. Act two re-applies
// with a changed environment variable, so the engine really updates the
// workload, and proves the size the "autoscaler" chose survived.
//
// The lanes rely on the kind cluster serving no metrics API: an HPA without
// metrics enforces its bounds and leaves any size between them alone, so a
// size act one sets stays put unless the apply moves it.
type workloadSizeLane string

const (
	// The inline HPA (floor 2, max 5) owns the size; act one scales to 4, and
	// act two (floor raised to 3) must keep 4.
	sizeLaneHpaKeepsSize workloadSizeLane = "behavioral-hpa-keeps-size"
	// Turning the inline HPA on (floor 3) for a Deployment running 3 replicas
	// must never drop it below 3, not even for the HPA's first sync.
	sizeLaneHpaEnableKeepsSize workloadSizeLane = "behavioral-hpa-enable-keeps-size"
	// replicas unset: an external scaler owns the size; act one scales to 3,
	// and act two must keep 3.
	sizeLaneExternalScalerKeepsSize workloadSizeLane = "behavioral-external-scaler-keeps-size"
)

const (
	// sizeMarkerEnv is the variable every second act adds, so the re-apply is a
	// real update of the workload, not a no-op the engine skips.
	sizeMarkerEnv   = "E2E_ACT"
	sizeMarkerValue = "two"
	// hpaBelowFloorReason is the HPA's rescale reason when it finds the workload
	// below its floor -- the trace an apply that reset the size to 1 leaves.
	hpaBelowFloorReason = "below Spec.MinReplicas"
	// hpaSyncWindow covers two HPA sync periods (15s each): long enough for
	// the HPA to notice and correct a reset that happened during the apply.
	hpaSyncWindow = 35 * time.Second
)

// workloadSizeLaneFor names the lane a scenario runs, from its manifest path
// (both acts' files carry the lane's name).
func workloadSizeLaneFor(manifestPath string) workloadSizeLane {
	for _, lane := range []workloadSizeLane{sizeLaneHpaKeepsSize, sizeLaneHpaEnableKeepsSize, sizeLaneExternalScalerKeepsSize} {
		if strings.Contains(manifestPath, string(lane)) {
			return lane
		}
	}
	return ""
}

// proveSize runs the lane's act: act one sets the size an autoscaler would,
// act two (the runner marks it with the first act's manifest path) proves the
// re-apply kept it.
func (v *WorkloadVerifier) proveSize(ctx context.Context, kubeconfig string) error {
	secondAct := ctx.Value(provider.FirstActManifestPathKey{}) != nil
	if v.SizeLane == sizeLaneHpaKeepsSize || v.SizeLane == sizeLaneHpaEnableKeepsSize {
		if err := refuseMetricsAPI(ctx, kubeconfig); err != nil {
			return err
		}
	}
	switch {
	case v.SizeLane == sizeLaneHpaKeepsSize && !secondAct:
		if err := v.awaitSize(ctx, kubeconfig, "2"); err != nil {
			return errors.Wrap(err, "the inline HPA never held the workload at its floor of 2")
		}
		return v.scaleAsAutoscaler(ctx, kubeconfig, 4)
	case v.SizeLane == sizeLaneHpaKeepsSize:
		return v.proveSizeKept(ctx, kubeconfig, "4")
	case v.SizeLane == sizeLaneHpaEnableKeepsSize && !secondAct:
		if err := kubectlRolloutStatus(ctx, kubeconfig, v.Kind+"/"+v.Name, v.Namespace, 3*time.Minute); err != nil {
			return err
		}
		return v.awaitSize(ctx, kubeconfig, "3")
	case v.SizeLane == sizeLaneHpaEnableKeepsSize:
		if err := v.proveSizeKept(ctx, kubeconfig, "3"); err != nil {
			return err
		}
		return v.proveNoResetBelowFloor(ctx, kubeconfig)
	case v.SizeLane == sizeLaneExternalScalerKeepsSize && !secondAct:
		if err := v.awaitSize(ctx, kubeconfig, "1"); err != nil {
			return errors.Wrap(err, "with replicas unset the workload should start with Kubernetes' 1 pod")
		}
		return v.scaleAsAutoscaler(ctx, kubeconfig, 3)
	case v.SizeLane == sizeLaneExternalScalerKeepsSize:
		return v.proveSizeKept(ctx, kubeconfig, "3")
	}
	return errors.Errorf("workload verifier has no size lane %q", v.SizeLane)
}

// awaitSize waits for the workload's declared size to read `want`.
func (v *WorkloadVerifier) awaitSize(ctx context.Context, kubeconfig, want string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		got, err := kubectlGetJSONPath(ctx, kubeconfig, v.Kind, v.Name, v.Namespace, "{.spec.replicas}")
		if err == nil && strings.TrimSpace(got) == want {
			fmt.Printf("  [verify] %s %s/%s holds %s replicas\n", v.Kind, v.Namespace, v.Name, want)
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Errorf("%s %s/%s has %q replicas, want %s (last error: %v)", v.Kind, v.Namespace, v.Name, strings.TrimSpace(got), want, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

// scaleAsAutoscaler writes the size through the /scale subresource under
// kubectl's field manager -- the same write an HPA or KEDA makes.
func (v *WorkloadVerifier) scaleAsAutoscaler(ctx context.Context, kubeconfig string, replicas int) error {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "scale", v.Kind, v.Name,
		"-n", v.Namespace, fmt.Sprintf("--replicas=%d", replicas)).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "scaling %s %s/%s to %d: %s", v.Kind, v.Namespace, v.Name, replicas, firstLines(string(out), 3))
	}
	fmt.Printf("  [verify] act one: scaled %s %s/%s to %d as an autoscaler would\n", v.Kind, v.Namespace, v.Name, replicas)
	return v.awaitSize(ctx, kubeconfig, fmt.Sprint(replicas))
}

// proveSizeKept is act two: the re-apply rolled out (the marker variable is in
// the pod template and the rollout finished) and the size is still `want`.
func (v *WorkloadVerifier) proveSizeKept(ctx context.Context, kubeconfig, want string) error {
	marker, err := kubectlGetJSONPath(ctx, kubeconfig, v.Kind, v.Name, v.Namespace,
		fmt.Sprintf(`{.spec.template.spec.containers[0].env[?(@.name=="%s")].value}`, sizeMarkerEnv))
	if err != nil {
		return err
	}
	if strings.TrimSpace(marker) != sizeMarkerValue {
		return errors.Errorf("act two did not update %s %s/%s: %s is %q, want %q -- the re-apply proved nothing about the size",
			v.Kind, v.Namespace, v.Name, sizeMarkerEnv, strings.TrimSpace(marker), sizeMarkerValue)
	}
	if err := kubectlRolloutStatus(ctx, kubeconfig, v.Kind+"/"+v.Name, v.Namespace, 5*time.Minute); err != nil {
		return err
	}
	got, err := kubectlGetJSONPath(ctx, kubeconfig, v.Kind, v.Name, v.Namespace, "{.spec.replicas}")
	if err != nil {
		return err
	}
	if strings.TrimSpace(got) != want {
		return errors.Errorf("the re-apply moved %s %s/%s to %s replicas; the autoscaler's %s must survive an apply",
			v.Kind, v.Namespace, v.Name, strings.TrimSpace(got), want)
	}
	fmt.Printf("  [verify] act two: the re-apply rolled out and %s %s/%s kept %s replicas\n", v.Kind, v.Namespace, v.Name, want)
	return nil
}

// proveNoResetBelowFloor watches the new HPA for two sync periods: an apply
// that reset the size to 1 makes the HPA rescale "below Spec.MinReplicas".
func (v *WorkloadVerifier) proveNoResetBelowFloor(ctx context.Context, kubeconfig string) error {
	deadline := time.Now().Add(hpaSyncWindow)
	for {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "events", "-n", v.Namespace,
			"--field-selector", "involvedObject.kind=HorizontalPodAutoscaler,reason=SuccessfulRescale",
			"-o", "jsonpath={.items[*].message}").CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "reading the HPA's rescale events: %s", firstLines(string(out), 3))
		}
		if strings.Contains(string(out), hpaBelowFloorReason) {
			return errors.Errorf("turning the HPA on dropped %s %s/%s below its floor: the HPA rescaled %q",
				v.Kind, v.Namespace, v.Name, strings.TrimSpace(string(out)))
		}
		if time.Now().After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	fmt.Printf("  [verify] act two: the HPA never found %s %s/%s below its floor\n", v.Kind, v.Namespace, v.Name)
	return v.awaitSize(ctx, kubeconfig, "3")
}

// refuseMetricsAPI stops an HPA lane on a cluster that serves metrics: there
// the HPA would scale the workload on real load, and a size act one set could
// change for reasons that have nothing to do with the apply.
func refuseMetricsAPI(ctx context.Context, kubeconfig string) error {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "apiservice",
		"v1beta1.metrics.k8s.io", "--ignore-not-found", "-o", "name").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "checking for a metrics API: %s", firstLines(string(out), 3))
	}
	if strings.TrimSpace(string(out)) != "" {
		return errors.New("this cluster serves the metrics API, so its HPAs scale on real load; " +
			"the replica-ownership lanes need a cluster without metrics-server to tell an apply's effect from the HPA's")
	}
	return nil
}
