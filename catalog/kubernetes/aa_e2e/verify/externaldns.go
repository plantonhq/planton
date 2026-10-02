package verify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ExternalDnsInstallVerifier checks an ExternalDNS installation: the
// controller Deployment is Available. The module pins the chart's fullname
// to metadata.name, so the Deployment name is deterministic per instance
// (multiple installations per cluster are first-class for this component).
//
// Deliberately NOT asserted here: record synchronization. The controller
// validates provider credentials at its first sync, not at startup, and
// real DNS writes need a cloud account — record-level assertions ride the
// batched real-cluster lanes. An Available controller with the install's
// full env/volume/args wiring is the kind-cluster contract.
//
// With ListenerSets (the spec's gateway_listener_sets), the install must also
// carry what that switch renders: the --gateway-listener-sets flag on the
// controller and the ClusterRole's permission to read ListenerSets.
type ExternalDnsInstallVerifier struct {
	Namespace     string
	ComponentName string
	ListenerSets  bool
}

func (v *ExternalDnsInstallVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] external-dns installation %q in namespace %q\n", v.ComponentName, v.Namespace)

	if err := KubectlResourceExists(ctx, kubeconfig, "namespace", v.Namespace, ""); err != nil {
		return errors.Wrapf(err, "namespace %q not found for external-dns %q", v.Namespace, v.ComponentName)
	}

	if err := kubectlWait(ctx, kubeconfig, "deployment", v.ComponentName, v.Namespace,
		"condition=Available", 2*time.Minute); err != nil {
		return errors.Wrapf(err, "deployment %q not available in namespace %q", v.ComponentName, v.Namespace)
	}

	if v.ListenerSets {
		args, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", v.ComponentName, v.Namespace,
			"{.spec.template.spec.containers[0].args}")
		if err != nil {
			return errors.Wrapf(err, "reading the controller's args")
		}
		if !strings.Contains(args, "--gateway-listener-sets") {
			return errors.Errorf("the controller runs without --gateway-listener-sets: %s", args)
		}
		resources, err := kubectlGetJSONPath(ctx, kubeconfig, "clusterrole", v.ComponentName, "",
			"{.rules[*].resources}")
		if err != nil {
			return errors.Wrapf(err, "reading the controller's ClusterRole")
		}
		if !strings.Contains(resources, "listenersets") {
			return errors.Errorf("the ClusterRole grants no listenersets access: %s", resources)
		}
		fmt.Printf("  [verify] LISTENER SETS: --gateway-listener-sets on the controller, listenersets readable\n")
	}

	return nil
}

func (v *ExternalDnsInstallVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	// Multiple instances can share a namespace, so the release's own
	// Deployment (not the namespace) is the absence signal.
	return KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.ComponentName, v.Namespace)
}
