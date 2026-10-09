// GKE bearer-token minting: an OAuth2 access token from a Google service-account key
// or, when no key is supplied, from the ambient Google credential chain of the process.
package kubetoken

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
)

// cloudPlatformScope is the OAuth2 scope GKE API servers accept for IAM-authenticated
// requests (the same scope gcloud requests for cluster access).
const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// GoogleOAuthAccessTokenEnvVar carries a pre-minted Google access token into the
// process. It is the same variable the Terraform/OpenTofu google providers read. Two
// callers set it: the kubeconfig builder, on the exec entry of a GKE config whose
// access_token is set (a keyless connection's token, exchanged before the engine ran),
// and a Planton runner that exports it when a runner-mode GCP connection names a
// gcloud configuration (the per-connection ambient-identity selector). Checked before
// ADC so the job authenticates as the identity the connection records, not the
// machine default.
const GoogleOAuthAccessTokenEnvVar = "GOOGLE_OAUTH_ACCESS_TOKEN"

// GkeTokenOptions carries the Google service-account key that signs the token request.
type GkeTokenOptions struct {
	// ServiceAccountKeyJSON is the raw key-file JSON downloaded from the GCP console
	// or minted via `gcloud iam service-accounts keys create`.
	//
	// Optional: empty selects the ambient Google credential chain --
	// GOOGLE_OAUTH_ACCESS_TOKEN when present, else Application Default Credentials
	// (GOOGLE_APPLICATION_CREDENTIALS, Workload Identity, GCE metadata, or
	// `gcloud auth application-default login`). Same contract as the EKS static-key
	// fields and the AKS client_secret.
	ServiceAccountKeyJSON string

	// TokenURL overrides Google's token endpoint. Empty in production; tests point it
	// at a local fake so token minting is verifiable fully offline.
	TokenURL string
}

// MintGkeToken exchanges the service-account key (or the ambient chain) for a
// short-lived OAuth2 access token (Google issues ~1h tokens; the expiry comes from
// the token response, never assumed).
func MintGkeToken(ctx context.Context, opts GkeTokenOptions) (Token, error) {
	if opts.ServiceAccountKeyJSON == "" {
		return mintGkeAmbientToken(ctx)
	}

	cfg, err := google.JWTConfigFromJSON([]byte(opts.ServiceAccountKeyJSON), cloudPlatformScope)
	if err != nil {
		return Token{}, errors.Wrap(err, "parsing Google service-account key")
	}
	if opts.TokenURL != "" {
		cfg.TokenURL = opts.TokenURL
	}

	token, err := cfg.TokenSource(ctx).Token()
	if err != nil {
		return Token{}, errors.Wrap(err, "exchanging service-account JWT for an access token")
	}

	return Token{Value: token.AccessToken, ExpiresAt: token.Expiry}, nil
}

// mintGkeAmbientToken resolves the ambient Google credential chain: a pre-minted
// GOOGLE_OAUTH_ACCESS_TOKEN first, then Application Default Credentials.
func mintGkeAmbientToken(ctx context.Context) (Token, error) {
	if envToken := os.Getenv(GoogleOAuthAccessTokenEnvVar); envToken != "" {
		return Token{Value: envToken, ExpiresAt: time.Now().Add(suppliedTokenTTL)}, nil
	}

	creds, err := google.FindDefaultCredentials(ctx, cloudPlatformScope)
	if err != nil {
		return Token{}, errors.Wrap(err, "resolving the ambient Google credential chain "+
			"(no service-account key on the connection, no GOOGLE_OAUTH_ACCESS_TOKEN in the "+
			"environment, and Application Default Credentials are unavailable)")
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return Token{}, errors.Wrap(err, "minting an access token from Application Default Credentials")
	}

	return Token{Value: token.AccessToken, ExpiresAt: token.Expiry}, nil
}
