package module

var vars = struct {
	// ApiVersion / Kind identify the custom resource the Planton operator
	// reconciles. The CRD is installed by KubernetesPlantonOperator
	// (module-owned there); this module only declares instances of it.
	ApiVersion string
	Kind       string

	// GatewayServiceSuffix / SetupCodeSecretSuffix / SetupCodeSecretKey
	// mirror the operator's deterministic per-platform naming — every
	// object the operator creates is "{platform name}-<suffix>". The
	// outputs derive from these so consumers (people, the desktop app's
	// connect-existing flow) get working handles from the first apply.
	GatewayServiceSuffix  string
	SetupCodeSecretSuffix string
	SetupCodeSecretKey    string
	GatewayDefaultPort    int
	GatewayServicePort    int

	// PostgresClusterSuffix is the operator's name for the platform's
	// database ("{platform name}-postgres"). The credential Secrets this
	// module materializes for the database's backup and recovery stores
	// hang off that name, so they read as the database's own beside the
	// ObjectStore and settings Secret the operator creates under it.
	PostgresClusterSuffix           string
	BackupCredentialsSecretSuffix   string
	RecoveryCredentialsSecretSuffix string
	BackupEndpointCaSecretSuffix    string
	RecoveryEndpointCaSecretSuffix  string
	// EndpointCaSecretKey is the one key of an endpoint-CA Secret; the CR
	// names it in endpointCASecretRef so the operator never guesses.
	EndpointCaSecretKey string

	// DeleteTimeout bounds destroy. Platform teardown is Kubernetes
	// garbage collection (every operator-created object is
	// owner-referenced to the CR), so the CR's own deletion normally
	// returns quickly — the budget exists for API-server pressure and
	// any future operator finalizer, never as an expected wait.
	DeleteTimeout string
}{
	ApiVersion:            "planton.ai/v1",
	Kind:                  "PlantonPlatform",
	GatewayServiceSuffix:  "-gateway",
	SetupCodeSecretSuffix: "-identity-setup-code",
	SetupCodeSecretKey:    "setup-code",
	GatewayDefaultPort:    8080,
	GatewayServicePort:    80,

	PostgresClusterSuffix:           "-postgres",
	BackupCredentialsSecretSuffix:   "-backup-creds",
	RecoveryCredentialsSecretSuffix: "-recovery-creds",
	BackupEndpointCaSecretSuffix:    "-backup-endpoint-ca",
	RecoveryEndpointCaSecretSuffix:  "-recovery-endpoint-ca",
	EndpointCaSecretKey:             "ca.crt",

	DeleteTimeout: "15m",
}
