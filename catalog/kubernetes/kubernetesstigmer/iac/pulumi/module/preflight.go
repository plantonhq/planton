package module

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadowner"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// preflight reads the cluster before anything registers and refuses, in a
// sentence, what Stigmer itself would only refuse later (or never say):
//
//   - the operator's namespace created here and also the Stigmer's, where a
//     destroy would delete the encryption keys;
//   - a namespace that is missing, or the operator's namespace owned by
//     someone else when this resource is asked to create it;
//   - a Stigmer definition that another Helm release owns. Stigmer ships its
//     operator only as this chart, which always carries the definition, so
//     every other install of the operator is a release that owns it; and
//     this install moved to another namespace in place shows up the same way;
//   - a Stigmer in the namespace that this resource did not declare (one in
//     another namespace is refused by Stigmer's own operator, in its status,
//     because the operator serves only the oldest Stigmer on a cluster);
//   - a version below the one the server runs;
//   - a CloudNativePG or backup-plugin definition the install needs that a
//     Helm release put on the cluster and nothing runs any more (its release
//     has no Deployment left). Stigmer's operator trusts a prerequisite
//     someone else installed and installs nothing over it, so it would wait
//     forever on a controller that is gone; the catalog's own CloudNativePG
//     kinds keep their definitions when destroyed, so this is what a cluster
//     looks like after one was removed. A definition installed without Helm
//     names no controller to look for, and is trusted as Stigmer trusts it.
//
// Nothing is read during a destroy: whatever runs on the cluster then, a
// refusal must never stand between a user and deleting their stack.
// Terraform twin: preflight.tf, the same checks as preconditions, in the
// same words.
func preflight(providerConfig *kubernetesprovider.KubernetesProviderConfig, locals *Locals) error {
	if iacinput.IsDestroy() {
		return nil
	}
	if locals.Spec.GetOperator().GetCreateNamespace() && locals.OperatorNamespace == locals.Namespace {
		return operatorNamespaceHoldsKeys(locals.Namespace)
	}
	client, err := workloadowner.Client(providerConfig)
	if err != nil {
		return err
	}
	if err := refuseNamespaces(client, locals); err != nil {
		return err
	}
	if err := refuseForeignDefinition(client, locals); err != nil {
		return err
	}
	if err := refuseStigmers(client, locals); err != nil {
		return err
	}
	return refuseOrphanedPrerequisites(client, locals)
}

var (
	namespaces      = schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
	definitions     = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	stigmerResource = schema.GroupVersionResource{Group: "stigmer.ai", Version: "v1", Resource: "stigmers"}
	deployments     = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

// A prerequisite Stigmer's operator installs when it is missing, by the
// definition it detects it by (deploy/operator/internal/component in
// stigmer/stigmer), and what waits on its controller.
type prerequisite struct {
	name       string // as a person calls it
	field      string // its spec.prerequisites field
	definition string
	group      string // its definitions' API group
	waits      string // what never happens without its controller
}

// neededPrerequisites are the prerequisites this install uses in auto mode:
// CloudNativePG for the bundled database, and its backup plugin when the
// database backs up or restores.
func neededPrerequisites(locals *Locals) []prerequisite {
	spec := locals.Spec
	if spec.GetDatabase().GetExternal() != nil {
		return nil
	}
	var needed []prerequisite
	if spec.GetPrerequisites().GetPostgresOperator() != "skip" {
		needed = append(needed, prerequisite{
			name: "CloudNativePG", field: "postgres_operator", definition: vars.CloudNativePgDefinition,
			group: "postgresql.cnpg.io", waits: "the bundled database would wait forever for a controller that is gone",
		})
	}
	bundled := spec.GetDatabase().GetBundled()
	if (bundled.GetBackup() != nil || bundled.GetRecoverFrom() != nil) && spec.GetPrerequisites().GetPostgresBackupPlugin() != "skip" {
		needed = append(needed, prerequisite{
			name: "CloudNativePG's backup plugin", field: "postgres_backup_plugin", definition: vars.BackupPluginDefinition,
			group: "barmancloud.cnpg.io", waits: "the database's backups and restores would wait forever for a controller that is gone",
		})
	}
	return needed
}

const readTimeout = 30 * time.Second

func refuseNamespaces(client dynamic.Interface, locals *Locals) error {
	stigmerNamespace, err := get(client, namespaces, "", locals.Namespace)
	if err != nil {
		return err
	}
	if stigmerNamespace == nil {
		return stigmerNamespaceMissing(locals.Namespace)
	}
	if locals.OperatorNamespace == locals.Namespace {
		return nil
	}
	operatorNamespace, err := get(client, namespaces, "", locals.OperatorNamespace)
	if err != nil {
		return err
	}
	ours := operatorNamespace != nil && operatorNamespace.GetLabels()[failure.WorkloadOwnerLabel] == kindName
	switch create := locals.Spec.GetOperator().GetCreateNamespace(); {
	case !create && operatorNamespace == nil:
		return operatorNamespaceMissing(locals.OperatorNamespace)
	case create && operatorNamespace != nil && !ours:
		return operatorNamespaceOwnedElsewhere(locals.OperatorNamespace)
	case !create && ours:
		return operatorNamespaceCreatedHere(locals.OperatorNamespace)
	}
	return nil
}

func refuseForeignDefinition(client dynamic.Interface, locals *Locals) error {
	definition, err := get(client, definitions, "", vars.StigmerDefinition)
	if err != nil || definition == nil {
		return err
	}
	release := definition.GetAnnotations()["meta.helm.sh/release-name"]
	namespace := definition.GetAnnotations()["meta.helm.sh/release-namespace"]
	if release == vars.OperatorRelease && namespace == locals.OperatorNamespace {
		return nil
	}
	return definitionOwnedElsewhere(release, namespace, locals.OperatorNamespace)
}

func refuseStigmers(client dynamic.Interface, locals *Locals) error {
	definition, err := get(client, definitions, "", vars.StigmerDefinition)
	if err != nil || definition == nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	list, err := client.Resource(stigmerResource).Namespace(locals.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to list the Stigmers in %s before declaring one", locals.Namespace)
	}
	for i := range list.Items {
		stigmer := &list.Items[i]
		ours := stigmer.GetNamespace() == locals.Namespace && stigmer.GetName() == vars.StigmerName &&
			stigmer.GetLabels()[failure.WorkloadOwnerLabel] == kindName
		if !ours {
			return stigmerDeclaredElsewhere(stigmer.GetNamespace(), stigmer.GetName())
		}
		running, _, _ := unstructured.NestedString(stigmer.Object, "status", "version")
		if versionBelow(locals.Spec.GetVersion(), running) {
			return downgradeRefused(locals.Namespace, running, locals.Spec.GetVersion())
		}
	}
	return nil
}

func refuseOrphanedPrerequisites(client dynamic.Interface, locals *Locals) error {
	for _, p := range neededPrerequisites(locals) {
		definition, err := get(client, definitions, "", p.definition)
		if err != nil {
			return err
		}
		if definition == nil || definition.GetLabels()[vars.StigmerManagedByLabel] == vars.StigmerManagedByValue {
			continue
		}
		release := definition.GetAnnotations()["meta.helm.sh/release-name"]
		namespace := definition.GetAnnotations()["meta.helm.sh/release-namespace"]
		if release == "" || namespace == "" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		controllers, err := client.Resource(deployments).Namespace(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/instance=" + release,
		})
		cancel()
		if err != nil {
			return errors.Wrapf(err, "failed to read the Deployments of the Helm release %s in %s that installed %s", release, namespace, p.definition)
		}
		if len(controllers.Items) == 0 {
			return prerequisiteOrphaned(p, release, namespace)
		}
	}
	return nil
}

// get reads one object; an absent one is nil.
func get(client dynamic.Interface, resource schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	object, err := client.Resource(resource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s %s before installing Stigmer", resource.Resource, name)
	}
	return object, nil
}

// versionBelow reports whether requested is a lower release than running
// (the server's status.version, spelled with a leading "v"). An empty or
// unreadable running version is never below anything.
func versionBelow(requested, running string) bool {
	r := semver(requested)
	v := semver(strings.TrimPrefix(running, "v"))
	if r == nil || v == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if r[i] != v[i] {
			return r[i] < v[i]
		}
	}
	return false
}

func semver(s string) []int {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil
	}
	out := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		out[i] = n
	}
	return out
}

// The refusals, each in the words the Terraform preconditions use.

func operatorNamespaceHoldsKeys(namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("spec.operator.namespace and spec.namespace both resolve to %s, and spec.operator.create_namespace is true", namespace),
		Meaning:  "this resource would create the namespace and delete it on destroy, and the Stigmer's encryption keys live there: deleting it would make every backup unreadable",
		NextStep: "install the operator in a namespace of its own (stigmer-system by convention), or turn spec.operator.create_namespace off",
	}
}

func stigmerNamespaceMissing(namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("namespace %s (spec.namespace) does not exist on the cluster", namespace),
		Meaning:  "this resource never creates the Stigmer's namespace, because deleting it would delete the encryption keys and make every backup unreadable",
		NextStep: fmt.Sprintf("deploy a KubernetesNamespace named %s first and reference it from spec.namespace, or create it yourself", namespace),
	}
}

func operatorNamespaceMissing(namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("namespace %s (spec.operator.namespace) does not exist on the cluster, and spec.operator.create_namespace is false", namespace),
		Meaning:  "the operator installs into that namespace, so its owner must create it first",
		NextStep: fmt.Sprintf("deploy a KubernetesNamespace named %s first and reference it from spec.operator.namespace, or set spec.operator.create_namespace to true", namespace),
	}
}

func operatorNamespaceOwnedElsewhere(namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("namespace %s already exists and was not created by this KubernetesStigmer, and spec.operator.create_namespace is true", namespace),
		Meaning:  "this resource would take the namespace over, and destroying it later would delete the namespace and everything other components put in it",
		NextStep: fmt.Sprintf("set spec.operator.create_namespace to false to install into the existing namespace; or, if you created it for this install, mark it as this kind's (`kubectl label namespace %s %s=%s`) and deploy again", namespace, failure.WorkloadOwnerLabel, kindName),
	}
}

func operatorNamespaceCreatedHere(namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("namespace %s was created by this KubernetesStigmer, and spec.operator.create_namespace is now false", namespace),
		Meaning:  "turning it off would delete the namespace, and the operator with it",
		NextStep: "keep spec.operator.create_namespace true; to hand the namespace to a KubernetesNamespace, destroy this resource first, then deploy it again with create_namespace false",
	}
}

func definitionOwnedElsewhere(release, namespace, operatorNamespace string) error {
	owner := fmt.Sprintf("the Helm release %s in namespace %s", release, namespace)
	if release == "" {
		owner = "no Helm release (it was applied another way)"
	}
	next := "remove the other Stigmer operator install first, or keep it and do not install KubernetesStigmer on this cluster"
	if release == vars.OperatorRelease {
		next = fmt.Sprintf("the operator cannot move to another namespace in place: destroy this resource, delete the kept definition once no Stigmer is left (kubectl delete crd %s), then deploy again with spec.operator.namespace %s; or set spec.operator.namespace back to %s", vars.StigmerDefinition, operatorNamespace, namespace)
	}
	return &failure.Failure{
		Observed: fmt.Sprintf("the Stigmer definition %s on the cluster belongs to %s, and this resource installs the operator as release %s in namespace %s", vars.StigmerDefinition, owner, vars.OperatorRelease, operatorNamespace),
		Meaning:  "Helm refuses to take over a definition another release owns, and Stigmer runs one operator per cluster",
		NextStep: next,
	}
}

func stigmerDeclaredElsewhere(namespace, name string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("Stigmer %s/%s is declared on the cluster, and this resource did not declare it", namespace, name),
		Meaning:  "Stigmer runs one install per cluster, and this resource declares the namespace's Stigmer as stigmer: the operator serves only the oldest Stigmer and refuses every other one",
		NextStep: fmt.Sprintf("delete the other Stigmer first (kubectl -n %s delete stigmer %s; save its keys Secret before you delete its namespace), or keep it and do not install KubernetesStigmer on this cluster", namespace, name),
	}
}

func downgradeRefused(namespace, running, requested string) error {
	return failure.KubernetesUpgradeRefused(kindName, "Stigmer", fmt.Sprintf("Stigmer %s/%s", namespace, vars.StigmerName),
		running, requested,
		"its server refuses to start on a database a newer release migrated",
		fmt.Sprintf("set spec.version back to %s or later; to go back to %s, restore a backup taken under it into a new install instead", strings.TrimPrefix(running, "v"), requested))
}

func prerequisiteOrphaned(p prerequisite, release, namespace string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("%s's definition %s is on the cluster, installed by the Helm release %s in namespace %s, and no controller runs it: that release has no Deployment left", p.name, p.definition, release, namespace),
		Meaning:  fmt.Sprintf("Stigmer's operator trusts a %s someone else installed and installs nothing over it, so %s", p.name, p.waits),
		NextStep: fmt.Sprintf("install %s again; or remove what is left and deploy again: deleting a definition deletes every object of its type, so first check none is left (kubectl get %s -A), then delete its definitions (kubectl get crd -o name | grep '\\.%s$' | xargs kubectl delete); to run %s yourself instead, set spec.prerequisites.%s to skip", p.name, p.definition, strings.ReplaceAll(p.group, ".", "\\."), p.name, p.field),
	}
}
