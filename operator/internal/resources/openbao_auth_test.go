package resources

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

// The access arrangement's words are pinned here: what each policy grants
// and, as importantly, what it does not; the names the roles, the Secret,
// and the grant carry; the note a person reads on the token Secret.

// The control plane's policy is its two engines and nothing under sys/ or
// auth/: the boundary that matters, stated at the mount.
func TestOpenBAOControlPlanePolicy(t *testing.T) {
	policy := OpenBAOControlPlanePolicy()
	for _, want := range []string{`path "secret/*"`, `path "transit/*"`, `"create", "read", "update", "delete", "list", "patch"`} {
		if !strings.Contains(policy, want) {
			t.Errorf("control plane policy must carry %s, got:\n%s", want, policy)
		}
	}
	for _, forbid := range []string{`"sys/`, `"auth/`, "sudo", "root"} {
		if strings.Contains(policy, forbid) {
			t.Errorf("control plane policy must not carry %s, got:\n%s", forbid, policy)
		}
	}
}

// The operator's policy is the platform's own objects and never the engines
// or a sudo path: it provisions, it does not read what the platform stores,
// and enabling the auth method (sudo) stays with the root token at init.
func TestOpenBAOOperatorPolicy(t *testing.T) {
	policy := OpenBAOOperatorPolicy("acme")
	for _, want := range []string{
		`path "sys/mounts"`, `path "sys/mounts/secret"`, `path "sys/mounts/transit"`,
		`path "sys/policies/acl/acme-operator"`, `path "sys/policies/acl/acme-control-plane"`,
		`path "auth/kubernetes/config"`, `path "auth/kubernetes/role/acme-operator"`,
		`path "auth/token/roles/acme-control-plane"`, `path "auth/token/create/acme-control-plane"`,
		`path "auth/token/lookup-accessor"`, `path "auth/token/renew-accessor"`, `path "auth/token/revoke-accessor"`,
	} {
		if !strings.Contains(policy, want) {
			t.Errorf("operator policy must carry %s, got:\n%s", want, policy)
		}
	}
	for _, forbid := range []string{`"secret/`, `"transit/`, "sudo", `"sys/auth`, `"auth/token/create"`, `"auth/token/create-orphan"`} {
		if strings.Contains(policy, forbid) {
			t.Errorf("operator policy must not carry %s, got:\n%s", forbid, policy)
		}
	}
	// Every path is the platform's own: nothing generic that another
	// platform's objects in the same vault could be reached through.
	for _, line := range strings.Split(policy, "\n") {
		if strings.HasPrefix(line, `path "sys/policies/acl/`) && !strings.Contains(line, "acme-") {
			t.Errorf("a policy path must name this platform's own policy, got %s", line)
		}
	}
}

// The names, in one place, so status sentences and the Deployment agree.
func TestOpenBAOAuthNames(t *testing.T) {
	if got := OpenBAOOperatorRoleName("acme"); got != "acme-operator" {
		t.Errorf("operator role = %q", got)
	}
	if got := OpenBAOControlPlaneRoleName("acme"); got != "acme-control-plane" {
		t.Errorf("control plane role = %q", got)
	}
	if got := OpenBAOTokenSecretName("acme"); got != "acme-openbao-token" {
		t.Errorf("token Secret = %q", got)
	}
	if got := OpenBAOServiceAccountName("acme"); got != "acme-openbao" {
		t.Errorf("vault ServiceAccount = %q (the chart names it after the release)", got)
	}
	if got := OpenBAOAuthDelegatorClusterRoleBindingName("prod", "acme"); got != "prod-acme-openbao-auth-delegator" {
		t.Errorf("auth-delegator binding = %q (namespace-qualified: cluster-scoped, platforms are namespaced)", got)
	}
	if OpenBAOKubernetesHost != "https://kubernetes.default.svc:443" {
		t.Errorf("the auth method's host must be the in-cluster name every cluster resolves, got %q", OpenBAOKubernetesHost)
	}
}

// The auth-delegator grant is a cluster-scoped satellite in the control
// plane's token-reviewer shape: the platform's UID as a label for the
// janitor, the vault's own ServiceAccount as the only subject, the cluster's
// own role.
func TestOpenBAOAuthDelegatorClusterRoleBinding(t *testing.T) {
	crb := OpenBAOAuthDelegatorClusterRoleBinding("prod", "acme", types.UID("uid-1"))
	if crb.Name != "prod-acme-openbao-auth-delegator" {
		t.Errorf("name = %q", crb.Name)
	}
	if crb.Namespace != "" {
		t.Errorf("a ClusterRoleBinding has no namespace, got %q", crb.Namespace)
	}
	if crb.Labels[PlatformUIDLabel] != "uid-1" {
		t.Errorf("the binding must carry the platform's UID for the janitor, got labels %v", crb.Labels)
	}
	if crb.Labels["app.kubernetes.io/managed-by"] != ManagedByLabel {
		t.Errorf("the binding must carry the operator's mark, got labels %v", crb.Labels)
	}
	if crb.RoleRef.Kind != "ClusterRole" || crb.RoleRef.Name != "system:auth-delegator" {
		t.Errorf("roleRef = %+v, want the cluster's system:auth-delegator", crb.RoleRef)
	}
	if len(crb.Subjects) != 1 || crb.Subjects[0].Kind != "ServiceAccount" || crb.Subjects[0].Name != "acme-openbao" || crb.Subjects[0].Namespace != "prod" {
		t.Errorf("subjects = %+v, want only the vault's ServiceAccount in the platform's namespace", crb.Subjects)
	}
	if len(crb.OwnerReferences) != 0 {
		t.Error("a cluster-scoped object cannot carry a namespaced owner; the UID label is its ownership")
	}
	if bare := OpenBAOAuthDelegatorClusterRoleBinding("prod", "acme", ""); bare.Labels[PlatformUIDLabel] != "" {
		t.Error("no UID means no UID label -- the janitor never guesses ownership")
	}
}

// The token Secret's note says what the token is and is not.
func TestOpenBAOTokenSecretNote(t *testing.T) {
	note := OpenBAOTokenSecretNote("acme")
	for _, want := range []string{"acme-openbao", "acme-control-plane", "secret/ and transit/", "accessor", "mints a new one", "restarts the control plane", "Nothing here is the vault's root token", "break-glass"} {
		if !strings.Contains(note, want) {
			t.Errorf("note must say %q, got: %s", want, note)
		}
	}
}
