package module

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
	tls "github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
)

// applySecrets registers every Secret the install reads, as typed Secret
// resources rather than manifest documents, so key material never sits in
// the documents the engines compare: the five CA pools and the actor token
// signing key, generated here and kept in the deployment's state (a
// re-apply never replaces them; a destroy does, exactly like an upstream
// reinstall), the API server's environment Secret, and an external
// database's server CA. Terraform twin: secrets.tf.
func applySecrets(ctx *pulumi.Context, s *settings, labels map[string]string, opts ...pulumi.ResourceOption) ([]pulumi.Resource, error) {
	var created []pulumi.Resource
	secret := func(name, namespace, secretType string, data pulumi.StringMapInput) error {
		r, err := kubernetescorev1.NewSecret(ctx, namespace+"-"+name, &kubernetescorev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(name),
				Namespace: pulumi.String(namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Type:       pulumi.String(secretType),
			StringData: data,
		}, opts...)
		if err != nil {
			return errors.Wrapf(err, "failed to register Secret %s/%s", namespace, name)
		}
		created = append(created, r)
		return nil
	}

	for _, pool := range caPools {
		keyArgs := &tls.PrivateKeyArgs{Algorithm: pulumi.String(pool.algorithm)}
		if pool.algorithm == "ECDSA" {
			keyArgs.EcdsaCurve = pulumi.String("P256")
		}
		key, err := tls.NewPrivateKey(ctx, pool.secret, keyArgs)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate the %s key", pool.secret)
		}
		cert, err := tls.NewSelfSignedCert(ctx, pool.secret, &tls.SelfSignedCertArgs{
			PrivateKeyPem:       key.PrivateKeyPem,
			Subject:             &tls.SelfSignedCertSubjectArgs{CommonName: pulumi.String(pool.commonName)},
			ValidityPeriodHours: pulumi.Int(caValidityHours),
			IsCaCertificate:     pulumi.Bool(true),
			AllowedUses:         pulumi.ToStringArray([]string{"digital_signature", "cert_signing"}),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to issue the %s certificate", pool.secret)
		}
		data := pulumi.All(key.PrivateKeyPem, cert.CertPem).ApplyT(func(args []interface{}) (map[string]string, error) {
			pkcs8, err := toPKCS8PEM(args[0].(string))
			if err != nil {
				return nil, err
			}
			return caPoolSecretData(pkcs8, args[1].(string)), nil
		}).(pulumi.StringMapOutput)
		if err := secret(pool.secret, pool.namespace, "kubernetes.io/tls", data); err != nil {
			return nil, err
		}
		if pool.secret == actorCAPool {
			if err := secret(secretActorCACerts, vars.Namespace, "Opaque", pulumi.StringMap{"ca.crt": cert.CertPem}); err != nil {
				return nil, err
			}
		}
	}

	jwtArgs := &tls.PrivateKeyArgs{Algorithm: pulumi.String("ECDSA"), EcdsaCurve: pulumi.String("P256")}
	if s.apiServer.jwtAlgorithm == "RS256" {
		jwtArgs = &tls.PrivateKeyArgs{Algorithm: pulumi.String("RSA"), RsaBits: pulumi.Int(4096)}
	}
	// A key's algorithm cannot change in place, so changing it replaces the
	// key, as the field doc says.
	jwtKey, err := tls.NewPrivateKey(ctx, secretActorJWTPool, jwtArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate the actor token signing key")
	}
	jwtData := pulumi.All(jwtKey.PrivateKeyPem, jwtKey.PublicKeyPem).ApplyT(func(args []interface{}) (map[string]string, error) {
		pkcs8, err := toPKCS8PEM(args[0].(string))
		if err != nil {
			return nil, err
		}
		jwks, err := jwksJSON(s.apiServer.jwtAlgorithm, args[1].(string))
		if err != nil {
			return nil, err
		}
		return map[string]string{
			"pool": jwtPoolJSON(s.apiServer.jwtAlgorithm, pkcs8, args[1].(string)),
			"jwks": jwks,
		}, nil
	}).(pulumi.StringMapOutput)
	if err := secret(secretActorJWTPool, vars.Namespace, "Opaque", jwtData); err != nil {
		return nil, err
	}

	if err := secret(secretAPIEnv, vars.Namespace, "Opaque", pulumi.ToStringMap(apiEnvSecret(s))); err != nil {
		return nil, err
	}
	if s.database.serverCA != "" {
		if err := secret(secretPostgresServerCA, vars.Namespace, "Opaque", pulumi.StringMap{"server-ca.pem": pulumi.String(s.database.serverCA)}); err != nil {
			return nil, err
		}
	}
	return created, nil
}

// toPKCS8PEM re-encodes a private key PEM as PKCS #8 ("PRIVATE KEY"), the
// form upstream's pools and tls.key hold. The TLS provider's Ed25519 keys
// are PKCS #8 already; its ECDSA and RSA keys come as SEC 1 and PKCS #1.
// x509.MarshalPKCS8PrivateKey is what the provider's own
// private_key_pem_pkcs8 (the Terraform twin's source) uses, so both engines
// write the same encoding.
func toPKCS8PEM(privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errors.New("the generated private key is not PEM")
	}
	var key interface{}
	var err error
	switch block.Type {
	case "PRIVATE KEY":
		return privateKeyPEM, nil
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return "", errors.Errorf("the generated private key has PEM type %q", block.Type)
	}
	if err != nil {
		return "", errors.Wrap(err, "parsing the generated private key")
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", errors.Wrap(err, "encoding the generated private key as PKCS #8")
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), nil
}
