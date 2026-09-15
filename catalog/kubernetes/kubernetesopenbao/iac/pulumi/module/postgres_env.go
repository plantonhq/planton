package module

import (
	"strconv"

	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
)

// The standard PostgreSQL connection environment. OpenBao's PostgreSQL
// backend runs on pgx, and pgx reads every connection component from the
// libpq environment when the configuration's `connection_url` is blank —
// so the connection is handed to the server as these variables and no
// connection string is ever composed. The non-secret facts ride the
// chart's plain extraEnvironmentVars (postgresPlainEnv); the password
// rides extraSecretEnvironmentVars from the referenced Secret and key
// (values.go), the same seam the seal credentials use. Nothing here
// reaches the configuration ConfigMap.
const (
	envPgHost     = "PGHOST"
	envPgPort     = "PGPORT"
	envPgDatabase = "PGDATABASE"
	envPgUser     = "PGUSER"
	envPgSslMode  = "PGSSLMODE"
	envPgPassword = "PGPASSWORD"
)

// Spec defaults the module applies when a PostgreSQL field is left empty
// (the same values the proto declares as (default); duplicated here
// because the modules read the wire value, and the Terraform twin
// coalesces to the same literals).
const (
	defaultPgPort     = 5432
	defaultPgUsername = "app"
	defaultPgSslMode  = "require"
)

// postgresEnv resolves the PostgreSQL arm into the driver's plain
// environment and the referenced password Secret and key. Every return is
// empty when the engine is not PostgreSQL, so callers can wire the results
// unconditionally.
func postgresEnv(pg *kubernetesopenbaov1alpha1.KubernetesOpenBaoPostgresqlStorage) (plainEnv map[string]string, secretName, secretKey string) {
	if pg == nil {
		return nil, "", ""
	}
	port := defaultPgPort
	if pg.Port != nil && pg.GetPort() != 0 {
		port = int(pg.GetPort())
	}
	username := defaultPgUsername
	if pg.Username != nil && pg.GetUsername() != "" {
		username = pg.GetUsername()
	}
	sslMode := defaultPgSslMode
	if pg.SslMode != nil && pg.GetSslMode() != "" {
		sslMode = pg.GetSslMode()
	}
	secretKey = "password"
	if pg.GetPasswordSecret().SecretKey != nil && pg.GetPasswordSecret().GetSecretKey() != "" {
		secretKey = pg.GetPasswordSecret().GetSecretKey()
	}
	return map[string]string{
		envPgHost:     pg.GetHost().GetValue(),
		envPgPort:     strconv.Itoa(port),
		envPgDatabase: pg.GetDatabase(),
		envPgUser:     username,
		envPgSslMode:  sslMode,
	}, pg.GetPasswordSecret().GetSecretName().GetValue(), secretKey
}
