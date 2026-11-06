package module

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// Agent Substrate keeps every certificate authority and the actor token
// signing key as a "pool": JSON naming one or more keys and the one active
// for signing, so a later rotation can add a key beside the old one. The
// module generates each key and certificate with the TLS provider (both
// engines run the same one) and writes the pool in upstream's exact wire
// format: Go's encoding/json of upstream's structs, so byte fields are
// standard base64 of DER (internal/localca and internal/localjwtauthority
// Marshal). A gated test parses both engines' pools with upstream's own
// Unmarshal.
//
// A PEM block's body is exactly that base64 of its DER, wrapped at 64
// columns, so a pool field is the PEM with its armour lines and line breaks
// removed. The Terraform twin does the same string surgery.

// caPoolID is the ID of the single CA in a generated pool, upstream's
// poolKeyID.
const caPoolID = "1"

// caValidityHours is how long a generated CA certificate lasts. Upstream's
// installer issues 365 days and never renews; every connection that chains
// to an expired root fails, and the TLS provider would replace an expired
// root on the next apply, rotating trust out from under every issued
// certificate. Ten years, the kubeadm convention for self-managed roots,
// keeps both from happening within any install's life; the leaves under it
// are short-lived (24 hours for pod certificates, 15 minutes for egress
// interception).
const caValidityHours = 87600

// pemBase64 is the base64 DER inside a PEM block.
func pemBase64(pem string) string {
	var b strings.Builder
	for _, line := range strings.Split(pem, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-----") {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

// caPoolJSON is a localca pool holding one CA, active for signing. The
// field order is upstream's struct order, which encoding/json keeps.
func caPoolJSON(privateKeyPKCS8PEM, certificatePEM string) string {
	pool := struct {
		CAs []struct {
			ID                 string
			SigningKeyPKCS8    string
			RootCertificateDER string
		}
		ActiveForSigning string
	}{ActiveForSigning: caPoolID}
	pool.CAs = append(pool.CAs, struct {
		ID                 string
		SigningKeyPKCS8    string
		RootCertificateDER string
	}{ID: caPoolID, SigningKeyPKCS8: pemBase64(privateKeyPKCS8PEM), RootCertificateDER: pemBase64(certificatePEM)})
	raw, _ := json.Marshal(pool)
	return string(raw)
}

// jwtKeyID names the actor token signing key: base64url (no padding) of the
// SHA-256 of its public key's PEM text. Upstream names a key by hashing the
// public key's DER, which Terraform cannot (it hashes only text); upstream's
// pool accepts any ID, and the published key set carries whatever ID the
// pool names, so both engines derive this one, of the same shape, from the
// text they share.
func jwtKeyID(publicKeyPEM string) string {
	sum := sha256.Sum256([]byte(publicKeyPEM))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// jwtPoolJSON is a localjwtauthority pool holding one signing key, active.
func jwtPoolJSON(algorithm, privateKeyPKCS8PEM, publicKeyPEM string) string {
	id := jwtKeyID(publicKeyPEM)
	pool := struct {
		Authorities []struct {
			ID              string
			Algorithm       string
			SigningKeyPKCS8 string
		}
		ActiveForSigning string
	}{ActiveForSigning: id}
	pool.Authorities = append(pool.Authorities, struct {
		ID              string
		Algorithm       string
		SigningKeyPKCS8 string
	}{ID: id, Algorithm: algorithm, SigningKeyPKCS8: pemBase64(privateKeyPKCS8PEM)})
	raw, _ := json.Marshal(pool)
	return string(raw)
}

// caPool is one generated certificate authority: its Secret's name and
// namespace, the key algorithm upstream uses for it, and the common name the
// module gives it (upstream's is random; nothing reads it, and a name tells
// an operator which CA a certificate chains to).
type caPool struct {
	secret     string
	namespace  string
	algorithm  string // ED25519 | ECDSA
	commonName string
	// trustBundle and signer name the ClusterTrustBundle the pod-certificate
	// controller publishes this pool's root under (trustBundles); empty for
	// the pools it does not sign from.
	trustBundle string
	signer      string
}

// The one label the pod-certificate controller gives each ClusterTrustBundle
// it publishes, which it restores on any bundle that differs.
const (
	trustBundleLabelKey   = "podcert.ate.dev/canarying"
	trustBundleLabelValue = "live"
)

// trustBundleAPIVersion is the ClusterTrustBundle API a cluster serves: v1
// from Kubernetes 1.37, v1beta1 on the releases before it that serve the
// pod-certificate APIs (GKE 1.36). Terraform twin: locals.tf
// trust_bundle_api_version.
func trustBundleAPIVersion(major, minor int) string {
	if major > 1 || minor >= 37 {
		return trustBundleV1
	}
	return trustBundleV1beta1
}

// The two ClusterTrustBundle API versions a cluster serving the
// pod-certificate APIs may offer.
const (
	trustBundleV1      = "certificates.k8s.io/v1"
	trustBundleV1beta1 = "certificates.k8s.io/v1beta1"
)

// caPools are the five CA pools upstream's installer creates
// (cmd/ate-setup/internal/steps/create.go): Ed25519 for the three in-cluster
// identity CAs, and ECDSA P-256 for the actor identity CA, whose root outside
// relying parties trust and P-256 is more broadly accepted by, and the egress
// interception CA, which Envoy and agentgateway mint leaves from.
var caPools = []caPool{
	{secret: "service-dns-ca-pool", namespace: vars.PodCertificatesNamespace, algorithm: "ED25519", commonName: "Agent Substrate service DNS CA",
		trustBundle: "servicedns.podcert.ate.dev:identity:primary-bundle", signer: "servicedns.podcert.ate.dev/identity"},
	{secret: "pod-identity-ca-pool", namespace: vars.PodCertificatesNamespace, algorithm: "ED25519", commonName: "Agent Substrate pod identity CA",
		trustBundle: "podidentity.podcert.ate.dev:identity:primary-bundle", signer: "podidentity.podcert.ate.dev/identity"},
	{secret: "postgres-ca-pool", namespace: vars.PodCertificatesNamespace, algorithm: "ED25519", commonName: "Agent Substrate PostgreSQL client CA",
		trustBundle: "postgres.podcert.ate.dev:identity:primary-bundle", signer: "postgres.podcert.ate.dev/identity"},
	{secret: "actor-id-ca-pool", namespace: vars.Namespace, algorithm: "ECDSA", commonName: "Agent Substrate actor identity CA"},
	{secret: "egress-mitm-ca-pool", namespace: vars.Namespace, algorithm: "ECDSA", commonName: "Agent Substrate egress interception CA"},
}

// The Secrets around the pools that the manifests mount.
const (
	secretActorJWTPool = "actor-id-jwt-pool"
	secretActorCACerts = "actor-id-ca-certs"
	actorCAPool        = "actor-id-ca-pool"
)

// caPoolSecretData is a CA pool Secret's data (type kubernetes.io/tls).
func caPoolSecretData(privateKeyPKCS8PEM, certificatePEM string) map[string]string {
	return map[string]string{
		"pool":    caPoolJSON(privateKeyPKCS8PEM, certificatePEM),
		"tls.crt": certificatePEM,
		"tls.key": privateKeyPKCS8PEM,
	}
}
