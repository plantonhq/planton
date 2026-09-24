package kubernetessecretv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSecretSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesSecretSpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal valid Opaque secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "my-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"api-key": "supersecret",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Opaque secret with multiple keys", func() {
			spec := &KubernetesSecretSpec{
				Name: "multi-key-secret",
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "production"},
				},
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"username": "admin",
							"password": "s3cret",
							"api-key":  "abc123",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a namespace provided as a resource reference", func() {
			spec := &KubernetesSecretSpec{
				Name: "ref-ns-secret",
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{Name: "team-namespace"},
					},
				},
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Opaque secret with valid base64 binary_data", func() {
			spec := &KubernetesSecretSpec{
				Name: "binary-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						BinaryData: map[string]string{
							"keystore.jks": "AQIDBA==",
							"cert.der":     "iVBORw0KGgo=",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// A base64 payload held in the platform's secret store reaches this field only as a
		// reference token; the literal arm must not refuse the token's characters ($, @, -).
		ginkgo.It("accepts a binary_data value that is a secret or variable reference token", func() {
			spec := &KubernetesSecretSpec{
				Name: "referenced-binary-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						BinaryData: map[string]string{
							"tls.crt": "$secret/@dev/runner-ca-cert",
							"tls.key": "$secret/runner-ca-key",
							"seed":    "$var/pki/seed-blob",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Opaque secret with disjoint data and binary_data keys", func() {
			spec := &KubernetesSecretSpec{
				Name: "mixed-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"api-key": "supersecret",
						},
						BinaryData: map[string]string{
							"keystore.jks": "AQIDBA==",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid TLS secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "tls-cert",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
						TlsKey: "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid DockerConfigJson secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "registry-creds",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "https://index.docker.io/v1/",
						Username:       "myuser",
						Password:       "mypassword",
						Email:          "user@example.com",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a DockerConfigJson secret without email", func() {
			spec := &KubernetesSecretSpec{
				Name: "registry-creds-no-email",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "_json_key",
						Password:       "{\"type\":\"service_account\"}",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid BasicAuth secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "basic-auth-creds",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "admin",
						Password: "password123",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid SSHAuth secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "ssh-key",
				SecretData: &KubernetesSecretSpec_SshAuth{
					SshAuth: &KubernetesSecretSshAuthData{
						SshPrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nb3Blb...\n-----END OPENSSH PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a ServiceAccountToken secret with a service account name reference", func() {
			spec := &KubernetesSecretSpec{
				Name: "app-identity-token",
				SecretData: &KubernetesSecretSpec_ServiceAccountToken{
					ServiceAccountToken: &KubernetesSecretServiceAccountTokenData{
						ServiceAccountName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{Name: "app-identity"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a ServiceAccountToken secret with a literal service account name", func() {
			spec := &KubernetesSecretSpec{
				Name: "app-identity-token",
				SecretData: &KubernetesSecretSpec_ServiceAccountToken{
					ServiceAccountToken: &KubernetesSecretServiceAccountTokenData{
						ServiceAccountName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "app-identity"},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a secret with labels and annotations", func() {
			spec := &KubernetesSecretSpec{
				Name: "labeled-secret",
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "kube-system"},
				},
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
				},
				Annotations: map[string]string{
					"description": "Platform API credentials",
				},
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a secret with immutable flag", func() {
			spec := &KubernetesSecretSpec{
				Name:      "immutable-secret",
				Immutable: true,
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"config": "frozen-value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a name with dots (DNS subdomain)", func() {
			spec := &KubernetesSecretSpec{
				Name: "my.dotted.secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects empty secret name", func() {
			spec := &KubernetesSecretSpec{
				Name: "",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name with uppercase letters", func() {
			spec := &KubernetesSecretSpec{
				Name: "MySecret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name starting with a dot", func() {
			spec := &KubernetesSecretSpec{
				Name: ".hidden-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name ending with a hyphen", func() {
			spec := &KubernetesSecretSpec{
				Name: "my-secret-",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name longer than 253 characters", func() {
			longName := "a"
			for i := 0; i < 253; i++ {
				longName += "a"
			}
			spec := &KubernetesSecretSpec{
				Name: longName,
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty namespace message", func() {
			spec := &KubernetesSecretSpec{
				Name:      "my-secret",
				Namespace: &foreignkeyv1.StringValueOrRef{},
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects spec without any secret data", func() {
			spec := &KubernetesSecretSpec{
				Name: "empty-secret",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Opaque secret with both data and binary_data empty", func() {
			spec := &KubernetesSecretSpec{
				Name: "empty-opaque",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data:       map[string]string{},
						BinaryData: map[string]string{},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Opaque secret with a binary_data value that is not valid base64", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-base64-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						BinaryData: map[string]string{
							"blob.bin": "not-valid-base64!!!",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		// The reference arm is the two prefixes and nothing else: a bare sigil, an empty path,
		// or an unknown prefix is still a malformed literal.
		ginkgo.It("rejects a binary_data value that only resembles a reference token", func() {
			for _, value := range []string{"$secret/", "$var/", "$token/abc", "secret/abc"} {
				spec := &KubernetesSecretSpec{
					Name: "almost-a-reference",
					SecretData: &KubernetesSecretSpec_Opaque{
						Opaque: &KubernetesSecretOpaqueData{
							BinaryData: map[string]string{"blob.bin": value},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil(), "value %q must be refused", value)
			}
		})

		ginkgo.It("rejects Opaque secret with the same key in data and binary_data", func() {
			spec := &KubernetesSecretSpec{
				Name: "overlap-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"shared-key": "text-value",
						},
						BinaryData: map[string]string{
							"shared-key": "AQIDBA==",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects TLS secret with empty certificate", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-tls",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "",
						TlsKey: "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects TLS secret with empty key", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-tls-key",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
						TlsKey: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty registry server", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "",
						Username:       "user",
						Password:       "pass",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty username", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker-user",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "",
						Password:       "pass",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty password", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker-pass",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "user",
						Password:       "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects BasicAuth with empty username", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-basic-auth",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "",
						Password: "password",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects BasicAuth with empty password", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-basic-auth-pass",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "admin",
						Password: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects SSHAuth with empty private key", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-ssh",
				SecretData: &KubernetesSecretSpec_SshAuth{
					SshAuth: &KubernetesSecretSshAuthData{
						SshPrivateKey: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ServiceAccountToken without a service account name", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-sa-token",
				SecretData: &KubernetesSecretSpec_ServiceAccountToken{
					ServiceAccountToken: &KubernetesSecretServiceAccountTokenData{},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ServiceAccountToken with an empty service account name message", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-sa-token-empty-ref",
				SecretData: &KubernetesSecretSpec_ServiceAccountToken{
					ServiceAccountToken: &KubernetesSecretServiceAccountTokenData{
						ServiceAccountName: &foreignkeyv1.StringValueOrRef{},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
