package aa_e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	permissionsv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
	"github.com/plantonhq/planton/pkg/iac/permissions"
	"sigs.k8s.io/yaml"
)

// The identities a Kubernetes scenario may declare with planton.dev/e2e-identity.
//
//	declared                                     a ServiceAccount bound to exactly the rules the
//	                                             kind's iac/permissions.yaml declares
//	declared-minus:<apiGroup>/<resource>:<v>,<v> the same, with the named verbs withheld from the
//	                                             named resource (the core API group is "")
//
// "declared" proves a module's least-privilege claim by running the lane as
// it; "declared-minus" proves that a right the module needs is refused with
// that right named, and nothing else in the lane fails first. The rules bind
// through one ClusterRole whether the permissions file marks them cluster
// scoped or not: the lane's namespace does not exist before the deploy that
// creates it, so a namespaced Role has nowhere to live yet. The identity is
// exact in verbs and resources, and wider than production only in scope.
const (
	identityDeclared            = "declared"
	identityDeclaredMinusPrefix = "declared-minus:"
	// identityNamespace holds every lane identity the harness mints. Shared
	// and idempotently created; the objects inside are per lane.
	identityNamespace = "planton-e2e-identities"
)

// ProvisionIdentity implements provider.IdentityProvisioner: a ServiceAccount
// bound to the kind's declared rules (minus what the spec withholds), a
// short-lived token, and a self_managed provider configuration that
// authenticates as it against the harness's cluster. The configuration
// reaches both engines through the iac-input path a real deploy uses.
func (h *Harness) ProvisionIdentity(ctx context.Context, tc *provider.KindTestContext, spec string) (string, func(), error) {
	withheld, err := parseIdentitySpec(spec)
	if err != nil {
		return "", nil, err
	}
	declared, err := permissions.Load(tc.RepoRoot, tc.Provider, tc.Kind)
	if err != nil {
		return "", nil, errors.Wrapf(err, "the %q identity is built from the kind's permissions manifest", spec)
	}
	rules, err := clusterRoleRules(declared.GetSpec().GetKubernetes().GetRules(), withheld)
	if err != nil {
		return "", nil, err
	}
	if len(rules) == 0 {
		return "", nil, errors.Errorf("%s declares no Kubernetes rules, so there is nothing to bind the %q identity to", permissions.Path(tc.RepoRoot, tc.Provider, tc.Kind), spec)
	}

	name := identityName(tc)
	manifest, err := identityManifest(name, rules)
	if err != nil {
		return "", nil, err
	}
	if out, err := h.kubectl(ctx, bytes.NewReader(manifest), "apply", "-f", "-"); err != nil {
		return "", nil, errors.Wrapf(err, "creating the lane identity %s: %s", name, out)
	}
	cleanup := func() {
		// The namespace stays (shared across lanes); the identity's own
		// objects go. A failed delete is reported, never fatal: the lane's
		// verdict is already in.
		if out, err := h.kubectl(context.Background(), nil, "delete", "--ignore-not-found",
			"clusterrolebinding/"+name, "clusterrole/"+name, "-n", identityNamespace, "serviceaccount/"+name); err != nil {
			fmt.Printf("  [identity] WARN: removing %s failed: %v %s\n", name, err, out)
		}
	}

	token, err := h.kubectl(ctx, nil, "create", "token", name, "-n", identityNamespace, "--duration=4h")
	if err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "minting a token for the lane identity %s: %s", name, token)
	}
	kubeconfig, err := h.kubeconfigForToken(name, strings.TrimSpace(token))
	if err != nil {
		cleanup()
		return "", nil, err
	}
	providerConfig, err := yaml.Marshal(map[string]interface{}{
		"provider":     "self_managed",
		"self_managed": map[string]interface{}{"kube_config": kubeconfig},
	})
	if err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "encoding the lane identity's provider configuration")
	}
	dir, err := os.MkdirTemp("", "planton-e2e-identity-")
	if err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "creating a directory for the lane identity's provider configuration")
	}
	path := filepath.Join(dir, "provider-config.yaml")
	if err := os.WriteFile(path, providerConfig, 0o600); err != nil {
		cleanup()
		return "", nil, errors.Wrap(err, "writing the lane identity's provider configuration")
	}
	fmt.Printf("  [identity] ServiceAccount %s/%s bound to %d rule(s) from permissions.yaml\n", identityNamespace, name, len(rules))
	return path, func() {
		cleanup()
		os.RemoveAll(dir)
	}, nil
}

// withheldVerbs names verbs to withhold from one resource of one API group.
type withheldVerbs struct {
	apiGroup string
	resource string
	verbs    map[string]bool
}

func parseIdentitySpec(spec string) (*withheldVerbs, error) {
	if spec == identityDeclared {
		return nil, nil
	}
	if !strings.HasPrefix(spec, identityDeclaredMinusPrefix) {
		return nil, errors.Errorf("unknown identity %q: the Kubernetes harness accepts %q or %q<apiGroup>/<resource>:<verb>,<verb>",
			spec, identityDeclared, identityDeclaredMinusPrefix)
	}
	rest := strings.TrimPrefix(spec, identityDeclaredMinusPrefix)
	target, verbList, ok := strings.Cut(rest, ":")
	group, resource, okTarget := strings.Cut(target, "/")
	if !ok || !okTarget || resource == "" || verbList == "" {
		return nil, errors.Errorf("identity %q is malformed: expected %q<apiGroup>/<resource>:<verb>,<verb> (the core group is written as an empty string before the slash)",
			spec, identityDeclaredMinusPrefix)
	}
	withheld := &withheldVerbs{apiGroup: group, resource: resource, verbs: map[string]bool{}}
	for _, verb := range strings.Split(verbList, ",") {
		withheld.verbs[strings.TrimSpace(verb)] = true
	}
	return withheld, nil
}

// rbacRule is one PolicyRule as the ClusterRole manifest carries it.
type rbacRule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// clusterRoleRules turns the kind's declared rules into PolicyRules,
// withholding the named verbs from the named resource. A declared rule that
// names the resource beside others is split so only that resource loses the
// verbs. A withhold that another declared rule would grant anyway (a
// wildcard, as the generic Helm kind honestly declares for the arbitrary
// chart it installs) is refused: the lane would deploy with the right it
// claims to lack, and its "refusal" would never come.
func clusterRoleRules(declared []*permissionsv1.KubernetesRule, withheld *withheldVerbs) ([]rbacRule, error) {
	if withheld != nil {
		for _, rule := range declared {
			if !containsString(rule.GetApiGroups(), "*") && !containsString(rule.GetResources(), "*") {
				continue
			}
			if containsAny(rule.GetApiGroups(), withheld.apiGroup) && containsAny(rule.GetResources(), withheld.resource) {
				for verb := range withheld.verbs {
					if containsAny(rule.GetVerbs(), verb) {
						return nil, errors.Errorf("cannot withhold %s on %s/%s: the kind's permissions.yaml grants it through a wildcard rule (%v on %v), so the identity would hold it anyway; put this lane on a kind whose declared rules are exact",
							verb, withheld.apiGroup, withheld.resource, rule.GetApiGroups(), rule.GetResources())
					}
				}
			}
		}
	}
	var rules []rbacRule
	for _, rule := range declared {
		if withheld == nil || !containsString(rule.GetApiGroups(), withheld.apiGroup) || !containsString(rule.GetResources(), withheld.resource) {
			rules = append(rules, rbacRule{APIGroups: rule.GetApiGroups(), Resources: rule.GetResources(), Verbs: rule.GetVerbs()})
			continue
		}
		var others []string
		for _, resource := range rule.GetResources() {
			if resource != withheld.resource {
				others = append(others, resource)
			}
		}
		if len(others) > 0 {
			rules = append(rules, rbacRule{APIGroups: rule.GetApiGroups(), Resources: others, Verbs: rule.GetVerbs()})
		}
		var kept []string
		for _, verb := range rule.GetVerbs() {
			if !withheld.verbs[verb] {
				kept = append(kept, verb)
			}
		}
		if len(kept) > 0 {
			rules = append(rules, rbacRule{APIGroups: []string{withheld.apiGroup}, Resources: []string{withheld.resource}, Verbs: kept})
		}
	}
	return rules, nil
}

// containsAny reports whether the list names want or the wildcard.
func containsAny(list []string, want string) bool {
	return containsString(list, want) || containsString(list, "*")
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// identityName is unique per lane (kind, engine, run) and valid as a
// Kubernetes object name.
func identityName(tc *provider.KindTestContext) string {
	raw := strings.ToLower(fmt.Sprintf("lane-%s-%s-%s", tc.Kind, tc.Engine, tc.RunID))
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if len(name) > 63 {
		name = strings.Trim(name[:63], "-")
	}
	return name
}

// identityManifest is the namespace, the ServiceAccount, the ClusterRole, and
// the binding, as one multi-document YAML for kubectl apply.
func identityManifest(name string, rules []rbacRule) ([]byte, error) {
	documents := []map[string]interface{}{
		{
			"apiVersion": "v1", "kind": "Namespace",
			"metadata": map[string]interface{}{"name": identityNamespace},
		},
		{
			"apiVersion": "v1", "kind": "ServiceAccount",
			"metadata": map[string]interface{}{"name": name, "namespace": identityNamespace},
		},
		{
			"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole",
			"metadata": map[string]interface{}{"name": name},
			"rules":    rules,
		},
		{
			"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRoleBinding",
			"metadata": map[string]interface{}{"name": name},
			"roleRef":  map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "ClusterRole", "name": name},
			"subjects": []map[string]interface{}{{"kind": "ServiceAccount", "name": name, "namespace": identityNamespace}},
		},
	}
	var out bytes.Buffer
	for _, doc := range documents {
		encoded, err := yaml.Marshal(doc)
		if err != nil {
			return nil, errors.Wrap(err, "encoding the lane identity manifest")
		}
		out.WriteString("---\n")
		out.Write(encoded)
	}
	return out.Bytes(), nil
}

// kubeconfigForToken is a kubeconfig for the harness's cluster (its server
// and CA, read from the harness's own kubeconfig) authenticating with the
// minted token.
func (h *Harness) kubeconfigForToken(name, token string) (string, error) {
	raw, err := os.ReadFile(h.kubeconfigPath)
	if err != nil {
		return "", errors.Wrap(err, "reading the harness kubeconfig to derive the lane identity's")
	}
	var admin struct {
		Clusters []struct {
			Name    string                 `json:"name"`
			Cluster map[string]interface{} `json:"cluster"`
		} `json:"clusters"`
	}
	if err := yaml.Unmarshal(raw, &admin); err != nil {
		return "", errors.Wrap(err, "parsing the harness kubeconfig")
	}
	if len(admin.Clusters) == 0 {
		return "", errors.New("the harness kubeconfig names no cluster")
	}
	config := map[string]interface{}{
		"apiVersion":      "v1",
		"kind":            "Config",
		"clusters":        []map[string]interface{}{{"name": "lane", "cluster": admin.Clusters[0].Cluster}},
		"users":           []map[string]interface{}{{"name": name, "user": map[string]interface{}{"token": token}}},
		"contexts":        []map[string]interface{}{{"name": "lane", "context": map[string]interface{}{"cluster": "lane", "user": name}}},
		"current-context": "lane",
	}
	encoded, err := yaml.Marshal(config)
	if err != nil {
		return "", errors.Wrap(err, "encoding the lane identity's kubeconfig")
	}
	return string(encoded), nil
}

// kubectl runs kubectl against the harness's cluster as the harness itself.
func (h *Harness) kubectl(ctx context.Context, stdin *bytes.Reader, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", append([]string{"--kubeconfig", h.kubeconfigPath}, args...)...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}
