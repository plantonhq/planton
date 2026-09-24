//go:build e2e

package e2e

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/plantonhq/planton/operator/test/utils"
)

// The manager's lifecycle on the Kind cluster, shared by every container in
// this suite: each Ordered container deploys the manager in its BeforeAll
// and removes it in its AfterAll, so a label-filtered run of one container
// stands on its own and two containers in one run never see each other's
// manager. The image was built and loaded once, in the suite's BeforeSuite.

// deployManager creates the manager's namespace under the restricted pod
// security standard, installs the working tree's CRDs, and deploys the
// working tree's manager image through the kubebuilder path.
func deployManager() {
	By("creating manager namespace")
	cmd := exec.Command("kubectl", "create", "ns", namespace)
	_, err := utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

	By("labeling the namespace to enforce the restricted security policy")
	cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
		"pod-security.kubernetes.io/enforce=restricted")
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

	By("installing CRDs")
	cmd = exec.Command("make", "install")
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

	By("deploying the controller-manager")
	cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", managerImage))
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
}

// undeployManager removes the manager, the CRDs, and the manager's
// namespace; every step is best-effort so a failed spec never blocks the
// next container's clean start.
func undeployManager() {
	By("undeploying the controller-manager")
	cmd := exec.Command("make", "undeploy")
	_, _ = utils.Run(cmd)

	By("uninstalling CRDs")
	cmd = exec.Command("make", "uninstall")
	_, _ = utils.Run(cmd)

	By("removing manager namespace")
	cmd = exec.Command("kubectl", "delete", "ns", namespace)
	_, _ = utils.Run(cmd)
}

// managerPodName is the one live manager pod, by the label the kubebuilder
// Deployment carries; it fails the spec when there is not exactly one.
func managerPodName(g Gomega) string {
	cmd := exec.Command("kubectl", "get",
		"pods", "-l", "control-plane=controller-manager",
		"-o", "go-template={{ range .items }}"+
			"{{ if not .metadata.deletionTimestamp }}"+
			"{{ .metadata.name }}"+
			"{{ \"\\n\" }}{{ end }}{{ end }}",
		"-n", namespace,
	)
	podOutput, err := utils.Run(cmd)
	g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
	podNames := utils.GetNonEmptyLines(podOutput)
	g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
	g.Expect(podNames[0]).To(ContainSubstring("controller-manager"))
	return podNames[0]
}

// managerLogs returns the manager's log so far; a spec asserts on the lines
// the operator writes when it does (or must never do) something.
func managerLogs() string {
	pod := managerPodName(Default)
	out, err := utils.Run(exec.Command("kubectl", "logs", pod, "-n", namespace))
	Expect(err).NotTo(HaveOccurred(), "Failed to read the manager's log")
	return out
}

// dumpManagerOnFailure writes the manager's log, the namespace's events, and
// the manager pod's description to the report when the current spec failed,
// so a red lane explains itself without a second run.
func dumpManagerOnFailure(controllerPodName string) {
	specReport := CurrentSpecReport()
	if !specReport.Failed() {
		return
	}
	By("Fetching controller manager pod logs")
	cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
	controllerLogs, err := utils.Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
	}

	By("Fetching Kubernetes events")
	cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
	eventsOutput, err := utils.Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
	}

	By("Fetching controller manager pod description")
	cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
	podDescription, err := utils.Run(cmd)
	if err == nil {
		fmt.Println("Pod description:\n", podDescription)
	} else {
		fmt.Println("Failed to describe controller pod")
	}
}
