// Package kubetoken mints short-lived Kubernetes API-server bearer tokens in-process
// for the managed-cluster providers whose control planes accept cloud-IAM credentials
// (EKS presigned-STS tokens, GKE OAuth2 access tokens, AKS Entra tokens). It is the single minting seam
// shared by the ExecCredential command the deploy engines re-invoke and by in-process
// Kubernetes clients; no cloud CLI is ever shelled out to.
package kubetoken

import "time"

// Token is a short-lived bearer credential for a Kubernetes API server, paired with
// the instant it stops being honored so callers can report an honest expiry (the
// ExecCredential protocol uses it to know when to re-invoke the credential command).
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// suppliedTokenTTL is the expiry reported for a token the caller already minted (a
// GKE token from GOOGLE_OAUTH_ACCESS_TOKEN, an AKS access token), which carries no
// expiry metadata of its own. The ExecCredential protocol requires an expiration
// timestamp (a zero time would read as already expired), so report a cadence well
// inside the real ~1h token lifetimes; every re-invoke re-reads the same input, so this
// value only sets how often client-go asks again.
const suppliedTokenTTL = 10 * time.Minute
