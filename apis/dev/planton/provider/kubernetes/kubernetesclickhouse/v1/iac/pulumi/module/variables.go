package module

var vars = struct {
	ClickhousePasswordKey string
	DefaultUsername       string
	ClickhouseHttpPort    int
	ClickhouseNativePort  int
	ClickhouseVersion     string
	ZookeeperPort         int
	OperatorNamespace     string
}{
	ClickhousePasswordKey: "admin-password",
	DefaultUsername:       "default",
	ClickhouseHttpPort:    8123,
	ClickhouseNativePort:  9000,
	// Default ClickHouse version - can be overridden in spec
	// Using a recent stable version compatible with Altinity operator 0.25.4
	ClickhouseVersion: "24.8",
	ZookeeperPort:     2181,
	// Namespace where Altinity operator is expected to be installed
	OperatorNamespace: "clickhouse-operator",
}
