package resources

import (
	"fmt"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// How the platform signs in to its vault. The root token /sys/init hands out
// is the vault's break-glass and nothing running depends on it: the operator
// signs in with its own Kubernetes identity through the auth method it
// enables at initialization, and the control plane signs in with a token
// the operator mints for it through a token role, keeps alive, and issues
// again when it is lost. This file holds the names and the words of that
// arrangement -- the mount path, the two policies, the two roles, the
// Secret the control plane's token lives in -- so the component that drives
// it, the tests that pin it, and the status that describes it all read one
// place.
//
// Reference: the OpenBao Kubernetes auth method (its TokenReview call uses
// the vault pod's own ServiceAccount token, which is why that account needs
// system:auth-delegator) and the token store's roles (orphan and period set
// on a role need no sudo to mint through).

const (
	// OpenBAOKubernetesAuthPath is where the auth method is mounted; logins
	// go to auth/<path>/login.
	OpenBAOKubernetesAuthPath = "kubernetes"

	// OpenBAOKubernetesHost is the API server as every pod in every cluster
	// reaches it. Deliberately the in-cluster name and never the address of
	// the cluster the vault was created in: this value is stored in the
	// vault's own data and comes back with a restore, so it must be true on
	// whatever cluster the restored vault runs in.
	OpenBAOKubernetesHost = "https://kubernetes.default.svc:443"

	// OpenBAOOperatorSessionTTL bounds the operator's sign-in: a session is
	// created at the start of a pass and revoked at its end, and if the
	// revoke is ever missed the token still dies here.
	OpenBAOOperatorSessionTTL = 10 * time.Minute

	// OpenBAOControlPlaneTokenPeriod is the control plane's token period: a
	// periodic token never expires while renewed within its period, and the
	// operator renews it well inside this window on every pass it runs.
	// Seven days is long enough that an operator down over a weekend does
	// not take the platform's secrets with it, and short enough that a token
	// nobody renews is gone within a week of the operator stopping.
	OpenBAOControlPlaneTokenPeriod = 7 * 24 * time.Hour

	// The control plane's token Secret's data keys: the token the control
	// plane presents, and its accessor -- the public handle the operator
	// looks the token up, renews it, and revokes it by without holding it.
	OpenBAOTokenSecretTokenKey    = "token"
	OpenBAOTokenSecretAccessorKey = "accessor"

	// OpenBAOTokenSecretAnnotation carries the plain-language note on the
	// control plane's token Secret (OpenBAOTokenSecretNote).
	OpenBAOTokenSecretAnnotation = "planton.ai/openbao-token"
)

// OpenBAOOperatorRoleName names both the operator's Kubernetes auth role and
// its policy: "{crName}-operator".
func OpenBAOOperatorRoleName(crName string) string { return fmt.Sprintf("%s-operator", crName) }

// OpenBAOControlPlaneRoleName names both the control plane's token role and
// its policy: "{crName}-control-plane".
func OpenBAOControlPlaneRoleName(crName string) string {
	return fmt.Sprintf("%s-control-plane", crName)
}

// OpenBAOTokenSecretName is the operator-owned Secret the control plane's
// token lives in: "{crName}-openbao-token". The one credential Secret in
// this operator that is rewritten rather than created once: its content is
// issued by the vault and re-issued by design (a restored platform, a
// revoked token), and the control plane is rolled to pick the new one up.
func OpenBAOTokenSecretName(crName string) string { return fmt.Sprintf("%s-openbao-token", crName) }

// OpenBAOServiceAccountName is the vault server's own ServiceAccount, named
// by the chart after the release: "{crName}-openbao". It is the identity
// the auth method reviews tokens with, so it carries system:auth-delegator.
func OpenBAOServiceAccountName(crName string) string { return openbaoReleaseName(crName) }

// OpenBAOControlPlanePolicy is what the control plane may do: its two
// engines, every capability but sudo, and nothing under sys/ or auth/.
// Scoped to the mounts and not to today's verbs on purpose -- the control
// plane is the sole tenant of both engines, and it releases independently
// of this operator; a verb-by-verb policy would make every new use of the
// vault in the platform a forced operator release with a 403 in between,
// for no boundary gain. Path patterns the control plane uses under these
// mounts include a prefix the adopter may rename, so nothing narrower than
// the mount would be true either.
func OpenBAOControlPlanePolicy() string {
	return `# The control plane's policy: the two engines the platform owns, nothing else.
# No sys/ and no auth/ -- provisioning is the operator's, through its own
# session, never through this token.
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list", "patch"]
}
path "transit/*" {
  capabilities = ["create", "read", "update", "delete", "list", "patch"]
}
`
}

// OpenBAOOperatorPolicy is what the operator's session may do: the
// platform's own objects in the vault and nothing the platform stores. It
// can enable the two engines, keep the two policies and the two roles
// current, mint the control plane's token through its role, and look up,
// renew, and revoke tokens by accessor. It does not include the engines
// themselves. Stated honestly: a policy that can rewrite policies is an
// administrator's, and it could widen itself -- what changes is that this
// power lives in a session created for one pass and revoked at its end,
// never in a Secret. Enabling the auth method (sys/auth) needs sudo and is
// deliberately absent: that is done with the root token at initialization.
func OpenBAOOperatorPolicy(crName string) string {
	operator := OpenBAOOperatorRoleName(crName)
	controlPlane := OpenBAOControlPlaneRoleName(crName)
	var b strings.Builder
	b.WriteString("# The operator's policy: provision the platform's vault, read none of its secrets.\n")
	b.WriteString("path \"sys/mounts\" {\n  capabilities = [\"read\"]\n}\n")
	b.WriteString("path \"sys/mounts/secret\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n")
	b.WriteString("path \"sys/mounts/transit\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n")
	fmt.Fprintf(&b, "path \"sys/policies/acl/%s\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n", operator)
	fmt.Fprintf(&b, "path \"sys/policies/acl/%s\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n", controlPlane)
	fmt.Fprintf(&b, "path \"auth/%s/config\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n", OpenBAOKubernetesAuthPath)
	fmt.Fprintf(&b, "path \"auth/%s/role/%s\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n", OpenBAOKubernetesAuthPath, operator)
	fmt.Fprintf(&b, "path \"auth/token/roles/%s\" {\n  capabilities = [\"create\", \"read\", \"update\"]\n}\n", controlPlane)
	fmt.Fprintf(&b, "path \"auth/token/create/%s\" {\n  capabilities = [\"create\", \"update\"]\n}\n", controlPlane)
	b.WriteString("path \"auth/token/lookup-accessor\" {\n  capabilities = [\"update\"]\n}\n")
	b.WriteString("path \"auth/token/renew-accessor\" {\n  capabilities = [\"update\"]\n}\n")
	b.WriteString("path \"auth/token/revoke-accessor\" {\n  capabilities = [\"update\"]\n}\n")
	return b.String()
}

// OpenBAOTokenSecretNote is the plain-language note on the control plane's
// token Secret: what it is, where it came from, what keeps it alive, and
// what touching it does.
func OpenBAOTokenSecretNote(crName string) string {
	return fmt.Sprintf(
		"The control plane's token for the bundled secrets manager (OpenBAO release %s), minted by the operator through the token role %s and limited to the secret/ and transit/ engines. "+
			"The operator keeps it alive by its accessor on every pass and mints a new one when this Secret is missing or the token is no longer valid (after a restore, for example), then restarts the control plane so it picks the new token up. "+
			"Nothing here is the vault's root token; that stays in the init Secret as break-glass. Editing or deleting this Secret only makes the operator mint again.",
		openbaoReleaseName(crName), OpenBAOControlPlaneRoleName(crName))
}

// OpenBAOAuthDelegatorClusterRoleBindingName is the cluster-scoped name of
// the vault's TokenReview grant: "{namespace}-{crName}-openbao-auth-delegator".
// The namespace is in the name for the reason the control plane's
// token-reviewer grant carries it: the object is cluster-scoped while
// platforms are namespaced, and same-named platforms in two namespaces must
// never share one binding.
func OpenBAOAuthDelegatorClusterRoleBindingName(namespace, crName string) string {
	return fmt.Sprintf("%s-%s-openbao-auth-delegator", namespace, crName)
}

// OpenBAOAuthDelegatorClusterRoleBinding grants the vault server's
// ServiceAccount system:auth-delegator -- the cluster's own role for
// reviewing tokens and access, which the Kubernetes auth method needs to
// verify a login's ServiceAccount token with the API server. It is applied
// as a typed object beside the control plane's token-reviewer grant and
// never through the chart's authDelegator switch: the chart's binding would
// be cluster-scoped output of a namespaced platform, which the apply seam
// refuses, and it could not carry the platform's UID. Like every
// cluster-scoped satellite it is labeled with that UID and collected by the
// janitor once the platform is gone; until then it is inert, because its
// only subject dies with the platform's namespace.
func OpenBAOAuthDelegatorClusterRoleBinding(namespace, crName string, platformUID types.UID) *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app.kubernetes.io/name":       "openbao",
		"app.kubernetes.io/instance":   crName,
		"app.kubernetes.io/managed-by": ManagedByLabel,
		"app.kubernetes.io/component":  "secrets-manager",
	}
	if platformUID != "" {
		labels[PlatformUIDLabel] = string(platformUID)
	}
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRoleBinding"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   OpenBAOAuthDelegatorClusterRoleBindingName(namespace, crName),
			Labels: labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:auth-delegator",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      OpenBAOServiceAccountName(crName),
			Namespace: namespace,
		}},
	}
}
