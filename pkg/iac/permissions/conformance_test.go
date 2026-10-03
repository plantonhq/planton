package permissions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/shared/catalogkind"
)

var (
	// awsActionPattern is IAM's "service:Action" spelling. Wildcards are
	// syntactically legal; the gate separately demands a defense in notes.
	awsActionPattern = regexp.MustCompile(`^[a-z0-9-]+:[A-Za-z0-9*]+$`)
	// gcpPermissionPattern is GCP IAM's dotted permission form,
	// e.g. "container.clusters.create" (3+ segments).
	gcpPermissionPattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*(\.[a-zA-Z0-9]+){2,}$`)
	// azureActionPattern is Azure RBAC's "Provider.Namespace/type/action"
	// form, e.g. "Microsoft.ContainerService/managedClusters/write".
	azureActionPattern = regexp.MustCompile(`^[A-Za-z]+\.[A-Za-z]+(/[A-Za-z0-9*]+)+$`)
	// kubernetesVerbs is the closed RBAC verb vocabulary. "*" is absent
	// deliberately -- a wildcard verb is never least privilege. escalate
	// (on roles and clusterroles) and bind (on their bindings) are the
	// verbs Kubernetes defines for a principal that grants permissions it
	// does not itself hold -- what installing an operator IS -- so a kind
	// that installs one can state that exactly instead of being run as
	// cluster-admin in practice and described as least privilege on paper.
	kubernetesVerbs = map[string]bool{
		"get": true, "list": true, "watch": true, "create": true,
		"update": true, "patch": true, "delete": true, "deletecollection": true,
		"escalate": true, "bind": true,
	}
	// cloudflareScopePattern is Cloudflare's scope identifier spelling,
	// e.g. "com.cloudflare.api.account.zone". The spelling is checked
	// here; EXISTENCE of the (name, scope) pair is the inventory gate's
	// job (pkg/iac/actioninventory) -- Cloudflare grows the scope
	// vocabulary, so the snapshot, never a regex, is the closed set.
	cloudflareScopePattern = regexp.MustCompile(`^com\.cloudflare\.[a-z0-9]+(\.[a-z0-9]+)*$`)
	// digitalOceanScopePattern is DigitalOcean's "resource:action" token
	// scope spelling. Underscores are legal in BOTH segments, and action
	// segments go beyond CRUD (view_credentials, access_cluster, admin) --
	// a closed verb vocabulary here would wrongly reject real published
	// scopes, so the spelling is checked here and existence against the
	// provider's own inventory is the real check (pkg/iac/actioninventory).
	digitalOceanScopePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)
	// digitalOceanAliasScopes are the global alias scopes. Each expands to
	// every current AND future endpoint, so neither can ever appear in a
	// least-privilege manifest -- the gate refuses them outright.
	digitalOceanAliasScopes = map[string]bool{"api:read": true, "api:write": true}
	// digitalOceanSpacesPermissions is the CLOSED grant-level vocabulary of
	// DigitalOcean's Spaces keys, verified against the provider's own
	// Spaces-keys API reference (a grant is {bucket, permission} with
	// exactly these levels). Three values from the provider's API contract
	// -- no machine inventory exists to snapshot, so this closed set IS
	// the existence check, stated here rather than silently absent from
	// the inventory gates.
	digitalOceanSpacesPermissions = map[string]bool{"read": true, "readwrite": true, "fullaccess": true}
	// auth0ScopePattern is Auth0's Management API "verb:resource" scope
	// spelling -- the verb FIRST, the reverse of DigitalOcean's. Verbs go
	// beyond CRUD ("blacklist:tokens"), so the verb is not a closed set
	// here; existence against the tenant's own Management API definition
	// is the real check (pkg/iac/actioninventory).
	auth0ScopePattern = regexp.MustCompile(`^[a-z]+:[a-z_]+$`)
	// tokenScopedProviders are catalog providers whose modules authenticate
	// with a bearer credential that offers NO finer-grained vocabulary --
	// so, per the schema's own absence semantics, their manifests may
	// legally carry no provider section. openfga is the one tenant: a
	// pre-shared key grants the server's entire API, so there is nothing
	// a manifest could truthfully declare. The exemption is the honest
	// model, not a schema gap.
	//
	// Cloudflare, DigitalOcean, and Auth0 left this map when their arms
	// landed: their manifests declare token permission groups / scopes as
	// first-class sections, held to the providers' own inventories. Stripe
	// never joined it: a restricted key has a per-resource vocabulary, so
	// its manifests declare a stripe section too.
	tokenScopedProviders = map[string]bool{
		"openfga": true,
	}
)

// TestPermissionsConformance holds every authored permissions manifest to
// its contract, offline:
//
//  1. The manifest parses strictly against its proto schema and names its
//     kind (metadata.name equals the kind directory).
//  2. It declares at least one provider section -- a permissions file that
//     grants nothing describes no module.
//  3. Every entry is structurally sound for its provider (action spelling,
//     RBAC verb vocabulary), carries provenance, and defends its trust
//     posture: derived entries cite the module resources they cover, and
//     any wildcard resource scope carries a defense in notes.
//
// What this gate deliberately CANNOT prove: that the actions are SUFFICIENT
// or MINIMAL against a live cloud. That proof is the capture harness's job
// (e2e running under a role built from this very manifest); until an entry
// graduates to proven, its provenance says so honestly.
func TestPermissionsConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	discovered, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}
	if len(discovered) == 0 {
		t.Skip("no permissions manifests authored yet")
	}

	for provider, kindDirs := range discovered {
		for _, kindDir := range kindDirs {
			kindDir := kindDir
			t.Run(provider+"/"+kindDir, func(t *testing.T) {
				manifest, err := Load(root, provider, kindDir)
				if err != nil {
					t.Fatalf("permissions manifest: %v", err)
				}
				if manifest.GetKind() != "CatalogKindPermissions" {
					t.Fatalf("kind is %q, want CatalogKindPermissions", manifest.GetKind())
				}
				if manifest.GetMetadata().GetName() != kindDir {
					t.Errorf("metadata.name is %q, want %q", manifest.GetMetadata().GetName(), kindDir)
				}

				spec := manifest.GetSpec()
				// Counted from the schema's own populated fields, never a
				// hand-written list of sections: a provider section added
				// to the schema counts here the day it lands.
				declared := 0
				spec.ProtoReflect().Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool {
					declared++
					return true
				})
				if declared == 0 && !tokenScopedProviders[provider] {
					t.Fatal("manifest declares no provider section -- a permissions file that grants nothing describes no module")
				}

				checkAws(t, spec.GetAws())
				checkGcp(t, spec.GetGcp())
				checkAzure(t, spec.GetAzure())
				checkKubernetes(t, spec.GetKubernetes())
				checkCloudflare(t, spec.GetCloudflare())
				checkDigitalOcean(t, spec.GetDigitalOcean())
				checkAuth0(t, spec.GetAuth0())
				checkStripe(t, spec.GetStripe())
				checkConditions(t, kindDir, spec)
			})
		}
	}
}

// imageDeployingGcpKinds deploy a container image their deploying identity must be able to
// read: Google's Cloud Run deploy checks the deployer's own Artifact Registry access to the
// image, separately from the service agent that pulls it when a revision starts.
var imageDeployingGcpKinds = []string{"gcpcloudrun", "gcpcloudrunjob"}

const artifactRegistryDownload = "artifactregistry.repositories.downloadArtifacts"

// A customer who grants exactly what a manifest declares must be able to deploy with it. The
// structural gate above cannot see an absent grant, so the one Google's deploy contract requires
// of every image-deploying kind is pinned here by name.
func TestImageDeployingKindsDeclareRegistryRead(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}
	for _, kindDir := range imageDeployingGcpKinds {
		manifest, err := Load(root, "gcp", kindDir)
		if err != nil {
			t.Fatalf("%s permissions manifest: %v", kindDir, err)
		}
		declared := false
		for _, group := range manifest.GetSpec().GetGcp().GetGroups() {
			for _, permission := range group.GetPermissions() {
				declared = declared || permission == artifactRegistryDownload
			}
		}
		if !declared {
			t.Errorf("%s deploys a container image but its manifest grants no %s -- a customer who grants "+
				"exactly this manifest is refused at deploy when the image lives in Artifact Registry",
				kindDir, artifactRegistryDownload)
		}
	}
}

// skipAwaitSetOutsideTheCall are the kinds that put pulumi.com/skipAwait on a kind where the
// gate cannot see it at the constructor call -- the annotation map is built elsewhere in the
// module and passed in. Each entry stays true only while the module still carries the
// annotation; a module that drops it fails the gate here instead of keeping an exemption it no
// longer earns.
var skipAwaitSetOutsideTheCall = map[string]string{
	// locals.go: a claim under a WaitForFirstConsumer class is Pending until a pod uses it.
	"kubernetespersistentvolumeclaim": "kubernetes:core/v1:PersistentVolumeClaim",
}

// yamlPartitionsOutsideSkipAwait are the kinds that split their YAML by kind across several
// yaml constructors and put the skipAwait transformation on all but the ones holding only these
// resources (manifest_documents.go in each module: the Namespace and the CRDs apply on their own,
// the operator's workloads under skipAwait). Neither listed kind has a readiness wait, so honouring
// the transformation for every other kind is exact. Each entry stays true only while the module
// still applies the transformation to some constructor and not to all of them.
var yamlPartitionsOutsideSkipAwait = map[string][]string{
	"kuberneteskeycloakoperator": {"namespaces", "customresourcedefinitions"},
	"kubernetestektonoperator":   {"namespaces", "customresourcedefinitions"},
}

// A customer who grants exactly what a manifest declares must be able to deploy and destroy with
// the Pulumi module, not only with OpenTofu. Pulumi's Kubernetes provider waits after every
// create, update and delete, and every wait reads the cluster through informers that never start
// when their list is forbidden -- the deploy then hangs with no error rather than failing
// (pulumikubernetes.go says where, line by line). So for every Kubernetes object a kind's
// Pulumi module constructs, its manifest must grant what the provider reads while it waits for
// that object: get on the object, list and watch on its kind for the delete wait, and the
// readiness wait's reads (pods, replicasets, endpoints, events, ...) unless the object carries
// pulumi.com/skipAwait.
//
// A yaml ConfigFile or ConfigGroup (or a Helm v3 Chart) registers every document as an ordinary
// resource of its kind, which the provider awaits and deletes like a typed object, but the kinds
// live in the YAML, not the source. For a module that uses one, every kind the manifest grants
// create on is held as a created object, honouring a skipAwait or RetainOnDelete transformation
// the module applies to all of them. CRDs the keptcrds helper applies are held the same way, their
// delete wait only where the manifest lets the module delete them (it retains them otherwise).
//
// Out of this gate, by construction: a Helm release -- Helm's own wait, bounded by the release's
// timeout, reads what it reads. The kind's other objects, the namespace beside the release
// above all, are still held here. A kind with no manifest publishes no role, so there is
// nothing to hold.
func TestPulumiKubernetesModulesDeclareWhatTheProviderReadsWhileItWaits(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}
	discovered, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}
	for _, kindDir := range discovered["kubernetes"] {
		moduleDir := filepath.Join(root, "catalog", "kubernetes", kindDir, "iac", "pulumi")
		if _, err := os.Stat(moduleDir); err != nil {
			continue
		}
		scan, err := scanPulumiModule(root, moduleDir)
		if err != nil {
			t.Fatalf("%s: reading the Pulumi module: %v", kindDir, err)
		}
		for _, problem := range scan.problems {
			t.Errorf("%s: %s", kindDir, problem)
		}
		manifest, err := Load(root, "kubernetes", kindDir)
		if err != nil {
			t.Fatalf("%s permissions manifest: %v", kindDir, err)
		}
		rules := manifest.GetSpec().GetKubernetes().GetRules()

		if exempted, ok := skipAwaitSetOutsideTheCall[kindDir]; ok {
			constructs := false
			for _, object := range scan.objects {
				constructs = constructs || (object.kind != nil && object.kind.Token == exempted)
			}
			if !scan.setsSkipAwait || !constructs {
				t.Errorf("%s is exempted from %s's readiness reads for setting pulumi.com/skipAwait, but its module no longer "+
					"creates that kind with the annotation set to \"true\" -- remove the exemption so the gate holds the readiness reads again", kindDir, exempted)
			}
		}
		if _, ok := yamlPartitionsOutsideSkipAwait[kindDir]; ok {
			some, all := false, len(scan.yamlCalls) > 0
			for _, call := range scan.yamlCalls {
				some, all = some || call.skipAwait, all && call.skipAwait
			}
			if !some || all {
				t.Errorf("%s is listed in yamlPartitionsOutsideSkipAwait, but its yaml constructors no longer split skipAwait that way -- update or remove the entry", kindDir)
			}
		}

		requirements, uncreatable := pulumiWaitRequirements(kindDir, scan, rules)
		missing := append([]string{}, uncreatable...)
		for _, g := range waitGaps(rules, requirements) {
			reasons := make([]string, 0, len(g.needs))
			for _, need := range g.needs {
				reasons = append(reasons, need.sentence())
			}
			missing = append(missing, fmt.Sprintf("%s: grant %s -- %s", g.key(), strings.Join(g.verbs, ", "), strings.Join(reasons, "; ")))
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			t.Errorf("%s: a customer who grants exactly this manifest hangs the Pulumi module -- pulumi-kubernetes %s reads, while it waits, what the manifest does not grant:\n  %s",
				kindDir, PulumiKubernetesProviderVersion, strings.Join(missing, "\n  "))
		}
	}
}

// waitNeed is one read the provider makes for one object the module creates.
type waitNeed struct {
	read  KubernetesRead
	phase string // "delete", "await" or "refresh"
	what  string // the object's token, a custom resource's kind, or a yaml child's resource
	where string // the constructor call, relative to the module directory
	via   string // the yaml constructor or helper that applies it, when not a constructor
}

func (n waitNeed) sentence() string {
	if n.via != "" {
		n.what = fmt.Sprintf("%s applied by the %s", n.what, n.via)
	}
	switch n.phase {
	case "delete":
		return fmt.Sprintf("destroying the %s created at %s waits on an informer of its own kind (%s)", n.what, n.where, n.read.Source)
	case "await":
		return fmt.Sprintf("the %s created at %s is awaited (%s)", n.what, n.where, n.read.Source)
	}
	return fmt.Sprintf("refreshing the %s created at %s reads it back (%s)", n.what, n.where, n.read.Source)
}

// pulumiWaitRequirements is every read the provider makes while it waits for the objects a
// module creates. uncreatable names the custom resources the manifest does not even let the
// module create -- the gate cannot say which scope their reads need, and the create is refused
// first anyway.
func pulumiWaitRequirements(kindDir string, scan *moduleScan, rules []*permissionsv1.KubernetesRule) (needs []waitNeed, uncreatable []string) {
	objects := append([]createdObject{}, scan.objects...)
	objects = append(objects, yamlChildren(kindDir, scan, rules)...)
	crdRow, _ := PulumiKubernetesKindByResource("apiextensions.k8s.io", "customresourcedefinitions")
	deletesCRDs := len(uncoveredVerbs(rules, KubernetesRead{KubernetesResource: crdRow.Self, Verbs: []string{"delete"}})) == 0
	for _, where := range scan.keptCRDs {
		row := crdRow
		objects = append(objects, createdObject{where: where, via: "keptcrds.Apply", kind: &row, retainDelete: !deletesCRDs})
	}
	for _, object := range scan.projectionApplies {
		group, kind, ok := kindProjection(kindDir)
		if !ok {
			uncreatable = append(uncreatable, fmt.Sprintf("%s: the module applies a custom resource through manifestcr.Apply, but the kind has no kubernetes_manifest_projection to say which", object.where))
			continue
		}
		object.crGroup, object.crKind = group, kind
		objects = append(objects, object)
	}
	for _, object := range objects {
		self, what := KubernetesResource{}, ""
		var awaitReads []KubernetesRead
		skipsDelete := false
		switch {
		case object.kind != nil:
			self, what, awaitReads, skipsDelete = object.kind.Self, object.kind.Token, object.kind.AwaitReads, object.kind.SkipAwaitSkipsDeleteWait
		case object.self != nil:
			self, what = *object.self, object.self.Resource
		default:
			resource := resourcePlural(object.crKind)
			clusterScoped, granted := createScope(rules, object.crGroup, resource)
			if !granted {
				uncreatable = append(uncreatable, fmt.Sprintf("%s/%s: the module creates a %s (%s) but no rule grants create on %s/%s -- "+
					"if the definition's plural is not %q, name the resource the definition declares",
					object.crGroup, resource, object.crKind, object.where, object.crGroup, resource, resource))
				continue
			}
			self, what = KubernetesResource{APIGroup: object.crGroup, Resource: resource, ClusterScoped: clusterScoped}, object.crKind
		}
		skipAwait := object.skipAwait || skipAwaitSetOutsideTheCall[kindDir] == what
		need := func(read KubernetesRead, phase string) {
			needs = append(needs, waitNeed{read: read, phase: phase, what: what, where: object.where, via: object.via})
		}
		if !object.retainDelete && !(skipAwait && skipsDelete) {
			need(DeletionReads(self), "delete")
		} else {
			need(KubernetesRead{KubernetesResource: self, Verbs: []string{"get"}, Source: "await.go:376-383 Read"}, "refresh")
		}
		if !skipAwait {
			for _, read := range awaitReads {
				need(read, "await")
			}
			if object.waitFor {
				for _, read := range WaitForReads(self) {
					need(read, "await")
				}
			}
		}
	}
	return needs, uncreatable
}

// yamlChildren is every object a module's yaml constructors may apply: each kind the manifest
// grants create on. Subresources and the review APIs are requests, not objects, and a wildcard
// names no kind to hold.
func yamlChildren(kindDir string, scan *moduleScan, rules []*permissionsv1.KubernetesRule) []createdObject {
	if len(scan.yamlCalls) == 0 {
		return nil
	}
	allSkip, someSkip, allRetain := true, false, true
	var sites, vias []string
	for _, call := range scan.yamlCalls {
		allSkip, someSkip, allRetain = allSkip && call.skipAwait, someSkip || call.skipAwait, allRetain && call.retain
		sites = append(sites, call.where)
		if !contains(vias, call.via) {
			vias = append(vias, call.via)
		}
	}
	outsideSkip := yamlPartitionsOutsideSkipAwait[kindDir]
	seen := map[KubernetesResource]bool{}
	var objects []createdObject
	for _, rule := range rules {
		if !contains(rule.GetVerbs(), "create") {
			continue
		}
		for _, group := range rule.GetApiGroups() {
			for _, resource := range rule.GetResources() {
				if group == "*" || resource == "*" || strings.Contains(resource, "/") || strings.HasSuffix(resource, "reviews") {
					continue
				}
				object := createdObject{
					where:        strings.Join(sites, ", "),
					via:          strings.Join(vias, " and "),
					skipAwait:    allSkip || (someSkip && outsideSkip != nil && !contains(outsideSkip, resource)),
					retainDelete: allRetain,
				}
				if row, ok := PulumiKubernetesKindByResource(group, resource); ok {
					if seen[row.Self] {
						continue
					}
					seen[row.Self] = true
					object.kind = &row
				} else {
					self := KubernetesResource{APIGroup: group, Resource: resource, ClusterScoped: rule.GetClusterScoped()}
					if seen[self] {
						continue
					}
					seen[self] = true
					object.self = &self
				}
				objects = append(objects, object)
			}
		}
	}
	return objects
}

// waitGap is one resource the manifest under-grants, with every need that asks for it -- a
// reader sees each missing grant once, with every object that needs it.
type waitGap struct {
	resource KubernetesResource
	verbs    []string
	needs    []waitNeed
}

func (g waitGap) key() string {
	scope := "in the object's namespace"
	if g.resource.ClusterScoped {
		scope = "cluster-wide"
	}
	return fmt.Sprintf("%s/%s %s", groupLabel(g.resource.APIGroup), g.resource.Resource, scope)
}

// waitGaps folds the needs the rules do not cover into one gap per resource, in a stable order.
func waitGaps(rules []*permissionsv1.KubernetesRule, needs []waitNeed) []waitGap {
	byResource := map[KubernetesResource]*waitGap{}
	verbs := map[KubernetesResource]map[string]bool{}
	var order []KubernetesResource
	for _, need := range needs {
		uncovered := uncoveredVerbs(rules, need.read)
		if len(uncovered) == 0 {
			continue
		}
		resource := need.read.KubernetesResource
		if byResource[resource] == nil {
			byResource[resource] = &waitGap{resource: resource}
			verbs[resource] = map[string]bool{}
			order = append(order, resource)
		}
		for _, verb := range uncovered {
			verbs[resource][verb] = true
		}
		byResource[resource].needs = append(byResource[resource].needs, need)
	}
	gaps := make([]waitGap, 0, len(order))
	for _, resource := range order {
		gap := byResource[resource]
		gap.verbs = sortedKeys(verbs[resource], verbOrder)
		gaps = append(gaps, *gap)
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i].key() < gaps[j].key() })
	return gaps
}

// uncoveredVerbs returns the read's verbs no rule grants on its resource. A cluster-scoped rule
// covers a namespaced read; a namespaced rule never covers a cluster-wide one.
func uncoveredVerbs(rules []*permissionsv1.KubernetesRule, read KubernetesRead) []string {
	var gaps []string
	for _, verb := range read.Verbs {
		covered := false
		for _, rule := range rules {
			if (rule.GetClusterScoped() || !read.ClusterScoped) &&
				contains(rule.GetApiGroups(), read.APIGroup) && contains(rule.GetResources(), read.Resource) && contains(rule.GetVerbs(), verb) {
				covered = true
				break
			}
		}
		if !covered {
			gaps = append(gaps, verb)
		}
	}
	return gaps
}

// createScope is the scope of the rule that lets the module create a custom resource -- the
// manifest's own statement of whether the kind is cluster-scoped.
func createScope(rules []*permissionsv1.KubernetesRule, group, resource string) (clusterScoped, granted bool) {
	for _, rule := range rules {
		if contains(rule.GetApiGroups(), group) && contains(rule.GetResources(), resource) && contains(rule.GetVerbs(), "create") {
			granted = true
			clusterScoped = clusterScoped || rule.GetClusterScoped()
		}
	}
	return clusterScoped, granted
}

// contains reports want among values, where RBAC's "*" matches anything.
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want || value == "*" {
			return true
		}
	}
	return false
}

func groupLabel(group string) string {
	if group == "" {
		return "core"
	}
	return group
}

// verbOrder lists verbs the way the manifests do.
var verbOrder = map[string]int{"get": 0, "list": 1, "watch": 2}

// sortedKeys returns a set's members, by rank where one is given and then alphabetically.
func sortedKeys(set map[string]bool, rank map[string]int) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if rank != nil && rank[keys[i]] != rank[keys[j]] {
			return rank[keys[i]] < rank[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys
}

func checkProvenance(t *testing.T, where string, provenance permissionsv1.Provenance, notes string) {
	t.Helper()
	switch provenance {
	case permissionsv1.Provenance_provenance_unspecified:
		t.Errorf("%s: provenance is unspecified -- derived and proven are never blurred", where)
	case permissionsv1.Provenance_derived:
		if strings.TrimSpace(notes) == "" {
			t.Errorf("%s: derived entry cites no module resources in notes -- a derivation without its source is unverifiable", where)
		}
	}
}

func checkAws(t *testing.T, aws *permissionsv1.AwsPermissions) {
	t.Helper()
	if aws == nil {
		return
	}
	if len(aws.GetStatements()) == 0 {
		t.Error("aws section is present but declares no statements")
	}
	sids := map[string]bool{}
	for _, statement := range aws.GetStatements() {
		sid := statement.GetSid()
		if sid == "" {
			t.Error("aws statement with empty sid")
			continue
		}
		if sids[sid] {
			t.Errorf("aws: duplicate sid %q", sid)
		}
		sids[sid] = true
		if len(statement.GetActions()) == 0 {
			t.Errorf("aws %s: no actions", sid)
		}
		for _, action := range statement.GetActions() {
			if !awsActionPattern.MatchString(action) {
				t.Errorf("aws %s: action %q is not service:Action spelling", sid, action)
			}
		}
		if len(statement.GetResources()) == 0 {
			t.Errorf("aws %s: no resources -- scope the statement or defend '*'", sid)
		}
		for _, resource := range statement.GetResources() {
			if resource == "*" && strings.TrimSpace(statement.GetNotes()) == "" {
				t.Errorf("aws %s: resource '*' without a defense in notes -- least privilege demands the reason", sid)
			}
		}
		checkProvenance(t, "aws "+sid, statement.GetProvenance(), statement.GetNotes())
	}
}

func checkGcp(t *testing.T, gcp *permissionsv1.GcpPermissions) {
	t.Helper()
	if gcp == nil {
		return
	}
	if len(gcp.GetGroups()) == 0 {
		t.Error("gcp section is present but declares no groups")
	}
	for _, group := range gcp.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("gcp group with empty purpose")
		}
		if len(group.GetPermissions()) == 0 {
			t.Errorf("gcp %s: no permissions", group.GetPurpose())
		}
		for _, permission := range group.GetPermissions() {
			if !gcpPermissionPattern.MatchString(permission) {
				t.Errorf("gcp %s: permission %q is not dotted IAM form", group.GetPurpose(), permission)
			}
		}
		checkProvenance(t, "gcp "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkAzure(t *testing.T, azure *permissionsv1.AzurePermissions) {
	t.Helper()
	if azure == nil {
		return
	}
	if len(azure.GetGroups()) == 0 {
		t.Error("azure section is present but declares no groups")
	}
	for _, group := range azure.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("azure group with empty purpose")
		}
		if len(group.GetActions()) == 0 && len(group.GetDataActions()) == 0 {
			t.Errorf("azure %s: no actions or data_actions", group.GetPurpose())
		}
		for _, action := range append(append([]string{}, group.GetActions()...), group.GetDataActions()...) {
			if !azureActionPattern.MatchString(action) {
				t.Errorf("azure %s: action %q is not Provider.Namespace/type/action form", group.GetPurpose(), action)
			}
		}
		checkProvenance(t, "azure "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkKubernetes(t *testing.T, kubernetes *permissionsv1.KubernetesPermissions) {
	t.Helper()
	if kubernetes == nil {
		return
	}
	if len(kubernetes.GetRules()) == 0 {
		t.Error("kubernetes section is present but declares no rules")
	}
	for i, rule := range kubernetes.GetRules() {
		if len(rule.GetResources()) == 0 {
			t.Errorf("kubernetes rule %d: no resources", i)
		}
		if len(rule.GetVerbs()) == 0 {
			t.Errorf("kubernetes rule %d: no verbs", i)
		}
		for _, verb := range rule.GetVerbs() {
			if !kubernetesVerbs[verb] {
				t.Errorf("kubernetes rule %d: verb %q is not in the RBAC verb vocabulary (wildcards are never least privilege)", i, verb)
			}
		}
		checkProvenance(t, "kubernetes rule "+ruleLabel(rule), rule.GetProvenance(), rule.GetNotes())
	}
}

func ruleLabel(rule *permissionsv1.KubernetesRule) string {
	return strings.Join(rule.GetResources(), ",")
}

func checkCloudflare(t *testing.T, cloudflare *permissionsv1.CloudflarePermissions) {
	t.Helper()
	if cloudflare == nil {
		return
	}
	if len(cloudflare.GetGroups()) == 0 {
		t.Error("cloudflare section is present but declares no groups")
	}
	for _, group := range cloudflare.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("cloudflare group with empty purpose")
		}
		if group.GetName() == "" {
			t.Errorf("cloudflare %s: no permission-group name", group.GetPurpose())
		}
		if !cloudflareScopePattern.MatchString(group.GetScope()) {
			t.Errorf("cloudflare %s: scope %q is not Cloudflare's scope identifier spelling (e.g. com.cloudflare.api.account.zone)", group.GetPurpose(), group.GetScope())
		}
		checkProvenance(t, "cloudflare "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkDigitalOcean(t *testing.T, digitalOcean *permissionsv1.DigitalOceanPermissions) {
	t.Helper()
	if digitalOcean == nil {
		return
	}
	if len(digitalOcean.GetGroups()) == 0 && len(digitalOcean.GetSpacesGrants()) == 0 {
		t.Error("digitalocean section is present but declares no groups and no spaces grants")
	}
	for _, grant := range digitalOcean.GetSpacesGrants() {
		if strings.TrimSpace(grant.GetPurpose()) == "" {
			t.Error("digitalocean spaces grant with empty purpose")
		}
		if !digitalOceanSpacesPermissions[grant.GetPermission()] {
			t.Errorf("digitalocean spaces %s: permission %q is not a Spaces key grant level (read, readwrite, or fullaccess)", grant.GetPurpose(), grant.GetPermission())
		}
		checkProvenance(t, "digitalocean spaces "+grant.GetPurpose(), grant.GetProvenance(), grant.GetNotes())
	}
	for _, group := range digitalOcean.GetGroups() {
		if strings.TrimSpace(group.GetPurpose()) == "" {
			t.Error("digitalocean group with empty purpose")
		}
		if len(group.GetScopes()) == 0 {
			t.Errorf("digitalocean %s: no scopes", group.GetPurpose())
		}
		for _, scope := range group.GetScopes() {
			if digitalOceanAliasScopes[scope] {
				t.Errorf("digitalocean %s: scope %q is a global alias that expands to every current and future endpoint -- never least privilege; declare the resource scopes the modules actually need", group.GetPurpose(), scope)
				continue
			}
			if !digitalOceanScopePattern.MatchString(scope) {
				t.Errorf("digitalocean %s: scope %q is not resource:action spelling", group.GetPurpose(), scope)
			}
		}
		checkProvenance(t, "digitalocean "+group.GetPurpose(), group.GetProvenance(), group.GetNotes())
	}
}

func checkAuth0(t *testing.T, auth0 *permissionsv1.Auth0Permissions) {
	t.Helper()
	if auth0 == nil {
		return
	}
	if len(auth0.GetGroups()) == 0 {
		t.Error("auth0 section is present but declares no groups")
	}
	purposes := map[string]bool{}
	for _, group := range auth0.GetGroups() {
		purpose := group.GetPurpose()
		if strings.TrimSpace(purpose) == "" {
			t.Error("auth0 group with empty purpose")
		}
		if purposes[purpose] {
			t.Errorf("auth0: duplicate purpose %q -- a reader tells groups apart by purpose, so one purpose is one group", purpose)
		}
		purposes[purpose] = true
		if len(group.GetScopes()) == 0 {
			t.Errorf("auth0 %s: no scopes", purpose)
		}
		for _, scope := range group.GetScopes() {
			if !auth0ScopePattern.MatchString(scope) {
				t.Errorf("auth0 %s: scope %q is not verb:resource spelling (Auth0 puts the verb first, e.g. \"update:connections\")", purpose, scope)
			}
		}
		checkProvenance(t, "auth0 "+purpose, group.GetProvenance(), group.GetNotes())
	}
}

// stripeResourcePattern is a restricted-key form label: a capitalized first
// word, then words in either case ("Webhook Endpoints", "Customer portal"),
// never a snake_case identifier or an endpoint path.
var stripeResourcePattern = regexp.MustCompile(`^[A-Z][A-Za-z]*( [A-Za-z]+)*$`)

// checkStripe holds a Stripe section to its structure. Stripe publishes no
// inventory of restricted-key resources, so there is nothing to check the
// names against beyond their spelling; a live run under a key built from the
// entries proves them (see StripePermissions).
func checkStripe(t *testing.T, stripe *permissionsv1.StripePermissions) {
	t.Helper()
	if stripe == nil {
		return
	}
	if len(stripe.GetGroups()) == 0 {
		t.Error("stripe section is present but declares no groups")
	}
	purposes := map[string]bool{}
	for _, group := range stripe.GetGroups() {
		purpose := group.GetPurpose()
		if strings.TrimSpace(purpose) == "" {
			t.Error("stripe group with empty purpose")
		}
		if purposes[purpose] {
			t.Errorf("stripe: duplicate purpose %q -- a reader tells groups apart by purpose, so one purpose is one group", purpose)
		}
		purposes[purpose] = true
		if len(group.GetPermissions()) == 0 {
			t.Errorf("stripe %s: no permissions", purpose)
		}
		resources := map[string]bool{}
		for _, permission := range group.GetPermissions() {
			resource := permission.GetResource()
			if !stripeResourcePattern.MatchString(resource) {
				t.Errorf("stripe %s: resource %q is not a restricted-key form label (e.g. \"Webhook Endpoints\")", purpose, resource)
			}
			if resources[resource] {
				t.Errorf("stripe %s: resource %q listed twice -- the form has one row per resource", purpose, resource)
			}
			resources[resource] = true
			if access := permission.GetAccess(); access != "read" && access != "write" {
				t.Errorf("stripe %s: %s access %q is not \"read\" or \"write\" (leave a resource out for none)", purpose, resource, access)
			}
		}
		checkProvenance(t, "stripe "+purpose, group.GetProvenance(), group.GetNotes())
	}
}

// repoRoot walks up from this test file to the directory containing go.mod. Under Bazel the
// catalog tree is not part of the test's inputs, so the repo-reading gates skip there (the same
// posture as the sibling repo-reading gates) and run under go test and the catalog-data lane.
func repoRoot(t *testing.T) string {
	t.Helper()
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

// checkConditions proves every entry's condition names a field that exists
// in the kind's own spec. A condition naming a field the spec does not
// have would mark an entry optional forever -- no manifest could ever set
// it -- so a typo here silently drops a required grant from every chart that
// unions it. The walk reads the permissions schema's own descriptor (every
// section, every repeated entry list, every entry's `condition`), so an
// entry type added to the schema is covered the day it lands.
func checkConditions(t *testing.T, kindDir string, spec *permissionsv1.CatalogKindPermissionsSpec) {
	t.Helper()
	var specFields protoreflect.MessageDescriptor
	resolveSpec := func() protoreflect.MessageDescriptor {
		if specFields != nil {
			return specFields
		}
		// The kind directory's name, normalized to its kind the way the
		// catalog locates a kind's directory (catalogkindreflect.KindVersionDir).
		kind := catalogkindreflect.KindFromString(kindDir)
		if kind == catalogkind.CatalogKind_unspecified {
			t.Fatalf("an entry carries a condition, but kind %q resolves to no kind", kindDir)
		}
		instance, err := catalogkindreflect.NewInstance(kind)
		if err != nil {
			t.Fatalf("an entry carries a condition, but kind %s has no message: %v", kind, err)
		}
		specField := instance.ProtoReflect().Descriptor().Fields().ByName("spec")
		if specField == nil || specField.Message() == nil {
			t.Fatalf("kind %s's message has no spec to hold a condition's field", kind)
		}
		specFields = specField.Message()
		return specFields
	}

	spec.ProtoReflect().Range(func(_ protoreflect.FieldDescriptor, section protoreflect.Value) bool {
		section.Message().Range(func(listField protoreflect.FieldDescriptor, entries protoreflect.Value) bool {
			if !listField.IsList() || listField.Message() == nil {
				return true
			}
			conditionField := listField.Message().Fields().ByName("condition")
			if conditionField == nil {
				return true
			}
			list := entries.List()
			for i := 0; i < list.Len(); i++ {
				entry := list.Get(i).Message()
				if !entry.Has(conditionField) {
					continue
				}
				path := entry.Get(conditionField).Message().Interface().(*permissionsv1.PermissionCondition).GetSpecFieldSet()
				where := fmt.Sprintf("%s entry %d", listField.FullName(), i)
				if strings.TrimSpace(path) == "" {
					t.Errorf("%s: condition names no spec field -- an entry needed by every deployment carries no condition at all", where)
					continue
				}
				if err := specFieldExists(resolveSpec(), path); err != nil {
					t.Errorf("%s: condition names spec field %q, which %s does not have (%v) -- no manifest could ever set it, so the entry would never be granted", where, path, resolveSpec().FullName(), err)
				}
			}
			return true
		})
		return true
	})
}

// specFieldExists walks a dot-separated path of proto field names through a
// message descriptor. A repeated message field may be crossed (the condition
// then reads "any element sets the rest"); a map or a scalar may not, because
// no reader can say which entry the rest of the path would mean.
func specFieldExists(message protoreflect.MessageDescriptor, path string) error {
	segments := strings.Split(path, ".")
	for i, segment := range segments {
		field := message.Fields().ByName(protoreflect.Name(segment))
		if field == nil {
			return fmt.Errorf("no field %q in %s", segment, message.FullName())
		}
		if i < len(segments)-1 {
			if field.Message() == nil || field.IsMap() {
				return fmt.Errorf("%q is not a message or a list of messages, so %q cannot be walked", segment, segments[i+1])
			}
			message = field.Message()
		}
	}
	return nil
}

const (
	serviceUsageEnable = "serviceusage.services.enable"
	serviceUsageList   = "serviceusage.services.list"
)

// Google's provider enables an API (google_project_service) by first listing the services the
// project already has enabled, and reads one back the same way on every refresh -- so a group that
// grants the enable without the list is refused at the first deploy, before anything is created
// ("Permission denied to list services for consumer container"). The structural gate above cannot
// see an absent grant, so the pairing is pinned here for every GCP manifest.
func TestApiEnablingGroupsDeclareServiceList(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}
	discovered, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering permissions manifests: %v", err)
	}
	for _, kindDir := range discovered["gcp"] {
		manifest, err := Load(root, "gcp", kindDir)
		if err != nil {
			t.Fatalf("%s permissions manifest: %v", kindDir, err)
		}
		for _, group := range manifest.GetSpec().GetGcp().GetGroups() {
			enables, lists := false, false
			for _, permission := range group.GetPermissions() {
				enables = enables || permission == serviceUsageEnable
				lists = lists || permission == serviceUsageList
			}
			if enables && !lists {
				t.Errorf("%s group %q grants %s but no %s -- Google's provider lists the project's enabled services "+
					"before it enables one, so a customer who grants exactly this manifest is refused at the first deploy",
					kindDir, group.GetPurpose(), serviceUsageEnable, serviceUsageList)
			}
		}
	}
}

// kindProjection returns the custom resource group and kind a
// kind's registry entry projects onto (kubernetes_manifest_projection),
// reading the kind folder name as the lowercased kind name.
func kindProjection(kindDir string) (group, kind string, ok bool) {
	for name, number := range catalogkind.CatalogKind_value {
		if strings.ToLower(name) != kindDir {
			continue
		}
		proj := manifestprojection.ProjectionOf(catalogkind.CatalogKind(number))
		if proj == nil {
			return "", "", false
		}
		apiVersion := proj.GetApiVersion()
		if i := strings.Index(apiVersion, "/"); i >= 0 {
			group = apiVersion[:i]
		}
		return group, proj.GetKind(), true
	}
	return "", "", false
}
