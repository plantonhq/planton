package permissions

import "strings"

// What Pulumi's Kubernetes provider reads while it waits.
//
// The provider the official Pulumi modules run (pulumi-kubernetes, pinned
// through this module's go.mod at sdk/v4 v4.33.0; the SDK asks the engine
// for the provider plugin of its own version) does not return when the API
// server accepts an object. It waits -- for a Deployment to roll out, a
// Service to get endpoints, a deleted object to be gone -- and every one of
// those waits reads the cluster through informers: a list, then a watch, of
// the object's own kind and of the kinds around it. A permissions manifest
// that grants create and delete but not those reads is refused by nothing
// and hangs Pulumi forever, which is worse than a refusal:
//
//   - informers/factory.go:99-145 (Factory.Subscribe) blocks in
//     cache.WaitForCacheSync on the provider's lifetime context
//     (provider/provider.go:207), not on the operation's timeout. An
//     informer whose list is forbidden never syncs, so Subscribe never
//     returns.
//   - The readiness waits subscribe BEFORE they start their clocks
//     (deployment.go:164-212: four Subscribes, then time.After(timeout)),
//     so the ten-minute timeout never begins. The Deployment reports Ready
//     and the deploy sits there, no error, until something outside the
//     provider kills it. OpenTofu's provider polls the one object with get
//     and never needs these reads, which is why the same manifest passes
//     there.
//   - Every delete waits on an informer of the deleted object's own kind
//     (await.go:984-1006 -> condition/source.go:142-154 ->
//     condition/deleted.go:69-96, which waits on that subscription even
//     after it has seen the object gone), so a kind the manifest can
//     delete but not list and watch hangs every destroy.
//
// Scope: every informer comes from a factory per namespace
// (informers/factory.go:51-82), created for the object's own namespace
// (await.go:317, 344, 502, 528, 987). A namespaced object's reads are
// therefore namespaced -- a Role in that namespace carries them -- and a
// cluster-scoped object's (a Namespace, a ClusterRole) are cluster-wide.
//
// The table models exactly the kinds the official modules construct through
// the SDK's typed constructors; a constructor it does not know fails the
// conformance gate rather than passing unread. Line numbers cite
// github.com/pulumi/pulumi-kubernetes at tag v4.33.0, provider/pkg/await/
// unless another directory is named. Re-derive the table when the SDK pin
// moves: the waits are the provider's implementation, not its contract.

// PulumiKubernetesProviderVersion is the provider release this table was
// read from.
const PulumiKubernetesProviderVersion = "v4.33.0"

// PulumiKubernetesSDK is the import path prefix of the provider's Go SDK.
const PulumiKubernetesSDK = "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

// KubernetesResource names one RBAC resource: its API group ("" for core),
// its lowercase plural, and whether it is cluster-scoped.
type KubernetesResource struct {
	APIGroup      string
	Resource      string
	ClusterScoped bool
}

// KubernetesRead is one read the provider makes while it waits, and where
// in the provider it makes it.
type KubernetesRead struct {
	KubernetesResource
	Verbs []string
	// Source cites the upstream line that makes the read.
	Source string
}

// PulumiKubernetesKind is one Kubernetes kind an official module creates
// through the SDK's typed constructor, and what the provider reads while it
// waits for one.
type PulumiKubernetesKind struct {
	// Token is the Pulumi resource type token the constructor registers.
	Token string
	// GoPackage and Constructor are how a module creates it.
	GoPackage   string
	Constructor string
	// Self is the kind's own API resource. Every kind is read back with
	// get on refresh (await.go:376-383) and, on delete, listed and watched
	// until it is gone (see DeletionReads).
	Self KubernetesResource
	// AwaitReads are the readiness wait's reads on create and update
	// (await.go:340-352, 524-536). Empty: the provider has no readiness
	// wait for the kind (awaiters.go:184-243 lists it as NONE or not at
	// all), so the object is done when the API server accepts it.
	AwaitReads []KubernetesRead
	// SkipAwaitSkipsDeleteWait is true for the kinds whose delete wait the
	// pulumi.com/skipAwait annotation also turns off
	// (metadata/overrides.go:168-208, allowsSkipAwaitWithDelete). For every
	// other kind the annotation turns off only the readiness wait.
	SkipAwaitSkipsDeleteWait bool
}

var (
	listWatch    = []string{"list", "watch"}
	getListWatch = []string{"get", "list", "watch"}
)

func informer(group, resource, source string) KubernetesRead {
	return KubernetesRead{KubernetesResource: KubernetesResource{APIGroup: group, Resource: resource}, Verbs: listWatch, Source: source}
}

// warningEvents is the event aggregator every create, update and delete
// runs beside its wait (await.go:352, 536, 1006 -> watchers.go:202-241): an
// informer on core Events in the object's namespace, filtered to the
// object, whose Warning events are what a person watching the deploy sees
// ("FailedScheduling", "BackOff pulling image"). It runs in its own
// goroutine (internal/awaiter.go:55-64), so a forbidden list does not stop
// the wait -- it silences those warnings and leaves an informer retrying
// for the life of the provider. It is required wherever the provider has a
// readiness wait, because that is where a person is waiting on it. It is
// core Events ("" group), never events.k8s.io.
var warningEvents = KubernetesRead{
	KubernetesResource: KubernetesResource{APIGroup: "", Resource: "events"},
	Verbs:              listWatch,
	Source:             "await.go:352 NewEventAggregator, watchers.go:213-215",
}

// DeletionReads is what the provider reads while it waits for a deleted
// object to be gone: its own kind, listed and watched in its own namespace
// (cluster-wide for a cluster-scoped kind), and read with get.
func DeletionReads(self KubernetesResource) KubernetesRead {
	return KubernetesRead{
		KubernetesResource: self,
		Verbs:              getListWatch,
		Source:             "await.go:984-1006 NewDeletionSource, condition/source.go:142-154, condition/deleted.go:69-96",
	}
}

func sdkKind(pkg, kind, group, resource string, clusterScoped bool, awaitReads ...KubernetesRead) PulumiKubernetesKind {
	// The token names the API group ("core" for the core group) where the
	// Go package path names the SDK's short one: networking/v1 registers
	// kubernetes:networking.k8s.io/v1:Ingress.
	tokenGroup := group
	if tokenGroup == "" {
		tokenGroup = "core"
	}
	version := pkg[strings.LastIndex(pkg, "/")+1:]
	return PulumiKubernetesKind{
		Token:       "kubernetes:" + tokenGroup + "/" + version + ":" + kind,
		GoPackage:   PulumiKubernetesSDK + "/" + pkg,
		Constructor: "New" + kind,
		Self:        KubernetesResource{APIGroup: group, Resource: resource, ClusterScoped: clusterScoped},
		AwaitReads:  awaitReads,
	}
}

// PulumiKubernetesKinds is the table, one row per typed constructor the
// official modules use (plus Pod, which the provider awaits and a module
// may create directly).
var PulumiKubernetesKinds = []PulumiKubernetesKind{
	// deployment.go:164-201 subscribes deployments, replicasets, pods and
	// persistentvolumeclaims before its clock starts (206-212); its Read
	// (224-253) lists the same three.
	withSkipDelete(sdkKind("apps/v1", "Deployment", "apps", "deployments", false,
		informer("apps", "deployments", "deployment.go:164-165"),
		informer("apps", "replicasets", "deployment.go:174-175"),
		informer("", "pods", "deployment.go:184-185"),
		informer("", "persistentvolumeclaims", "deployment.go:194-195"),
		warningEvents)),
	// statefulset.go:195-206 subscribes statefulsets and pods; its Read
	// lists pods (244).
	withSkipDelete(sdkKind("apps/v1", "StatefulSet", "apps", "statefulsets", false,
		informer("apps", "statefulsets", "statefulset.go:195-196"),
		informer("", "pods", "statefulset.go:205-206"),
		warningEvents)),
	// daemonset.go:142-153 subscribes daemonsets and pods.
	withSkipDelete(sdkKind("apps/v1", "DaemonSet", "apps", "daemonsets", false,
		informer("apps", "daemonsets", "daemonset.go:142-143"),
		informer("", "pods", "daemonset.go:152-153"),
		warningEvents)),
	// job.go:102-114 subscribes jobs and pods.
	withSkipDelete(sdkKind("batch/v1", "Job", "batch", "jobs", false,
		informer("batch", "jobs", "job.go:102-103"),
		informer("", "pods", "job.go:113-114"),
		warningEvents)),
	// A CronJob has no readiness wait: awaiters.go:184-243 has no entry.
	sdkKind("batch/v1", "CronJob", "batch", "cronjobs", false),
	// pod.go:150-151 subscribes pods.
	withSkipDelete(sdkKind("core/v1", "Pod", "", "pods", false,
		informer("", "pods", "pod.go:150-151"),
		warningEvents)),
	// service.go:137-148 subscribes services and endpoints (core
	// Endpoints, not EndpointSlices); its Read lists endpoints (185).
	sdkKind("core/v1", "Service", "", "services", false,
		informer("", "services", "service.go:137-138"),
		informer("", "endpoints", "service.go:147-148"),
		warningEvents),
	// ingress.go:116-137 subscribes ingresses, endpoints and services; its
	// Read lists endpoints and services (171-177).
	sdkKind("networking/v1", "Ingress", "networking.k8s.io", "ingresses", false,
		informer("networking.k8s.io", "ingresses", "ingress.go:116-117"),
		informer("", "endpoints", "ingress.go:126-127"),
		informer("", "services", "ingress.go:136-137"),
		warningEvents),
	// awaiters.go:288-322 polls the claim with get for up to five minutes
	// (watcher/watcher.go:48-63) and reads its StorageClass's binding mode
	// (326-349). Without that get the mode is unknown, so a claim under a
	// WaitForFirstConsumer class -- correctly Pending until a pod uses it
	// -- is awaited to Bound and fails the deploy after five minutes.
	sdkKind("core/v1", "PersistentVolumeClaim", "", "persistentvolumeclaims", false,
		KubernetesRead{
			KubernetesResource: KubernetesResource{APIGroup: "storage.k8s.io", Resource: "storageclasses", ClusterScoped: true},
			Verbs:              []string{"get"},
			Source:             "awaiters.go:326-349, pvcBindMode",
		},
		warningEvents),
	// awaiters.go:400-423 polls the quota with get for up to a minute;
	// get on the kind itself is already a Self read.
	sdkKind("core/v1", "ResourceQuota", "", "resourcequotas", false, warningEvents),
	// A Secret is awaited only when it is a service-account-token Secret
	// (awaiters.go:431-462), by get on itself; the official modules create
	// Opaque and dockerconfigjson Secrets, and get is a Self read anyway.
	sdkKind("core/v1", "Secret", "", "secrets", false),
	// A ServiceAccount is not awaited on Kubernetes 1.24 and later
	// (awaiters.go:470-477).
	sdkKind("core/v1", "ServiceAccount", "", "serviceaccounts", false),
	withSkipDelete(sdkKind("core/v1", "Namespace", "", "namespaces", true)),
	sdkKind("core/v1", "ConfigMap", "", "configmaps", false),
	sdkKind("core/v1", "LimitRange", "", "limitranges", false),
	sdkKind("networking/v1", "NetworkPolicy", "networking.k8s.io", "networkpolicies", false),
	sdkKind("policy/v1", "PodDisruptionBudget", "policy", "poddisruptionbudgets", false),
	// autoscaling/v2 is absent from awaiters.go:184-243 (only v1 is listed,
	// as NONE), so the HPA has no readiness wait.
	sdkKind("autoscaling/v2", "HorizontalPodAutoscaler", "autoscaling", "horizontalpodautoscalers", false),
	sdkKind("rbac/v1", "Role", "rbac.authorization.k8s.io", "roles", false),
	sdkKind("rbac/v1", "RoleBinding", "rbac.authorization.k8s.io", "rolebindings", false),
	sdkKind("rbac/v1", "ClusterRole", "rbac.authorization.k8s.io", "clusterroles", true),
	sdkKind("rbac/v1", "ClusterRoleBinding", "rbac.authorization.k8s.io", "clusterrolebindings", true),
	sdkKind("storage/v1", "StorageClass", "storage.k8s.io", "storageclasses", true),
	sdkKind("scheduling/v1", "PriorityClass", "scheduling.k8s.io", "priorityclasses", true),
	sdkKind("admissionregistration/v1", "ValidatingWebhookConfiguration", "admissionregistration.k8s.io", "validatingwebhookconfigurations", true),
	sdkKind("admissionregistration/v1", "MutatingWebhookConfiguration", "admissionregistration.k8s.io", "mutatingwebhookconfigurations", true),
	// No readiness wait: awaiters.go:184-243 has no CustomResourceDefinition
	// entry. The CRDs a yaml file or the keptcrds helper applies are held to
	// this row.
	sdkKind("apiextensions/v1", "CustomResourceDefinition", "apiextensions.k8s.io", "customresourcedefinitions", true),
}

// WaitForReads is what an object annotated pulumi.com/waitFor adds: its
// JSONPath or condition is judged by an observer of the object's own kind
// (metadata/overrides.go:112-160, condition/jsonpath.go:26-34), a readiness
// wait like any other -- the kind listed and watched before the clock
// starts, beside the Warning events.
func WaitForReads(self KubernetesResource) []KubernetesRead {
	return []KubernetesRead{
		{KubernetesResource: self, Verbs: listWatch, Source: "metadata/overrides.go:112-160, condition/jsonpath.go:26-34"},
		warningEvents,
	}
}

// PulumiKubernetesKindByResource returns the table's row for an API
// resource, for objects a yaml file applies, which the source names only by
// what the manifest grants.
func PulumiKubernetesKindByResource(group, resource string) (PulumiKubernetesKind, bool) {
	for _, kind := range PulumiKubernetesKinds {
		if kind.Self.APIGroup == group && kind.Self.Resource == resource {
			return kind, true
		}
	}
	return PulumiKubernetesKind{}, false
}

// withSkipDelete marks the kinds metadata/overrides.go:191-208 lets
// pulumi.com/skipAwait release from the delete wait too.
func withSkipDelete(kind PulumiKubernetesKind) PulumiKubernetesKind {
	kind.SkipAwaitSkipsDeleteWait = true
	return kind
}

// PulumiKubernetesCustomResource is the SDK's untyped constructor for any
// custom resource (apiextensions.CustomResource). Its kind is read from the
// ApiVersion and Kind the module passes; the provider awaits no custom
// resource by default (metadata/overrides.go:112-116), so only the delete
// wait's reads apply -- and skipAwait never releases a custom resource from
// that (allowsSkipAwaitWithDelete lists built-in kinds only).
const PulumiKubernetesCustomResource = "NewCustomResource"

// PulumiKubernetesCustomResourcePackage is the package that holds it.
const PulumiKubernetesCustomResourcePackage = PulumiKubernetesSDK + "/apiextensions"

// PulumiKubernetesDelegated are the constructors whose children the source
// does not name, each with how the gate holds them.
//
// The yaml constructors register every document as an ordinary resource of
// its kind, which the provider awaits and deletes exactly as it would a
// typed object. The gate cannot read the kinds from the source, so it holds
// every kind the component's manifest grants create on to this table (see
// the conformance gate). A Helm release is Helm's: installed, awaited and
// uninstalled by Helm inside the provider, whose own wait is bounded by the
// release's timeout.
var PulumiKubernetesDelegated = map[string]string{
	PulumiKubernetesSDK + "/helm/v3.NewRelease":     "a Helm release is installed, awaited and uninstalled by Helm inside the provider -- Helm's own wait, bounded by the release's timeout, reads what it needs; this table models the provider's informer waits only",
	PulumiKubernetesSDK + "/helm/v3.NewChart":       "a Helm v3 Chart renders its templates into child resources at run time, each awaited by its own kind; no component with a manifest uses one, and one that does is held like a yaml file",
	PulumiKubernetesSDK + "/yaml.NewConfigFile":     "yaml",
	PulumiKubernetesSDK + "/yaml.NewConfigGroup":    "yaml",
	PulumiKubernetesSDK + "/yaml/v2.NewConfigGroup": "yaml",
	PulumiKubernetesSDK + "/yaml/v2.NewConfigFile":  "yaml",
	PulumiKubernetesSDK + ".NewProvider":            "the provider itself is a Pulumi resource, not a cluster object",
}

// PulumiKubernetesKindByConstructor returns the table's row for a typed
// constructor.
func PulumiKubernetesKindByConstructor(goPackage, constructor string) (PulumiKubernetesKind, bool) {
	for _, kind := range PulumiKubernetesKinds {
		if kind.GoPackage == goPackage && kind.Constructor == constructor {
			return kind, true
		}
	}
	return PulumiKubernetesKind{}, false
}
