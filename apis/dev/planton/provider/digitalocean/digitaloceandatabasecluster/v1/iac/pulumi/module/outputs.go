package module

const (
	// OpClusterId is the UUID of the created database cluster.
	OpClusterId = "cluster_id"
	// OpConnectionUri is the full connection URI (includes user/pass/db).
	OpConnectionUri = "connection_uri"
	// OpHost is the host or IP of the cluster.
	OpHost = "host"
	// OpPort is the port the cluster listens on.
	OpPort = "port"
	// OpDatabaseUser is the default user.
	OpDatabaseUser = "database_user"
	// OpDatabasePassword is that user's password.
	OpDatabasePassword = "database_password"
)
