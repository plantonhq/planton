// AKS bearer-token minting: a Microsoft Entra ID access token issued for the AKS
// AAD server application.
package kubetoken

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/pkg/errors"
)

// aksServerAppID is the application ID of the AKS AAD Server -- a first-party Azure
// application that is the same in every environment. An Entra-integrated AKS API
// server only honors bearer tokens whose audience is this application, so every
// token minted here is requested for it (kubelogin's --server-id equivalent).
const aksServerAppID = "6dae42f8-4368-4678-94ff-3960e28e3630"

// aksTokenScope is the OAuth2 scope form of the audience: Entra's client-credentials
// flow takes "<application>/.default", meaning "every app-role the principal already
// holds on that application".
const aksTokenScope = aksServerAppID + "/.default"

// AksTokenOptions carries the credential behind the token: an AccessToken a caller
// already minted (used as is), or the Entra service-principal credential that requests
// one. When both AccessToken and ClientSecret are empty the ambient Azure credential
// chain of the process authenticates instead (environment variables, managed
// identity, Azure CLI login).
type AksTokenOptions struct {
	// TenantID is the Entra tenant the service principal lives in. Required whenever
	// ClientSecret is set; in ambient mode it optionally steers the chain to a
	// specific tenant on multi-tenant hosts (empty keeps the chain's own default).
	TenantID string

	// ClientID identifies the service principal. Required whenever ClientSecret is set.
	ClientID string

	// ClientSecret authenticates the service principal. Empty selects ambient mode.
	ClientSecret string

	// AccessToken is an Entra access token for the AKS server application that the
	// caller already minted -- a Planton runner exchanges a keyless Azure connection for
	// one before the engine starts. It is returned as is; setting it together with
	// ClientSecret is refused, since the two name different credentials.
	AccessToken string

	// Transport overrides the HTTP transport of the credential. Empty in production;
	// tests point it at a local fake so token minting is verifiable fully offline
	// (setting it also skips Entra instance discovery, which would otherwise need
	// the network before the token request is ever sent).
	Transport policy.Transporter
}

// MintAksToken exchanges the service-principal credential (or the ambient chain) for
// a short-lived Entra access token scoped to the AKS AAD server application. The
// expiry comes from the token response, never assumed -- except for a supplied
// AccessToken, which carries no expiry of its own (see suppliedTokenTTL).
func MintAksToken(ctx context.Context, opts AksTokenOptions) (Token, error) {
	if opts.AccessToken != "" {
		if opts.ClientSecret != "" {
			return Token{}, errors.New("both a client secret and an access token were supplied for the AKS " +
				"token; exactly one credential may be given")
		}
		return Token{Value: opts.AccessToken, ExpiresAt: time.Now().Add(suppliedTokenTTL)}, nil
	}

	cred, err := newAksCredential(opts)
	if err != nil {
		return Token{}, err
	}
	return MintAksTokenWithCredential(ctx, cred)
}

// MintAksTokenWithCredential mints the AKS token from an already-resolved Entra
// credential -- the seam for callers that obtain their credential elsewhere (the
// CloudOps Azure client resolves a delegate connection per its own auth mode into
// an azcore.TokenCredential). Keeping the request here keeps the AKS
// server-application scope private to this one minting seam.
func MintAksTokenWithCredential(ctx context.Context, cred azcore.TokenCredential) (Token, error) {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{aksTokenScope}})
	if err != nil {
		return Token{}, errors.Wrap(err, "requesting an Entra access token for the AKS server application")
	}

	return Token{Value: token.Token, ExpiresAt: token.ExpiresOn}, nil
}

// newAksCredential picks the credential source the options describe: an explicit
// service-principal secret, or the process's ambient Azure credential chain.
func newAksCredential(opts AksTokenOptions) (azcore.TokenCredential, error) {
	clientOptions := azcore.ClientOptions{Transport: opts.Transport}

	if opts.ClientSecret == "" {
		cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			ClientOptions: clientOptions,
			TenantID:      opts.TenantID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "building the ambient Azure credential chain")
		}
		return cred, nil
	}

	if opts.TenantID == "" || opts.ClientID == "" {
		return nil, errors.New("tenant ID and client ID are required to mint an AKS token from a client secret")
	}

	cred, err := azidentity.NewClientSecretCredential(opts.TenantID, opts.ClientID, opts.ClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: clientOptions,
			// A fake transport cannot serve Entra's instance-discovery metadata, and
			// production never sets one -- so the override doubles as the offline switch.
			DisableInstanceDiscovery: opts.Transport != nil,
		})
	if err != nil {
		return nil, errors.Wrap(err, "building the client-secret credential")
	}
	return cred, nil
}
