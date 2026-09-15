package component

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/bootstrap"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The operator's own sign-in. Every pass that touches the vault begins by
// signing in through the Kubernetes auth method as the operator's own
// ServiceAccount and ends by revoking that session; nothing about the
// operator's access is ever stored. The root token is not the operator's
// working credential: it is used for exactly one thing, repairing the
// operator's sign-in when the vault does not know the operator's current
// identity -- the operator reinstalled under another name or namespace, a
// role somebody edited -- and only when the init Secret still holds it.
//
// Why per pass and not cached: the operator's components are constructed
// fresh on every reconcile and keep nothing between passes (they decide
// from the live cluster every time), so a cached session would be the first
// stateful component; and the session must be a service token, because a
// batch token cannot create tokens and this session mints the control
// plane's. The cost is two small writes to the platform's own database per
// pass -- a token created, then revoked -- and it is stated here so nobody
// later wonders.

// vaultSession is one signed-in pass: the token to present and how to end it.
type vaultSession struct {
	token   string
	apiAddr string
	http    *http.Client
	// identity is who signed in -- what the operator's role is bound to and
	// what the access arrangement is re-asserted for.
	identity operatorIdentity
	// login is true for a session obtained by signing in (revoked at close);
	// false for the root token used at initialization, which is never
	// revoked -- it is the break-glass the init Secret keeps.
	login bool
}

// close ends the session. A revoke that fails is logged, never returned:
// the pass's work is done, and the token dies at its TTL regardless.
func (s *vaultSession) close(ctx context.Context) {
	if s == nil || !s.login {
		return
	}
	if err := bootstrap.RevokeSelf(ctx, s.http, s.apiAddr, s.token); err != nil {
		logf.FromContext(ctx).Info("OpenBAO session token could not be revoked; it expires on its own", "error", err.Error())
	}
}

// rootSession wraps the root token for the initialization pass.
func rootSession(rootToken, apiAddr string, httpClient *http.Client) *vaultSession {
	return &vaultSession{token: rootToken, apiAddr: apiAddr, http: httpClient}
}

// desiredKubernetesAuthConfig is the auth method's configuration as the
// operator writes it: the in-cluster API server and nothing else, so the
// method reads the vault pod's own CA and token and works on any cluster
// the vault is restored into.
func desiredKubernetesAuthConfig() bootstrap.KubernetesAuthConfig {
	return bootstrap.KubernetesAuthConfig{Host: resources.OpenBAOKubernetesHost}
}

// desiredOperatorRole binds the operator's identity to its policy for a
// session-length token.
func desiredOperatorRole(crName string, id operatorIdentity) bootstrap.KubernetesAuthRole {
	return bootstrap.KubernetesAuthRole{
		ServiceAccountNames:      []string{id.ServiceAccount},
		ServiceAccountNamespaces: []string{id.Namespace},
		Policies:                 []string{resources.OpenBAOOperatorRoleName(crName)},
		TTL:                      resources.OpenBAOOperatorSessionTTL,
		MaxTTL:                   resources.OpenBAOOperatorSessionTTL,
	}
}

// desiredControlPlaneTokenRole is the shape the control plane's token takes:
// orphan (it must not die with the operator's session), periodic (it never
// expires while renewed), renewable, limited to the control plane's policy.
func desiredControlPlaneTokenRole(crName string) bootstrap.TokenRole {
	return bootstrap.TokenRole{
		AllowedPolicies: []string{resources.OpenBAOControlPlaneRoleName(crName)},
		Orphan:          true,
		Renewable:       true,
		Period:          resources.OpenBAOControlPlaneTokenPeriod,
		TokenType:       "service",
	}
}

// configureVaultAccess writes the whole access arrangement with a token that
// may do everything -- the root token, at initialization and in the repair
// arm: the auth method enabled and configured, both policies, the operator's
// role, the control plane's token role. Idempotent; every write is a
// replace.
func configureVaultAccess(ctx context.Context, s *vaultSession, crName string, id operatorIdentity) error {
	if err := bootstrap.EnableKubernetesAuth(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath); err != nil {
		return err
	}
	if err := bootstrap.WriteKubernetesAuthConfig(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath, desiredKubernetesAuthConfig()); err != nil {
		return err
	}
	if err := bootstrap.WritePolicy(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOOperatorRoleName(crName), resources.OpenBAOOperatorPolicy(crName)); err != nil {
		return err
	}
	if err := bootstrap.WritePolicy(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOControlPlaneRoleName(crName), resources.OpenBAOControlPlanePolicy()); err != nil {
		return err
	}
	if err := bootstrap.WriteKubernetesAuthRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath, resources.OpenBAOOperatorRoleName(crName), desiredOperatorRole(crName, id)); err != nil {
		return err
	}
	return bootstrap.WriteTokenRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOControlPlaneRoleName(crName), desiredControlPlaneTokenRole(crName))
}

// ensureVaultAccess re-asserts the access arrangement with the operator's own
// session, by read, compare, and write only on drift -- so an operator
// upgrade that changes a policy's text or a role's shape lands on the next
// pass without touching the root token, and the steady state writes nothing.
func ensureVaultAccess(ctx context.Context, s *vaultSession, crName string, id operatorIdentity) error {
	log := logf.FromContext(ctx)

	for name, want := range map[string]string{
		resources.OpenBAOOperatorRoleName(crName):     resources.OpenBAOOperatorPolicy(crName),
		resources.OpenBAOControlPlaneRoleName(crName): resources.OpenBAOControlPlanePolicy(),
	} {
		got, found, err := bootstrap.ReadPolicy(ctx, s.http, s.apiAddr, s.token, name)
		if err != nil {
			return err
		}
		if !found || got != want {
			log.Info("Writing OpenBAO policy", "policy", name, "existed", found)
			if err := bootstrap.WritePolicy(ctx, s.http, s.apiAddr, s.token, name, want); err != nil {
				return err
			}
		}
	}

	wantCfg := desiredKubernetesAuthConfig()
	gotCfg, found, err := bootstrap.ReadKubernetesAuthConfig(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath)
	if err != nil {
		return err
	}
	if !found || gotCfg.Host != wantCfg.Host {
		log.Info("Writing OpenBAO kubernetes auth config", "existed", found)
		if err := bootstrap.WriteKubernetesAuthConfig(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath, wantCfg); err != nil {
			return err
		}
	}

	wantRole := desiredOperatorRole(crName, id)
	gotRole, found, err := bootstrap.ReadKubernetesAuthRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath, resources.OpenBAOOperatorRoleName(crName))
	if err != nil {
		return err
	}
	if !found || !slices.Equal(gotRole.ServiceAccountNames, wantRole.ServiceAccountNames) ||
		!slices.Equal(gotRole.ServiceAccountNamespaces, wantRole.ServiceAccountNamespaces) ||
		!slices.Equal(gotRole.Policies, wantRole.Policies) ||
		gotRole.TTL != wantRole.TTL || gotRole.MaxTTL != wantRole.MaxTTL {
		log.Info("Writing OpenBAO kubernetes auth role", "role", resources.OpenBAOOperatorRoleName(crName), "existed", found)
		if err := bootstrap.WriteKubernetesAuthRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOKubernetesAuthPath, resources.OpenBAOOperatorRoleName(crName), wantRole); err != nil {
			return err
		}
	}

	wantToken := desiredControlPlaneTokenRole(crName)
	gotToken, found, err := bootstrap.ReadTokenRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOControlPlaneRoleName(crName))
	if err != nil {
		return err
	}
	if !found || !slices.Equal(gotToken.AllowedPolicies, wantToken.AllowedPolicies) ||
		gotToken.Orphan != wantToken.Orphan || gotToken.Renewable != wantToken.Renewable ||
		gotToken.Period != wantToken.Period || gotToken.TokenType != wantToken.TokenType {
		log.Info("Writing OpenBAO token role", "role", resources.OpenBAOControlPlaneRoleName(crName), "existed", found)
		if err := bootstrap.WriteTokenRole(ctx, s.http, s.apiAddr, s.token, resources.OpenBAOControlPlaneRoleName(crName), wantToken); err != nil {
			return err
		}
	}
	return nil
}

// signIn opens the pass's session: the operator's identity read from its
// pod, a login through the auth method, and -- when the vault refuses the
// identity and the init Secret still holds the root token -- one repair of
// the access arrangement followed by a second login. Returns the session,
// or the Result a person reads when no session could be opened, or an
// error for what is transient (the network, the API server).
//
// rootToken is the break-glass available to this pass: the token /sys/init
// just returned (initialization), the init Secret's (a steady pass with the
// Secret present), or "" (the Secret is gone). initSecretName is only for
// the sentence.
func signIn(ctx context.Context, planton *v1.PlantonPlatform, apiAddr string, httpClient *http.Client, rootToken, initSecretName string) (*vaultSession, *Result, error) {
	log := logf.FromContext(ctx).WithValues("component", "openbao")

	id, err := readOperatorIdentity()
	if err != nil {
		return nil, &Result{
			Ready:   false,
			Reason:  v1.ComponentReasonReconcileFailed,
			Message: err.Error(),
		}, nil
	}

	role := resources.OpenBAOOperatorRoleName(planton.Name)
	tok, err := bootstrap.KubernetesLogin(ctx, httpClient, apiAddr, resources.OpenBAOKubernetesAuthPath, role, id.JWT)
	if err == nil {
		return &vaultSession{token: tok.ClientToken, apiAddr: apiAddr, http: httpClient, identity: *id, login: true}, nil, nil
	}
	refusal, isRefusal := bootstrap.AsAPIError(err)
	if !isRefusal {
		return nil, nil, err // the network or the server, not the identity
	}

	if rootToken == "" {
		return nil, &Result{
			Ready:   false,
			Reason:  v1.ComponentReasonConfigurationRefused,
			Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
			Message: loginRefusedMessage(planton, *id, initSecretName, refusal),
		}, nil
	}

	log.Info("OpenBAO refused the operator's sign-in; repairing the access arrangement with the root token", "identity", id.String(), "reason", refusal.Error())
	root := rootSession(rootToken, apiAddr, httpClient)
	if err := configureVaultAccess(ctx, root, planton.Name, *id); err != nil {
		if apiErr, ok := bootstrap.AsAPIError(err); ok {
			return nil, &Result{
				Ready:   false,
				Reason:  v1.ComponentReasonConfigurationRefused,
				Object:  &v1.ComponentObjectReference{Kind: "Secret", Name: initSecretName},
				Message: repairRefusedMessage(planton, *id, initSecretName, refusal, apiErr),
			}, nil
		}
		return nil, nil, err
	}
	tok, err = bootstrap.KubernetesLogin(ctx, httpClient, apiAddr, resources.OpenBAOKubernetesAuthPath, role, id.JWT)
	if err != nil {
		if apiErr, ok := bootstrap.AsAPIError(err); ok {
			return nil, &Result{
				Ready:   false,
				Reason:  v1.ComponentReasonConfigurationRefused,
				Object:  &v1.ComponentObjectReference{Kind: "ClusterRoleBinding", Name: resources.OpenBAOAuthDelegatorClusterRoleBindingName(planton.Namespace, planton.Name)},
				Message: loginStillRefusedMessage(planton, *id, apiErr),
			}, nil
		}
		return nil, nil, err
	}
	log.Info("OpenBAO accepted the operator's sign-in after repair", "identity", id.String())
	return &vaultSession{token: tok.ClientToken, apiAddr: apiAddr, http: httpClient, identity: *id, login: true}, nil, nil
}

// loginRefusedMessage: the vault refused the operator's identity and there
// is no root token to repair with. Two ways back, one per cause.
func loginRefusedMessage(planton *v1.PlantonPlatform, id operatorIdentity, initSecret string, refusal *bootstrap.APIError) string {
	return fmt.Sprintf(
		"The vault refused the operator's sign-in as %s (%s), and Secret %s in namespace %s does not exist, so the operator holds no root token to repair its access with. "+
			"Either the vault knows the operator under another identity (it was initialized by an operator installed under a different name or namespace) -- install the operator under that identity, or recreate Secret %s from the copy you kept and the operator repairs its own access on the next pass; "+
			"or the vault cannot verify the operator's token with the API server -- check that ClusterRoleBinding %s exists and grants system:auth-delegator to ServiceAccount %s.",
		id.String(), refusal.Error(), initSecret, planton.Namespace,
		initSecret,
		resources.OpenBAOAuthDelegatorClusterRoleBindingName(planton.Namespace, planton.Name), resources.OpenBAOServiceAccountName(planton.Name))
}

// repairRefusedMessage: the vault refused the identity, the init Secret is
// present, and the root token in it no longer works.
func repairRefusedMessage(planton *v1.PlantonPlatform, id operatorIdentity, initSecret string, refusal, repair *bootstrap.APIError) string {
	return fmt.Sprintf(
		"The vault refused the operator's sign-in as %s (%s), and the root token in Secret %s (namespace %s) could not repair its access (%s) -- the token was revoked or replaced. "+
			"Generate a new root token with the vault's recovery quorum (the keys in that Secret) and write it back under %s, or install the operator under the identity the vault already knows.",
		id.String(), refusal.Error(), initSecret, planton.Namespace, repair.Error(),
		resources.OpenBAOInitSecretRootTokenKey)
}

// loginStillRefusedMessage: the access arrangement was just rewritten for
// this identity and the login still failed -- the role is right, so what is
// left is the vault's ability to verify the token with the API server.
func loginStillRefusedMessage(planton *v1.PlantonPlatform, id operatorIdentity, refusal *bootstrap.APIError) string {
	return fmt.Sprintf(
		"The vault still refuses the operator's sign-in as %s after its role was rewritten for that identity (%s). "+
			"The role is right, so the vault most likely cannot verify the operator's token with the API server: check that ClusterRoleBinding %s grants system:auth-delegator to ServiceAccount %s in namespace %s, and that the vault pod can reach %s.",
		id.String(), refusal.Error(),
		resources.OpenBAOAuthDelegatorClusterRoleBindingName(planton.Namespace, planton.Name), resources.OpenBAOServiceAccountName(planton.Name), planton.Namespace,
		resources.OpenBAOKubernetesHost)
}
