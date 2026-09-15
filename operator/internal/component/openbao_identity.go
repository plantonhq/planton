package component

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/plantonhq/planton/operator/internal/singleton"
)

// The operator signs in to the vault as itself, so it has to know who it is:
// the ServiceAccount name and namespace the vault's auth role is bound to.
// Neither is a constant anywhere. The Helm chart names the ServiceAccount
// after the release and puts it in the release namespace, kustomize names
// it differently again, and the manager pod carries no environment for
// either. The one thing every pod has is the ServiceAccount token the
// kubelet mounted, whose subject is exactly the identity the API server
// will report when the vault reviews it -- so the operator reads its
// identity from there, by construction the same identity the vault sees.

// operatorIdentity is the operator's own ServiceAccount, as the vault will
// see it, plus the token that proves it.
type operatorIdentity struct {
	Namespace      string
	ServiceAccount string
	// JWT is the mounted ServiceAccount token, presented at login.
	JWT string
}

// String is the identity as a sentence names it: "namespace/account".
func (id operatorIdentity) String() string {
	return id.Namespace + "/" + id.ServiceAccount
}

// serviceAccountTokenFile is where the kubelet mounts the pod's token.
const serviceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// readOperatorIdentity is the seam the component reads its identity
// through; tests replace it. In a pod it is readOperatorIdentityFromPod.
var readOperatorIdentity = readOperatorIdentityFromPod

// readOperatorIdentityFromPod reads the mounted token and takes the identity
// from its subject claim, "system:serviceaccount:<namespace>:<name>". The
// claim is read without verifying the signature: the operator is not
// deciding whether to trust the token, only reading what the cluster wrote
// into it; the vault is what verifies it, with the cluster. The namespace
// is cross-checked against the mounted namespace file, which every pod
// also has.
func readOperatorIdentityFromPod() (*operatorIdentity, error) {
	raw, err := os.ReadFile(serviceAccountTokenFile)
	if err != nil {
		return nil, fmt.Errorf("the operator's own ServiceAccount token is not mounted at %s, so it cannot sign in to the vault as itself -- the operator must run in a pod to manage a platform's vault: %w", serviceAccountTokenFile, err)
	}
	id, err := identityFromServiceAccountToken(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, err
	}
	if ns, ok := singleton.OwnNamespace(); ok && ns != id.Namespace {
		return nil, fmt.Errorf("the operator's ServiceAccount token names namespace %q but the pod's namespace file says %q; the two are written by the same kubelet and cannot honestly differ -- the pod's mounts are not the pod's own", id.Namespace, ns)
	}
	return id, nil
}

// identityFromServiceAccountToken parses a ServiceAccount token's subject.
func identityFromServiceAccountToken(jwt string) (*operatorIdentity, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("the operator's ServiceAccount token is not a JWT (expected three dot-separated parts)")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding the operator's ServiceAccount token claims: %w", err)
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("reading the operator's ServiceAccount token claims: %w", err)
	}
	const prefix = "system:serviceaccount:"
	if !strings.HasPrefix(claims.Subject, prefix) {
		return nil, fmt.Errorf("the operator's token subject %q is not a ServiceAccount (expected %s<namespace>:<name>)", claims.Subject, prefix)
	}
	rest := strings.TrimPrefix(claims.Subject, prefix)
	ns, name, ok := strings.Cut(rest, ":")
	if !ok || ns == "" || name == "" {
		return nil, fmt.Errorf("the operator's token subject %q does not name a namespace and an account", claims.Subject)
	}
	return &operatorIdentity{Namespace: ns, ServiceAccount: name, JWT: jwt}, nil
}
