package module

var vars = struct {
	HelmChartName          string
	HelmChartRepoUrl       string
	HelmChartVersion       string
	RedisPasswordSecretKey string
	RedisPort              int
	RedisImageRegistry     string
	RedisImageRepository   string
	RedisImageTag          string
}{
	HelmChartName:          "redis",
	HelmChartRepoUrl:       "https://charts.bitnami.com/bitnami",
	HelmChartVersion:       "17.10.1",
	RedisPasswordSecretKey: "password",
	RedisPort:              6379,
	RedisImageRegistry:     "docker.io",
	RedisImageRepository:   "bitnamilegacy/redis",
	RedisImageTag:          "8.2.1-debian-12-r0",
}
