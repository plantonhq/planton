package module

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestBothEnginesPoolsParseWithUpstreamsCode proves the generated key pools
// are the pools Agent Substrate reads: each engine's CA pools and actor token
// pool (ES256 and RS256) are parsed by upstream's own localca.Unmarshal and
// localjwtauthority.Unmarshal at the module's release, and checked the way
// upstream's installer creates them: one CA, active, a CA certificate whose
// key is the pool's key, of the algorithm upstream uses for that pool. Each
// engine's published key set must be, byte for byte, what upstream's
// localjwtauthority.PublicJWKS derives from its pool.
//
// The Pulumi side builds its pools here with the module's own builders over
// keys of each algorithm (the TLS provider emits the same PEM forms). The
// Terraform side applies pools.tf alone, with no cluster, and reads its
// pools back. The parser runs from a throwaway module whose path sits under
// upstream's, which Go's internal-package rule requires to import
// upstream's internal packages; it fetches upstream through the module
// proxy, and Terraform fetches the TLS provider, so the test runs when
// asked: AGENT_SUBSTRATE_UPSTREAM_PARSE=1, with `tofu` on PATH.
func TestBothEnginesPoolsParseWithUpstreamsCode(t *testing.T) {
	if os.Getenv("AGENT_SUBSTRATE_UPSTREAM_PARSE") != "1" {
		t.Skip("set AGENT_SUBSTRATE_UPSTREAM_PARSE=1 (with tofu on PATH and module-proxy access) to parse both engines' pools with upstream's code")
	}
	type poolCase struct {
		Name      string `json:"name"`
		Kind      string `json:"kind"`      // ca | jwt
		Algorithm string `json:"algorithm"` // ED25519 | ECDSA | ES256 | RS256
		Pool      string `json:"pool"`
		JWKS      string `json:"jwks"` // jwt: the key set the engine publishes
	}
	var cases []poolCase

	// Pulumi side: the module's builders over each key form the provider
	// emits (Ed25519 PKCS #8; ECDSA SEC 1 and RSA PKCS #1, re-encoded).
	ed, edCert := testCA(t, "ED25519")
	cases = append(cases, poolCase{"pulumi/ed25519-ca", "ca", "ED25519", caPoolJSON(ed, edCert), ""})
	ec, ecCert := testCA(t, "ECDSA")
	cases = append(cases, poolCase{"pulumi/ecdsa-ca", "ca", "ECDSA", caPoolJSON(ec, ecCert), ""})
	for _, alg := range []string{"ES256", "RS256"} {
		key, public := testJWTKey(t, alg)
		jwks, err := jwksJSON(alg, public)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, poolCase{"pulumi/" + alg, "jwt", alg, jwtPoolJSON(alg, key, public), jwks})
	}

	// Terraform side: pools.tf applied alone, once per signing algorithm.
	for _, alg := range []string{"es256", "rs256"} {
		dir := t.TempDir()
		for _, file := range []string{"pools.tf", "variables.tf"} {
			raw, err := os.ReadFile(filepath.Join("..", "..", "tf", file))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, file), raw, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		harness := `terraform {
  required_providers {
    tls = { source = "hashicorp/tls", version = "~> 4.0" }
  }
}
output "ca_pools" {
  value     = { for name, pool in local.ca_pools : name => { algorithm = pool.algorithm, pool = local.ca_pool_data[name].pool } }
  sensitive = true
}
output "jwt_pool" {
  value     = local.actor_jwt_pool
  sensitive = true
}
output "jwks" {
  value = local.actor_jwt_jwks
}
`
		tfvars := `metadata = { name = "pools" }
spec = {
  namespace = "ate-system"
  platform  = { generic = { service_account_issuer = "https://kubernetes.default.svc.cluster.local" } }
  database  = { bundled = {} }
  storage   = { gcs = { bucket = "b" } }
  egress    = { credential_provider = { disabled = {} } }
  api_server = { actor_jwt = { algorithm = "` + alg + `" } }
}
`
		if err := os.WriteFile(filepath.Join(dir, "harness.tf"), []byte(harness), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "terraform.tfvars"), []byte(tfvars), 0o644); err != nil {
			t.Fatal(err)
		}
		cache := t.TempDir()
		for _, args := range [][]string{{"init", "-input=false", "-backend=false"}, {"apply", "-input=false", "-auto-approve"}} {
			if _, err := tofu(dir, cache, args...); err != nil {
				t.Fatal(err)
			}
		}
		raw, err := tofu(dir, cache, "output", "-json")
		if err != nil {
			t.Fatal(err)
		}
		var outputs struct {
			CAPools struct {
				Value map[string]struct {
					Algorithm string `json:"algorithm"`
					Pool      string `json:"pool"`
				} `json:"value"`
			} `json:"ca_pools"`
			JWTPool struct {
				Value string `json:"value"`
			} `json:"jwt_pool"`
			JWKS struct {
				Value string `json:"value"`
			} `json:"jwks"`
		}
		if err := json.Unmarshal(raw, &outputs); err != nil {
			t.Fatal(err)
		}
		if alg == "es256" {
			for name, pool := range outputs.CAPools.Value {
				cases = append(cases, poolCase{"terraform/" + name, "ca", pool.Algorithm, pool.Pool, ""})
			}
		}
		cases = append(cases, poolCase{"terraform/" + alg, "jwt", map[string]string{"es256": "ES256", "rs256": "RS256"}[alg], outputs.JWTPool.Value, outputs.JWKS.Value})
	}

	// Upstream's parsers, from a module under upstream's path. Upstream needs
	// its own Go release, which the toolchain fetches.
	checker := t.TempDir()
	casesJSON, _ := json.Marshal(cases)
	files := map[string]string{
		"go.mod":     "module " + upstreamModule + "/plantonpoolcheck\n\ngo 1.27\n\nrequire " + upstreamModule + " " + vars.Release + "\n",
		"cases.json": string(casesJSON),
		"main.go":    poolChecker,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(checker, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = checker
	tidy.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod", "GOTOOLCHAIN=auto")
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("resolving upstream %s: %v\n%s", vars.Release, err, out)
	}
	run := exec.Command("go", "run", ".")
	run.Dir = checker
	run.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=auto")
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("upstream's parsers refused a pool: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

// upstreamModule is Agent Substrate's Go module path.
const upstreamModule = "go.ate.dev/substrate"

func testCA(t *testing.T, algorithm string) (string, string) {
	t.Helper()
	var key interface{}
	var public interface{}
	var keyPEM string
	switch algorithm {
	case "ED25519":
		_, ed, _ := ed25519.GenerateKey(rand.Reader)
		key, public = ed, ed.Public()
		der, _ := x509.MarshalPKCS8PrivateKey(ed)
		keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	case "ECDSA":
		ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		key, public = ec, ec.Public()
		sec1, _ := x509.MarshalECPrivateKey(ec)
		converted, err := toPKCS8PEM(string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: sec1})))
		if err != nil {
			t.Fatal(err)
		}
		keyPEM = converted
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, template, template, public, key)
	if err != nil {
		t.Fatal(err)
	}
	return keyPEM, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func testJWTKey(t *testing.T, algorithm string) (string, string) {
	t.Helper()
	var keyPEM string
	var public interface{}
	if algorithm == "RS256" {
		rk, _ := rsa.GenerateKey(rand.Reader, 4096)
		converted, err := toPKCS8PEM(string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rk)})))
		if err != nil {
			t.Fatal(err)
		}
		keyPEM, public = converted, rk.Public()
	} else {
		ek, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		sec1, _ := x509.MarshalECPrivateKey(ek)
		converted, err := toPKCS8PEM(string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: sec1})))
		if err != nil {
			t.Fatal(err)
		}
		keyPEM, public = converted, ek.Public()
	}
	spki, _ := x509.MarshalPKIXPublicKey(public)
	return keyPEM, string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: spki}))
}

// poolChecker parses every case with upstream's code and checks it as
// upstream's installer creates it.
const poolChecker = `package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"

	"go.ate.dev/substrate/internal/localca"
	"go.ate.dev/substrate/internal/localjwtauthority"
)

type poolCase struct {
	Name, Kind, Algorithm, Pool, JWKS string
}

func publicOf(key crypto.PrivateKey) []byte {
	der, _ := x509.MarshalPKIXPublicKey(key.(crypto.Signer).Public())
	return der
}

func main() {
	raw, _ := os.ReadFile("cases.json")
	var cases []poolCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		panic(err)
	}
	failed := false
	fail := func(c poolCase, format string, args ...interface{}) {
		failed = true
		fmt.Printf("FAIL %s: %s\n", c.Name, fmt.Sprintf(format, args...))
	}
	for _, c := range cases {
		switch c.Kind {
		case "ca":
			pool, err := localca.Unmarshal([]byte(c.Pool))
			if err != nil {
				fail(c, "localca.Unmarshal: %v", err)
				continue
			}
			if len(pool.CAs) != 1 || pool.ActiveForSigning != pool.CAs[0].ID {
				fail(c, "want one CA, active for signing")
				continue
			}
			ca := pool.CAs[0]
			if !ca.RootCertificate.IsCA {
				fail(c, "the root certificate is not a CA")
			}
			if ca.RootCertificate.KeyUsage&x509.KeyUsageCertSign == 0 {
				fail(c, "the root certificate cannot sign certificates")
			}
			if !bytes.Equal(publicOf(ca.SigningKey), ca.RootCertificate.RawSubjectPublicKeyInfo) {
				fail(c, "the signing key is not the certificate's key")
			}
			switch k := ca.SigningKey.(type) {
			case ed25519.PrivateKey:
				if c.Algorithm != "ED25519" {
					fail(c, "an Ed25519 key where %s was expected", c.Algorithm)
				}
			case *ecdsa.PrivateKey:
				if c.Algorithm != "ECDSA" || k.Curve.Params().Name != "P-256" {
					fail(c, "an ECDSA %s key where %s was expected", k.Curve.Params().Name, c.Algorithm)
				}
			default:
				fail(c, "unexpected key type %T", k)
			}
		case "jwt":
			pool, err := localjwtauthority.Unmarshal([]byte(c.Pool))
			if err != nil {
				fail(c, "localjwtauthority.Unmarshal: %v", err)
				continue
			}
			if len(pool.Authorities) != 1 || pool.ActiveForSigning != pool.Authorities[0].ID {
				fail(c, "want one signing key, active")
				continue
			}
			a := pool.Authorities[0]
			if a.Algorithm != c.Algorithm {
				fail(c, "algorithm %s, want %s", a.Algorithm, c.Algorithm)
			}
			switch k := a.SigningKey.(type) {
			case *ecdsa.PrivateKey:
				if c.Algorithm != "ES256" || k.Curve.Params().Name != "P-256" {
					fail(c, "an ECDSA key for %s", c.Algorithm)
				}
			case *rsa.PrivateKey:
				if c.Algorithm != "RS256" || k.N.BitLen() != 4096 {
					fail(c, "an RSA %d key for %s", k.N.BitLen(), c.Algorithm)
				}
			default:
				fail(c, "unexpected key type %T", k)
			}
			want, err := localjwtauthority.PublicJWKS([]byte(c.Pool))
			if err != nil {
				fail(c, "localjwtauthority.PublicJWKS: %v", err)
			} else if c.JWKS != string(want) {
				fail(c, "the published key set differs from upstream's:\n got %s\nwant %s", c.JWKS, want)
			}
		}
		if !failed {
			fmt.Printf("ok   %s\n", c.Name)
		}
	}
	if failed {
		os.Exit(1)
	}
}
`
