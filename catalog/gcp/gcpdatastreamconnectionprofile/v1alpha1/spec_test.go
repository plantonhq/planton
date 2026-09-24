package gcpdatastreamconnectionprofilev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDatastreamConnectionProfileSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const secretVersion = "projects/p/secrets/orders-replication/versions/1"

var _ = ginkgo.Describe("GcpDatastreamConnectionProfileSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	base := func() *GcpDatastreamConnectionProfile {
		return &GcpDatastreamConnectionProfile{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDatastreamConnectionProfile",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders"},
			Spec:       &GcpDatastreamConnectionProfileSpec{Location: "us-central1"},
		}
	}
	postgres := func() *GcpDatastreamConnectionProfilePostgresqlProfile {
		return &GcpDatastreamConnectionProfilePostgresqlProfile{
			Hostname: litRef("203.0.113.10"),
			Username: "datastream",
			Password: "s3cret",
			Database: "orders",
		}
	}
	mongo := func() *GcpDatastreamConnectionProfileMongodbProfile {
		return &GcpDatastreamConnectionProfileMongodbProfile{
			HostAddresses: []*GcpDatastreamConnectionProfileMongodbHostAddress{{Hostname: "cluster0.example.mongodb.net"}},
			Username:      "datastream",
			SecretManagerStoredPassword: litRef(secretVersion),
			SrvConnectionFormat:         true,
		}
	}

	ginkgo.It("should accept each profile type on its own", func() {
		arms := []func(*GcpDatastreamConnectionProfileSpec){
			func(s *GcpDatastreamConnectionProfileSpec) { s.BigqueryProfile = true },
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.GcsProfile = &GcpDatastreamConnectionProfileGcsProfile{Bucket: litRef("raw-cdc"), RootPath: "/datastream"}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.MysqlProfile = &GcpDatastreamConnectionProfileMysqlProfile{
					Hostname: litRef("203.0.113.11"), Port: 3306, Username: "datastream",
					SecretManagerStoredPassword: litRef(secretVersion),
					SslConfig: &GcpDatastreamConnectionProfileMysqlSslConfig{
						CaCertificate: "ca", ClientCertificate: "cert", ClientKey: "key",
					},
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				p := postgres()
				p.SslConfig = &GcpDatastreamConnectionProfilePostgresqlSslConfig{
					ServerVerification: &GcpDatastreamConnectionProfilePostgresqlServerVerification{CaCertificate: "ca"},
				}
				s.PostgresqlProfile = p
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.OracleProfile = &GcpDatastreamConnectionProfileOracleProfile{
					Hostname: "oracle.internal", Username: "datastream", Password: "s3cret",
					DatabaseService: "ORCL", ConnectionAttributes: map[string]string{"TRANSPORT_CONNECT_TIMEOUT": "10"},
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.SqlServerProfile = &GcpDatastreamConnectionProfileSqlServerProfile{
					Hostname: litRef("203.0.113.12"), Username: "datastream", Password: "s3cret", Database: "orders",
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) { s.MongodbProfile = mongo() },
		}
		for i, arm := range arms {
			r := base()
			arm(r.Spec)
			gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "arm %d", i)
		}
	})

	ginkgo.It("should reject no profile type and two profile types", func() {
		r := base()
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r.Spec.BigqueryProfile = true
		r.Spec.PostgresqlProfile = postgres()
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept private connectivity or an SSH tunnel, and reject both", func() {
		r := base()
		r.Spec.PostgresqlProfile = postgres()
		r.Spec.PrivateConnection = litRef("projects/p/locations/us-central1/privateConnections/data-vpc")
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())

		r.Spec.ForwardSshConnectivity = &GcpDatastreamConnectionProfileForwardSshConnectivity{
			Hostname: "bastion.example.com", Username: "tunnel", PrivateKey: "pem",
		}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())

		r.Spec.PrivateConnection = nil
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an SSH tunnel with both a password and a key", func() {
		r := base()
		r.Spec.PostgresqlProfile = postgres()
		r.Spec.ForwardSshConnectivity = &GcpDatastreamConnectionProfileForwardSshConnectivity{
			Hostname: "bastion.example.com", Username: "tunnel", Password: "pw", PrivateKey: "pem",
		}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a password beside a Secret Manager password on every source", func() {
		setters := []func(*GcpDatastreamConnectionProfileSpec){
			func(s *GcpDatastreamConnectionProfileSpec) {
				p := postgres()
				p.SecretManagerStoredPassword = litRef(secretVersion)
				s.PostgresqlProfile = p
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.MysqlProfile = &GcpDatastreamConnectionProfileMysqlProfile{
					Hostname: litRef("h"), Username: "u", Password: "pw", SecretManagerStoredPassword: litRef(secretVersion),
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.OracleProfile = &GcpDatastreamConnectionProfileOracleProfile{
					Hostname: "h", Username: "u", Password: "pw", SecretManagerStoredPassword: litRef(secretVersion), DatabaseService: "ORCL",
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				s.SqlServerProfile = &GcpDatastreamConnectionProfileSqlServerProfile{
					Hostname: litRef("h"), Username: "u", Password: "pw", SecretManagerStoredPassword: litRef(secretVersion), Database: "d",
				}
			},
			func(s *GcpDatastreamConnectionProfileSpec) {
				m := mongo()
				m.Password = "pw"
				s.MongodbProfile = m
			},
		}
		for i, set := range setters {
			r := base()
			set(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "source %d", i)
		}
	})

	ginkgo.It("should enforce the MongoDB connection-format rules", func() {
		r := base()
		m := mongo()
		m.SrvConnectionFormat = false
		r.Spec.MongodbProfile = m
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no format")

		m.StandardConnectionFormat = &GcpDatastreamConnectionProfileMongodbStandardConnectionFormat{DirectConnection: true}
		m.ReplicaSet = "rs0"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "standard with a replica set")

		m.SrvConnectionFormat = true
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both formats")

		m.StandardConnectionFormat = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "SRV with a replica set")

		m.ReplicaSet = ""
		m.HostAddresses = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no hosts")
	})

	ginkgo.It("should enforce Google's MongoDB client-identity rules", func() {
		r := base()
		m := mongo()
		r.Spec.MongodbProfile = m
		m.SslConfig = &GcpDatastreamConnectionProfileMongodbSslConfig{CaCertificate: "ca"}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "server verification only")

		m.SslConfig = &GcpDatastreamConnectionProfileMongodbSslConfig{
			CaCertificate: "ca", ClientCertificate: "cert", SecretManagerStoredClientKey: litRef(secretVersion),
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "key from Secret Manager")

		m.SslConfig.ClientKey = "key"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "two key sources")

		m.SslConfig = &GcpDatastreamConnectionProfileMongodbSslConfig{ClientCertificate: "cert", ClientKey: "key"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "client identity without a CA")

		m.SslConfig = &GcpDatastreamConnectionProfileMongodbSslConfig{CaCertificate: "ca", ClientCertificate: "cert"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "certificate without a key")
	})

	ginkgo.It("should enforce Google's MySQL client-certificate rules", func() {
		r := base()
		p := &GcpDatastreamConnectionProfileMysqlProfile{Hostname: litRef("h"), Username: "u", Password: "pw"}
		r.Spec.MysqlProfile = p
		p.SslConfig = &GcpDatastreamConnectionProfileMysqlSslConfig{}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "TLS without verification")

		p.SslConfig = &GcpDatastreamConnectionProfileMysqlSslConfig{ClientCertificate: "cert", ClientKey: "key"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "client pair without a CA")

		p.SslConfig = &GcpDatastreamConnectionProfileMysqlSslConfig{CaCertificate: "ca", ClientKey: "key"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "key without a certificate")
	})

	ginkgo.It("should reject both PostgreSQL SSL modes", func() {
		r := base()
		p := postgres()
		p.SslConfig = &GcpDatastreamConnectionProfilePostgresqlSslConfig{
			ServerVerification: &GcpDatastreamConnectionProfilePostgresqlServerVerification{CaCertificate: "ca"},
			ServerAndClientVerification: &GcpDatastreamConnectionProfilePostgresqlServerAndClientVerification{
				CaCertificate: "ca", ClientCertificate: "cert", ClientKey: "key",
			},
		}
		r.Spec.PostgresqlProfile = p
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a port outside 1-65535 and missing required source fields", func() {
		r := base()
		p := postgres()
		p.Port = 70000
		r.Spec.PostgresqlProfile = p
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "port")

		p.Port = 0
		p.Database = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "database")

		p.Database = "orders"
		p.Hostname = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "hostname")
	})

	ginkgo.It("should reject a Cloud Storage profile without a bucket", func() {
		r := base()
		r.Spec.GcsProfile = &GcpDatastreamConnectionProfileGcsProfile{RootPath: "/x"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy and a malformed location", func() {
		r := base()
		r.Spec.BigqueryProfile = true
		r.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r.Spec.DeletionPolicy = ""
		r.Spec.Location = "US"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})
})
