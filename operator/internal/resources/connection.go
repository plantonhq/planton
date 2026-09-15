package resources

import (
	"fmt"
)

// PostgreSQLConnectionInfo provides host, port, and credential references for
// connecting to PostgreSQL. Every consumer (control plane, identity server,
// OpenFGA, Temporal) reads only this struct -- it is the one seam between the
// database's deployment mechanics and everything that speaks to it.
type PostgreSQLConnectionInfo struct {
	Host       string
	Port       int32
	User       string
	SecretName string
	UserKey    string
	PassKey    string
}

// PostgreSQLConnection returns connection info for the platform's PostgreSQL
// cluster (deployed via CloudNativePG). One cluster serves every consumer
// through the primary read-write Service, as the superuser, with credentials
// from the CloudNativePG-generated superuser Secret -- see
// PostgreSQLSuperuser for why the single-user contract is deliberate.
func PostgreSQLConnection(crName, namespace string) PostgreSQLConnectionInfo {
	return PostgreSQLConnectionInfo{
		Host:       PostgreSQLHost(crName, namespace),
		Port:       PostgreSQLPort,
		User:       PostgreSQLSuperuser,
		SecretName: PostgreSQLSuperuserSecretName(crName),
		UserKey:    "username",
		PassKey:    "password",
	}
}

// OpenBAOConnectionInfo provides API address and credential references.
type OpenBAOConnectionInfo struct {
	APIAddr        string
	Port           int32
	InitSecretName string
	RootTokenKey   string
}

// OpenBAOConnection returns connection info for the OpenBAO instance.
// initSecretName is the Secret the vault's root token lives in -- the
// adopter's own when the declaration names one, the operator's otherwise --
// so the consumer reads the token from wherever the operator wrote it.
func OpenBAOConnection(crName, namespace, initSecretName string) OpenBAOConnectionInfo {
	return OpenBAOConnectionInfo{
		APIAddr:        OpenBAOAPIAddr(crName, namespace),
		Port:           OpenBAOPort,
		InitSecretName: initSecretName,
		RootTokenKey:   OpenBAOInitSecretRootTokenKey,
	}
}

// Neo4jConnectionInfo provides host, port, and credential references for Neo4j.
type Neo4jConnectionInfo struct {
	BoltURI        string
	Host           string
	BoltPort       int32
	HTTPPort       int32
	AuthSecretName string
	PasswordKey    string
	Username       string
}

// Neo4jConnection returns connection info for the Neo4j instance.
func Neo4jConnection(crName, namespace string) Neo4jConnectionInfo {
	return Neo4jConnectionInfo{
		BoltURI:        Neo4jBoltURI(crName, namespace),
		Host:           Neo4jServiceHost(crName, namespace),
		BoltPort:       Neo4jBoltPort,
		HTTPPort:       Neo4jHTTPPort,
		AuthSecretName: Neo4jAuthSecretName(crName),
		PasswordKey:    Neo4jPasswordSecretKey,
		Username:       Neo4jDefaultUser,
	}
}

// ---------------------------------------------------------------------------
// Redis
// ---------------------------------------------------------------------------

// RedisConnectionInfo provides host, port, and credential references for the
// redis-protocol cache (served by Valkey -- see valkey_helm.go).
type RedisConnectionInfo struct {
	Host       string
	Port       int32
	SecretName string
	PassKey    string
}

// RedisConnection returns connection info for the central redis-protocol cache.
func RedisConnection(crName, namespace string) RedisConnectionInfo {
	return RedisConnectionInfo{
		Host:       RedisServiceHost(crName, namespace),
		Port:       int32(RedisPort),
		SecretName: RedisSecretName(crName),
		PassKey:    RedisSecretKey,
	}
}

// ---------------------------------------------------------------------------
// OpenFGA
// ---------------------------------------------------------------------------

// OpenFGAConnectionInfo provides the HTTP endpoint and the name of the
// bootstrap ConfigMap that holds the store_id (the model inside the store is
// the control plane's own, established at its boot).
type OpenFGAConnectionInfo struct {
	HTTPURL                string
	BootstrapConfigMapName string
}

// OpenFGAConnection returns connection info for the OpenFGA instance.
func OpenFGAConnection(crName, namespace string) OpenFGAConnectionInfo {
	return OpenFGAConnectionInfo{
		HTTPURL:                OpenFGAHTTPURL(crName, namespace),
		BootstrapConfigMapName: fmt.Sprintf("%s-fga-bootstrap", crName),
	}
}

// ---------------------------------------------------------------------------
// Temporal
// ---------------------------------------------------------------------------

// TemporalConnectionInfo provides the frontend gRPC endpoint and namespace.
type TemporalConnectionInfo struct {
	FrontendEndpoint string
	Namespace        string
}

// TemporalConnection returns connection info for the Temporal frontend.
func TemporalConnection(crName, namespace string) TemporalConnectionInfo {
	return TemporalConnectionInfo{
		FrontendEndpoint: TemporalFrontendEndpoint(crName, namespace),
		Namespace:        "default",
	}
}
