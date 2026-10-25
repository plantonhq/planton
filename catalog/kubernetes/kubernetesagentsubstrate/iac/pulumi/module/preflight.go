package module

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadowner"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// preflight runs every refusal before anything registers: a namespace
// reference that resolved to another name; a workload or namespace another
// installation holds, or a namespace its owner has not created
// (workloadowner); and a release move upstream's upgrade policy forbids,
// read from the stamps on the API server Deployment and on the bundled
// database's volume (upgrade.go). Nothing is read during a destroy.
// Terraform twin: preflight.tf, the same checks as preconditions on a
// terraform_data resource, in the same words.
func preflight(providerConfig *kubernetesprovider.KubernetesProviderConfig, s *settings) error {
	if s.requestedNamespace != vars.Namespace {
		return namespaceNotUpstreams(s.requestedNamespace)
	}
	if iacinput.IsDestroy() {
		return nil
	}
	namespaces := []workloadowner.Namespace{
		{Name: vars.Namespace, Create: s.createNamespace},
		{Name: vars.PodCertificatesNamespace, Create: true},
	}
	if err := workloadowner.Refuse(providerConfig, kindName, namespaces, ownedWorkloads(s)); err != nil {
		return err
	}
	client, err := workloadowner.Client(providerConfig)
	if err != nil {
		return err
	}
	return refuseUpgrade(client, s)
}

// refuseUpgrade reads the two version stamps.
func refuseUpgrade(client dynamic.Interface, s *settings) error {
	deployments := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	installed, err := readStamp(client, deployments, vars.APIServerName, func(labels, annotations map[string]string) string {
		if labels[failure.WorkloadOwnerLabel] != kindName {
			return ""
		}
		return annotations[vars.VersionAnnotation]
	})
	if err != nil {
		return err
	}
	if rule := upgradeRule(installed, s.version); rule != "" {
		return deploymentUpgradeRefused(installed, s.version, rule)
	}
	if s.database.kind != "bundled" {
		return nil
	}
	claims := schema.GroupVersionResource{Version: "v1", Resource: "persistentvolumeclaims"}
	minor, err := readStamp(client, claims, vars.PostgresVolumeClaim, func(labels, _ map[string]string) string {
		if labels[vars.MinorLabel] == "" {
			return ""
		}
		return "v" + labels[vars.MinorLabel]
	})
	if err != nil {
		return err
	}
	if rule := upgradeRule(minor, s.version); rule != "" {
		return volumeUpgradeRefused(minor, s.version, rule)
	}
	return nil
}

// readStamp reads one object in ate-system and extracts its stamp; an
// absent object has none.
func readStamp(client dynamic.Interface, resource schema.GroupVersionResource, name string, extract func(labels, annotations map[string]string) string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	object, err := client.Resource(resource).Namespace(vars.Namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrapf(err, "failed to read %s %s/%s for its installed-version stamp", resource.Resource, vars.Namespace, name)
	}
	return extract(object.GetLabels(), object.GetAnnotations()), nil
}

// ownedWorkloads are the workloads the module applies under upstream's
// fixed names, which no other installation may already hold.
func ownedWorkloads(s *settings) []workloadowner.Workload {
	deployment := func(namespace, name string) workloadowner.Workload {
		return workloadowner.Workload{Kind: workloadowner.Deployment, Namespace: namespace, Name: name}
	}
	workloads := []workloadowner.Workload{
		deployment(vars.PodCertificatesNamespace, vars.PodCertificatesName),
		deployment(vars.Namespace, vars.APIServerName),
		deployment(vars.Namespace, vars.IdentityServerName),
		deployment(vars.Namespace, vars.ControllerName),
		{Kind: workloadowner.DaemonSet, Namespace: vars.Namespace, Name: vars.NodeAgentName},
		deployment(vars.Namespace, vars.RouterName),
		deployment(vars.Namespace, vars.EgressName),
	}
	if s.egress.provider == "kubernetes" {
		workloads = append(workloads, deployment(vars.Namespace, vars.CredentialProviderName))
	}
	if s.database.kind == "bundled" {
		workloads = append(workloads, workloadowner.Workload{Kind: workloadowner.StatefulSet, Namespace: vars.Namespace, Name: vars.PostgresName})
	}
	return workloads
}

// namespaceNotUpstreams refuses a spec.namespace reference that resolved to
// another name; the schema refuses another literal, but cannot see what a
// reference resolves to.
func namespaceNotUpstreams(requested string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("spec.namespace resolved to %q", requested),
		Meaning:  fmt.Sprintf("Agent Substrate runs only in %s: upstream's manifests, the API's server name and its token audience name it", vars.Namespace),
		NextStep: fmt.Sprintf("reference a KubernetesNamespace named %s, or set spec.namespace to %s", vars.Namespace, vars.Namespace),
	}
}
