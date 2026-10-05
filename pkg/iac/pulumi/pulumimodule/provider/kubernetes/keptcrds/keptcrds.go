// Package keptcrds applies a Helm chart's CustomResourceDefinitions as
// module-owned, kept resources ahead of the chart's release: the Pulumi half of
// the catalog's derive-branch primitive for charts that carry CRDs.
//
// A module on the derive branch does three things, and this package does the
// first two for it: derive the CRD set from the pinned version (through
// pkg/kubernetes/helmcrds, in-process), apply each CRD keyed by its own name
// as a kept resource that re-adopts on reinstall, refuses a schema downgrade,
// refuses to take over a CRD someone else owns, and refuses before touching
// anything when the deploy's identity may not write CRDs; and then install the
// release with CRDs skipped and the chart's own CRD switch pinned off (the
// module's release code). The Terraform twin is the generated helm_crds.tf
// every module carries byte-identically.
//
// Why the mechanics look the way they do, each verified live:
//
//   - Keep-on-uninstall rides a retainOnDelete resource TRANSFORMATION on a
//     classic-yaml ConfigGroup. Neither yaml package forwards ordinary options
//     to the CRD children (classic passes only parent/version/pluginDownloadURL;
//     yaml/v2 creates children provider-side, beyond any SDK option), and a
//     transformation is the one mechanism the SDK propagates down the parent
//     chain to in-process children.
//   - Re-adoption rides a dedicated upsert provider (server-side apply with
//     upsertExistingObjects and force): a destroy leaves kept CRDs on the
//     cluster by design, so the next install finds them there and a plain
//     create fails AlreadyExists. Only the CRDs use that provider; the release
//     keeps create-conflict semantics.
//   - The same server-side apply would also take over a CRD the module never
//     stamped (a hand-run helm install, a kubectl apply, another tool) without
//     a word, and the version check can only order versions it stamped, so a
//     newer schema could be lowered silently. Every CRD about to be written is
//     therefore read BY NAME first, and one read answers both questions: is it
//     ours (the stamp), and is it newer than what we would write (the version).
//   - The permission probe asks the API server (SelfSubjectAccessReview)
//     whether this identity may write CRDs, so a namespace-admin identity is
//     refused at preview with the missing right named, instead of at the
//     first apply with the server's bare "forbidden".
//   - All reads run in-process during preview too, before any resource
//     registers, through the same kubeconfig the provider uses.
package keptcrds

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// Args describes one module's CRD set and how to keep it.
type Args struct {
	// Source is where the CRDs come from at the pinned version. Its Values
	// MUST be the release's own values documents in helm -f order, with only
	// the CRD switch turned on in CRDOverride: templated CRDs depend on
	// other values (fullname, webhook trust, ports), and a render from a
	// minimal set produces CRDs that point at the wrong release.
	Source helmcrds.Source
	// Policy says what the kind accepts: whether a chart without CRDs is a
	// failure, and whether CRDs Helm would own are an acceptable choice. A
	// typed kind expects CRDs and allows nothing Helm-managed; the generic
	// Helm kind expects none and passes the user's spec.crds.allow_helm_managed.
	Policy helmcrds.Policy
	// ReleaseName and Namespace are the identity the render runs under.
	ReleaseName string
	Namespace   string
	// Install false is the bring-your-own-CRDs arm: nothing is derived or
	// applied and Apply returns no resources. The release must still skip
	// CRDs; with the CRDs absent the operator cannot start, so this is never
	// a lighter install.
	Install bool
	// KeepOnUninstall true (the default posture) retains every CRD on
	// destroy so an operator uninstall never cascade-deletes the custom
	// resources built on it. False lets destroy delete them: the arm e2e
	// lanes use to leave a cluster clean.
	KeepOnUninstall bool
	// ProviderConfig is the module's Kubernetes connection; nil means the
	// host kubeconfig (the local workflow), exactly as the provider getter
	// treats it.
	ProviderConfig *kubernetesprovider.KubernetesProviderConfig
	// ProviderName names the dedicated upsert provider resource; modules
	// keep it distinct from their plain provider ("kubernetes-crd-upsert").
	ProviderName string
}

// liveReadTimeout bounds each read the primitive makes before registering. A
// cluster that cannot answer a get in this time will not take an apply either.
const liveReadTimeout = 30 * time.Second

// Apply derives, checks, and registers the CRD set, returning the resources
// the release must depend on. Errors are helmcrds.Failure values wherever the
// cause is one the primitive anticipates.
func Apply(ctx *pulumi.Context, args Args) ([]pulumi.Resource, error) {
	if !args.Install {
		return nil, nil
	}
	if args.ProviderName == "" {
		args.ProviderName = "kubernetes-crd-upsert"
	}

	// A destroy re-runs the program only so delete hooks can fire; every
	// resource in state is deleted whatever is registered here. Deriving
	// can fail for reasons that have nothing to do with the teardown (the
	// manifest currently pins a version that was never published, the
	// repository is unreachable, a downgrade is pending), and none of them
	// may stand between a user and deleting their stack: register nothing
	// and let the destroy proceed.
	if iacinput.IsDestroy() {
		return nil, nil
	}

	cluster, err := connect(args)
	if err != nil {
		return nil, err
	}

	// The render sees exactly the Kubernetes the install will see: charts
	// declare kubeVersion constraints and gate templates on the version,
	// and a client-only render would otherwise assume Helm's built-in
	// default and refuse charts the install would accept.
	if args.Source.KubeVersion == "" {
		version, err := cluster.serverVersion()
		if err != nil {
			return nil, err
		}
		args.Source.KubeVersion = version
	}

	derived, err := helmcrds.Derive(context.Background(), args.Source, args.Policy, args.ReleaseName, args.Namespace)
	if err != nil {
		return nil, err
	}
	for _, name := range derived.HelmManaged {
		// Accepted by the kind's policy: say so once, where the deploy log is
		// read, so nobody later wonders why this CRD carries no stamp.
		_ = ctx.Log.Warn("CRD "+name+" is templated by the chart without "+helmcrds.HelmKeepAnnotation+": keep; Helm owns it and will delete it with the release", nil)
	}
	crds := derived.Owned
	if len(crds) == 0 {
		// Nothing to write, so nothing to own and no right to hold: a chart
		// without CRDs must never be refused for a CRD permission.
		return nil, nil
	}

	// The checks run during preview too: a foreign owner, a downgrade, and a
	// missing right are all refused at plan time in both engines, before
	// anything is touched. Ownership goes first: a CRD another source stamped
	// at a higher version is an ownership question, not a downgrade.
	existing, err := cluster.readExisting(crds)
	if err != nil {
		return nil, err
	}
	if err := helmcrds.CheckOwnership(existing, args.Source); err != nil {
		return nil, err
	}
	if err := helmcrds.CheckNoDowngrade(existing, args.Source.Version); err != nil {
		return nil, err
	}
	if err := cluster.refuseIfDenied(args); err != nil {
		return nil, err
	}
	for _, crd := range existing {
		// The explained success: a kept CRD from an earlier install is about
		// to be re-adopted, and the deploy log should say so rather than
		// leave a reader to infer it from a "create" that changed nothing.
		_ = ctx.Log.Info(fmt.Sprintf("re-adopting CRD %s the cluster already carries from %s at chart version %s", crd.Name, args.Source.SourceDescription(), crd.Version), nil)
	}

	upsertProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfigUpsert(ctx,
		args.ProviderConfig, args.ProviderName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the CRD upsert kubernetes provider")
	}

	options := []pulumi.ResourceOption{pulumi.Provider(upsertProvider)}
	if args.KeepOnUninstall {
		options = append(options, pulumi.Transformations([]pulumi.ResourceTransformation{
			func(targs *pulumi.ResourceTransformationArgs) *pulumi.ResourceTransformationResult {
				return &pulumi.ResourceTransformationResult{
					Props: targs.Props,
					Opts:  append(targs.Opts, pulumi.RetainOnDelete(true)),
				}
			},
		}))
	}

	// One ConfigGroup per CRD, named by the CRD's own metadata.name (helmcrds
	// sorts them), so state addresses stay stable across chart reorderings
	// and match the Terraform twin's for_each keys.
	resources := make([]pulumi.Resource, 0, len(crds))
	for _, crd := range crds {
		group, err := pulumiyaml.NewConfigGroup(ctx, crd.Name,
			&pulumiyaml.ConfigGroupArgs{YAML: []string{crd.YAML}}, options...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to apply CRD %s", crd.Name)
		}
		resources = append(resources, group)
	}
	return resources, nil
}

// clusterReads is the module's own connection to the cluster for the reads
// the primitive makes before anything registers: the server version (the
// render's capabilities), each CRD about to be written (ownership and the
// never-downgrade check), and the permission probe. It uses the same
// kubeconfig the provider uses.
type clusterReads struct {
	dynamic   dynamic.Interface
	discovery discovery.DiscoveryInterface
}

func connect(args Args) (*clusterReads, error) {
	restConfig, err := kubeconfig.RESTConfig(args.ProviderConfig, os.Getenv(execcredential.CommandPathEnvVar))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve the cluster connection for the CRD lifecycle reads")
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the cluster client for the CRD lifecycle reads")
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the cluster client for the server version read")
	}
	return &clusterReads{dynamic: dynamicClient, discovery: discoveryClient}, nil
}

var (
	crdGVR = schema.GroupVersionResource{
		Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions",
	}
	selfSubjectAccessReviewGVR = schema.GroupVersionResource{
		Group: "authorization.k8s.io", Version: "v1", Resource: "selfsubjectaccessreviews",
	}
	selfSubjectReviewGVR = schema.GroupVersionResource{
		Group: "authentication.k8s.io", Version: "v1", Resource: "selfsubjectreviews",
	}
)

// readExisting gets each CRD about to be written, by name. A name the cluster
// does not have is simply absent from the result (a first install finds none);
// any other error is the cluster's, and stops the deploy as itself.
func (c *clusterReads) readExisting(crds []helmcrds.CRD) ([]helmcrds.ExistingCRD, error) {
	existing := make([]helmcrds.ExistingCRD, 0, len(crds))
	for _, crd := range crds {
		ctx, cancel := context.WithTimeout(context.Background(), liveReadTimeout)
		object, err := c.dynamic.Resource(crdGVR).Get(ctx, crd.Name, metav1.GetOptions{})
		cancel()
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read CRD %s before applying it", crd.Name)
		}
		existing = append(existing, toExisting(*object))
	}
	return existing, nil
}

// toExisting maps an object as the API server returns it onto what the
// ownership and version checks compare.
func toExisting(object unstructured.Unstructured) helmcrds.ExistingCRD {
	managers := make([]string, 0, len(object.GetManagedFields()))
	for _, entry := range object.GetManagedFields() {
		managers = append(managers, entry.Manager)
	}
	return helmcrds.ExistingFromObject(object.GetName(), object.GetLabels(), object.GetAnnotations(), managers)
}

// serverVersion is the cluster's Kubernetes version as the API server
// reports it (for example "v1.31.0").
func (c *clusterReads) serverVersion() (string, error) {
	info, err := c.discovery.ServerVersion()
	if err != nil {
		return "", errors.Wrap(err, "failed to read the cluster's Kubernetes version for the CRD render")
	}
	return info.GitVersion, nil
}

// crdVerbs are the rights the apply needs on CustomResourceDefinitions:
// server-side apply is a patch that creates when absent, and the provider
// reads the object back; delete is exercised only when a destroy is meant
// to take the CRDs with it.
func crdVerbs(keepOnUninstall bool) []string {
	verbs := []string{"get", "create", "patch"}
	if !keepOnUninstall {
		verbs = append(verbs, "delete")
	}
	return verbs
}

// refuseIfDenied asks the API server whether this identity may take each
// verb the apply needs on CRDs at the cluster scope, and refuses with the
// first one denied, named, before anything registers. The review APIs are
// open to every authenticated identity (system:basic-user), so the probe
// itself cannot be the thing that is forbidden.
func (c *clusterReads) refuseIfDenied(args Args) error {
	for _, verb := range crdVerbs(args.KeepOnUninstall) {
		allowed, reason, err := c.canI(verb)
		if err != nil {
			return err
		}
		if !allowed {
			return helmcrds.CRDApplyDeniedFailure(c.whoAmI(), verb, reason)
		}
	}
	return nil
}

func (c *clusterReads) canI(verb string) (bool, string, error) {
	review := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "authorization.k8s.io/v1",
		"kind":       "SelfSubjectAccessReview",
		"spec": map[string]interface{}{
			"resourceAttributes": map[string]interface{}{
				"group":    crdGVR.Group,
				"resource": crdGVR.Resource,
				"verb":     verb,
			},
		},
	}}
	ctx, cancel := context.WithTimeout(context.Background(), liveReadTimeout)
	defer cancel()
	answer, err := c.dynamic.Resource(selfSubjectAccessReviewGVR).Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, "", errors.Wrapf(err, "failed to ask the cluster whether this identity may %s CustomResourceDefinitions", verb)
	}
	allowed, _, _ := unstructured.NestedBool(answer.Object, "status", "allowed")
	reason, _, _ := unstructured.NestedString(answer.Object, "status", "reason")
	if reason == "" {
		reason = fmt.Sprintf("SelfSubjectAccessReview for %s on %s.%s answered denied", verb, crdGVR.Resource, crdGVR.Group)
	}
	return allowed, reason, nil
}

// whoAmI names the identity the deploy runs as, from the cluster's own
// answer (SelfSubjectReview). Older clusters without the API get a plain
// description rather than a guess.
func (c *clusterReads) whoAmI() string {
	review := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "authentication.k8s.io/v1",
		"kind":       "SelfSubjectReview",
	}}
	ctx, cancel := context.WithTimeout(context.Background(), liveReadTimeout)
	defer cancel()
	answer, err := c.dynamic.Resource(selfSubjectReviewGVR).Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return "the identity in the kubeconfig this deploy uses"
	}
	username, _, _ := unstructured.NestedString(answer.Object, "status", "userInfo", "username")
	if strings.TrimSpace(username) == "" {
		return "the identity in the kubeconfig this deploy uses"
	}
	return username
}
