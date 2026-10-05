package module

import (
	"strconv"

	kubernetespersistentvolumeclaimv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespersistentvolumeclaim/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// A claim under a WaitForFirstConsumer StorageClass is correctly Pending
// until a pod uses it — so the module must never await Bound. This
// annotation is Pulumi's opt-out of its PVC readiness await; the Terraform
// module's wait_until_bound=false is the same decision on the other engine.
const skipAwaitAnnotation = "pulumi.com/skipAwait"

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetespersistentvolumeclaimv1alpha1.KubernetesPersistentVolumeClaimSpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// Access modes with the Kubernetes-default fallback (ReadWriteOnce)
	// applied module-side, so both engines submit identical claims when the
	// spec omits the field. (The API itself REQUIRES accessModes — there is
	// no server default to defer to.)
	AccessModes []string

	// The API string for volume_mode, resolved with the server default
	// (Filesystem).
	VolumeMode string

	// The storageClassName wire value: the FK-resolved class name, "" when
	// dynamic provisioning is explicitly disabled, or nil (unset — cluster
	// default applies) when the spec names nothing. The empty-vs-absent
	// distinction is load-bearing upstream.
	StorageClassName *string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetespersistentvolumeclaimv1alpha1.KubernetesPersistentVolumeClaimIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesPersistentVolumeClaim.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}
	for k, v := range spec.GetLabels() {
		if _, isIdentityKey := labels[k]; !isIdentityKey {
			labels[k] = v
		}
	}

	// The skip-await opt-out is module-managed and set AFTER the user map so
	// it can never be overridden: re-enabling the bind-await through an
	// annotation would deadlock WaitForFirstConsumer claims on this engine
	// only (the Terraform module hardcodes wait_until_bound=false with no
	// override), which is exactly the cross-engine divergence the modules
	// exist to prevent.
	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}
	annotations[skipAwaitAnnotation] = "true"

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When omitted entirely, fall back to the cluster's
	// "default" namespace — the same behavior as kubectl without a namespace flag.
	namespace := spec.GetNamespace().GetValue()
	if namespace == "" {
		namespace = "default"
	}

	accessModes := spec.GetAccessModes()
	if len(accessModes) == 0 {
		accessModes = []string{"ReadWriteOnce"}
	}

	volumeMode := "Filesystem"
	if spec.GetVolumeMode() == kubernetespersistentvolumeclaimv1alpha1.KubernetesPersistentVolumeClaimSpec_block {
		volumeMode = "Block"
	}

	return &Locals{
		Context:          ctx,
		Spec:             spec,
		Namespace:        namespace,
		Name:             spec.GetName(),
		Labels:           labels,
		Annotations:      annotations,
		AccessModes:      accessModes,
		VolumeMode:       volumeMode,
		StorageClassName: resolveStorageClassName(spec),
	}
}

// resolveStorageClassName computes the storageClassName wire value. Kubernetes
// distinguishes an EMPTY class name (bind only pre-provisioned volumes) from
// an ABSENT one (the cluster's default class applies) — which is exactly why
// the spec carries disable_dynamic_provisioning as its own field.
func resolveStorageClassName(spec *kubernetespersistentvolumeclaimv1alpha1.KubernetesPersistentVolumeClaimSpec) *string {
	if spec.GetDisableDynamicProvisioning() {
		empty := ""
		return &empty
	}
	if name := spec.GetStorageClassName().GetValue(); name != "" {
		return &name
	}
	return nil
}
