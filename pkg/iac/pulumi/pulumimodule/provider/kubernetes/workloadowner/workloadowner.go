// Package workloadowner refuses to install a kind's workloads over ones
// another installation created.
//
// A kind that installs an upstream's own release under the upstream's fixed
// names (one install per cluster is the upstream contract) cannot share those
// names with a second installation of the same software: a managed add-on
// that installs it, a `kubectl apply` of the upstream release, another tool.
// Server-side apply would adopt the existing workload silently, two owners
// would then fight over one controller, and destroying the kind later would
// delete workloads it never created.
//
// So before the module registers anything, it reads each workload it is about
// to apply (a Deployment, DaemonSet or StatefulSet). One that exists without
// the kind's ownership label (failure.WorkloadOwnerLabel, stamped by the
// module on every workload it applies) stops the deploy with the owner named.
// A workload the kind created passes, so a re-run, an upgrade and a reinstall
// are unaffected.
//
// The namespaces those workloads live in get the same care, because each
// one's owner is a choice: when the kind creates a namespace, one that exists
// without the kind's label belongs to someone else and taking it over would
// make a later destroy delete their objects; when another component owns it,
// it must exist, and it must not be one the kind created, which only turning
// create_namespace off on a live installation produces and which would delete
// the namespace.
//
// The Terraform twin reads the same objects with kubernetes_resources data
// sources and refuses in preconditions, in the same words.
package workloadowner

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// WorkloadKind is the apps/v1 kind of a workload the module applies.
type WorkloadKind string

const (
	Deployment  WorkloadKind = "Deployment"
	DaemonSet   WorkloadKind = "DaemonSet"
	StatefulSet WorkloadKind = "StatefulSet"
)

// Workload names one workload the module is about to apply.
type Workload struct {
	Kind      WorkloadKind
	Namespace string
	Name      string
}

// Namespace names a namespace the workloads live in, and whether the module
// creates it (true) or another component owns it (false).
type Namespace struct {
	Name   string
	Create bool
}

// readTimeout bounds each read; a cluster that cannot answer a get in this
// time will not take an apply either.
const readTimeout = 30 * time.Second

var (
	workloadResources = map[WorkloadKind]schema.GroupVersionResource{
		Deployment:  {Group: "apps", Version: "v1", Resource: "deployments"},
		DaemonSet:   {Group: "apps", Version: "v1", Resource: "daemonsets"},
		StatefulSet: {Group: "apps", Version: "v1", Resource: "statefulsets"},
	}
	namespaceGVR = schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
)

// Refuse reads each workload, then each namespace, and returns the explained
// refusal for the first one another installation holds (or for a namespace
// its owner has not created). kind is the catalog kind's name, the ownership
// label's value. Workloads are read first because the other installation's
// controller names its owner best. Nothing is read during a destroy: whatever
// runs on the cluster then, the resources in state are deleted, and a refusal
// must never stand between a user and deleting their stack.
func Refuse(providerConfig *kubernetesprovider.KubernetesProviderConfig, kind string, namespaces []Namespace, workloads []Workload) error {
	if iacinput.IsDestroy() {
		return nil
	}
	client, err := Client(providerConfig)
	if err != nil {
		return err
	}
	return refuse(client, kind, namespaces, workloads)
}

// Client is the dynamic client Refuse reads with, for a module that reads
// more of the cluster before registering anything (an installed-version
// stamp, say) and wants one connection for all of it.
func Client(providerConfig *kubernetesprovider.KubernetesProviderConfig) (dynamic.Interface, error) {
	restConfig, err := kubeconfig.RESTConfig(providerConfig, os.Getenv(execcredential.CommandPathEnvVar))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve the cluster connection for the workload ownership check")
	}
	client, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the cluster client for the workload ownership check")
	}
	return client, nil
}

func refuse(client dynamic.Interface, kind string, namespaces []Namespace, workloads []Workload) error {
	for _, workload := range workloads {
		resource, known := workloadResources[workload.Kind]
		if !known {
			return errors.Errorf("workload %s/%s has kind %q; the ownership check reads Deployments, DaemonSets and StatefulSets", workload.Namespace, workload.Name, workload.Kind)
		}
		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		object, err := client.Resource(resource).Namespace(workload.Namespace).Get(ctx, workload.Name, metav1.GetOptions{})
		cancel()
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "failed to read %s %s/%s before applying it", workload.Kind, workload.Namespace, workload.Name)
		}
		if object.GetLabels()[failure.WorkloadOwnerLabel] == kind {
			continue
		}
		return failure.KubernetesWorkloadOwnedElsewhere(kind, string(workload.Kind), workload.Namespace, workload.Name, ownerDescription(object))
	}
	for _, namespace := range namespaces {
		if err := refuseNamespace(client, kind, namespace); err != nil {
			return err
		}
	}
	return nil
}

func refuseNamespace(client dynamic.Interface, kind string, namespace Namespace) error {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	object, err := client.Resource(namespaceGVR).Get(ctx, namespace.Name, metav1.GetOptions{})
	cancel()
	switch {
	case apierrors.IsNotFound(err):
		if !namespace.Create {
			return failure.KubernetesNamespaceMissing(kind, namespace.Name)
		}
		return nil
	case err != nil:
		return errors.Wrapf(err, "failed to read namespace %s before installing into it", namespace.Name)
	case namespace.Create && object.GetLabels()[failure.WorkloadOwnerLabel] != kind:
		return failure.KubernetesNamespaceOwnedElsewhere(kind, namespace.Name, ownerDescription(object))
	case !namespace.Create && object.GetLabels()[failure.WorkloadOwnerLabel] == kind:
		return failure.KubernetesNamespaceCreatedHere(kind, namespace.Name)
	}
	return nil
}

// ownerDescription names who wrote an object from the API server's own
// record (the field managers), in the words the Terraform twin composes.
func ownerDescription(object *unstructured.Unstructured) string {
	managers := make([]string, 0, len(object.GetManagedFields()))
	for _, entry := range object.GetManagedFields() {
		managers = append(managers, entry.Manager)
	}
	if len(managers) == 0 {
		return "it carries no ownership marks"
	}
	return "last written by field manager(s) " + strings.Join(managers, ", ")
}
