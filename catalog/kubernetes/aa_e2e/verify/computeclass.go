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

// ComputeClassVerifier verifies a GKE ComputeClass and, in behavioral mode,
// proves the class FALLS BACK: a verifier-owned driver pod that selects the
// class must land on a node GKE created from the scenario's second priority
// rule, because the first rule (a machine no default project quota can hold)
// fails node creation the way a stockout does -- the same path, made
// deterministic. After the driver is deleted, the class's consolidation delay
// (one minute in the scenario) must reclaim the node, so the lane leaves no
// billed capacity behind.
type ComputeClassVerifier struct {
	Name string
	// Behavioral switches on the live fallback proof (see above).
	Behavioral bool
}

// computeClass* identify the verifier-owned driver workload and the family
// the scenario's fallback rule names -- fixed so the behavioral scenario and
// this proof cannot drift apart silently.
const (
	computeClassResource        = "computeclasses.cloud.google.com"
	computeClassDriverNamespace = "e2e-computeclass-drive"
	computeClassDriverPod       = "e2e-computeclass-driver"
	computeClassSelectorKey     = "cloud.google.com/compute-class"
	computeClassFamilyLabel     = "cloud.google.com/machine-family"
	computeClassFallbackFamily  = "e2"
)

func (v *ComputeClassVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] GKE ComputeClass %q\n", v.Name)

	if err := KubectlResourceExists(ctx, kubeconfig, computeClassResource, v.Name, ""); err != nil {
		return err
	}
	// GKE validates the class against the cluster (machine families, disk
	// types, reservations' shapes) and reports the verdict as the Health
	// condition -- the object-grade contract that GKE accepted the class.
	if err := kubectlWait(ctx, kubeconfig, computeClassResource, v.Name, "",
		"condition=Health", 3*time.Minute); err != nil {
		return errors.Wrap(err, "ComputeClass never reported Health=True (GKE refused its configuration; read `kubectl describe computeclass`)")
	}

	if !v.Behavioral {
		return nil
	}
	return v.proveFallback(ctx, kubeconfig)
}

func (v *ComputeClassVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, computeClassResource, v.Name, "")
}

// proveFallback runs the provision-and-reclaim loop. Every resource is
// verifier-owned and removed in defers so a failed assertion cannot strand a
// pod that would pin the created node past the lane.
func (v *ComputeClassVerifier) proveFallback(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] behavioral fallback: a pod selecting the class must land on the second rule's %s family\n", computeClassFallbackFamily)

	if err := v.kubectl(ctx, kubeconfig, "create", "namespace", computeClassDriverNamespace); err != nil {
		return errors.Wrap(err, "failed to create driver namespace")
	}
	defer func() {
		_ = v.kubectl(context.Background(), kubeconfig, "delete", "namespace", computeClassDriverNamespace,
			"--ignore-not-found", "--wait=false")
	}()

	// The driver: selects the class (GKE adds the class taint's toleration;
	// it is spelled out so the proof does not lean on that admission step),
	// sized well under the smallest machine the fallback rule allows.
	driver := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  nodeSelector:
    %s: %s
  tolerations:
    - key: %s
      operator: Equal
      value: %s
      effect: NoSchedule
  containers:
    - name: driver
      image: registry.k8s.io/pause:3.10
      resources:
        requests:
          cpu: 100m
          memory: 64Mi
`, computeClassDriverPod, computeClassDriverNamespace,
		computeClassSelectorKey, v.Name, computeClassSelectorKey, v.Name)
	driverFile, err := writeTempManifest(driver)
	if err != nil {
		return err
	}
	defer os.Remove(driverFile)
	if err := v.kubectl(ctx, kubeconfig, "apply", "-f", driverFile); err != nil {
		return errors.Wrap(err, "failed to apply driver pod")
	}
	defer func() {
		_ = v.kubectl(context.Background(), kubeconfig, "delete", "pod", computeClassDriverPod,
			"-n", computeClassDriverNamespace, "--ignore-not-found", "--wait=false")
	}()

	// Node pool auto-creation plus node boot lands in 3-6 minutes; 15 covers
	// a slow pool creation in a busy zone.
	if err := kubectlWait(ctx, kubeconfig, "pod", computeClassDriverPod, computeClassDriverNamespace,
		"condition=Ready", 15*time.Minute); err != nil {
		return errors.Wrap(err, "the driver pod never became Ready (GKE created no node for the class)")
	}
	node, err := v.output(ctx, kubeconfig, "get", "pod", computeClassDriverPod, "-n", computeClassDriverNamespace,
		"-o", "jsonpath={.spec.nodeName}")
	if err != nil {
		return err
	}
	family, err := v.output(ctx, kubeconfig, "get", "node", node,
		"-o", "jsonpath={.metadata.labels."+strings.ReplaceAll(computeClassFamilyLabel, ".", "\\.")+"}")
	if err != nil {
		return err
	}
	class, err := v.output(ctx, kubeconfig, "get", "node", node,
		"-o", "jsonpath={.metadata.labels."+strings.ReplaceAll(computeClassSelectorKey, ".", "\\.")+"}")
	if err != nil {
		return err
	}
	if class != v.Name {
		return errors.Errorf("driver landed on node %s of compute class %q, want %q", node, class, v.Name)
	}
	if family != computeClassFallbackFamily {
		return errors.Errorf("driver landed on a %q node (%s), want the fallback rule's %q family -- the class did not fall back as ordered", family, node, computeClassFallbackFamily)
	}
	fmt.Printf("  [verify] driver scheduled on %s (%s family, class %s) -- the unsatisfiable first rule fell back to the second\n", node, family, class)

	// The reclaim half: with the driver gone the node is empty, and the
	// class's one-minute consolidation delay must remove it.
	if err := v.kubectl(ctx, kubeconfig, "delete", "pod", computeClassDriverPod,
		"-n", computeClassDriverNamespace, "--wait=true", "--timeout=2m"); err != nil {
		return errors.Wrap(err, "failed to delete driver pod")
	}
	if err := v.waitForClassNodes(ctx, kubeconfig, 0, 20*time.Minute); err != nil {
		return errors.Wrap(err, "the empty class node was never reclaimed")
	}
	fmt.Printf("  [verify] empty class node reclaimed -- the full fallback/reclaim loop is proven\n")
	return nil
}

// waitForClassNodes polls until exactly want nodes carry the class label
// (GKE labels every node it creates for a class with the class name).
func (v *ComputeClassVerifier) waitForClassNodes(ctx context.Context, kubeconfig string, want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		out, err := v.output(ctx, kubeconfig, "get", "nodes", "-l", computeClassSelectorKey+"="+v.Name,
			"-o", "jsonpath={.items[*].metadata.name}")
		names := strings.Fields(out)
		if err == nil && len(names) == want {
			return nil
		}
		last = fmt.Sprintf("nodes=%v err=%v", names, err)
		time.Sleep(15 * time.Second)
	}
	return errors.Errorf("class node count never reached %d (last: %s)", want, last)
}

func (v *ComputeClassVerifier) kubectl(ctx context.Context, kubeconfig string, args ...string) error {
	_, err := v.output(ctx, kubeconfig, args...)
	return err
}

func (v *ComputeClassVerifier) output(ctx context.Context, kubeconfig string, args ...string) (string, error) {
	full := append([]string{"--kubeconfig", kubeconfig}, args...)
	out, err := exec.CommandContext(ctx, "kubectl", full...).CombinedOutput()
	if err != nil {
		return "", errors.Errorf("kubectl %s: %v: %s", strings.Join(args, " "), err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
