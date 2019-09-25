package config

import "os"

type Config struct {
	UPSContainerName        string
	PostgresContainerName   string
	OauthProxyContainerName string

	UPSImage        string
	PostgresImage   string
	OauthProxyImage string

	BackupImage string

	UPSMemoryLimit   string
	UPSMemoryRequest string
	UPSCpuLimit      string
	UPSCpuRequest    string

	OauthMemoryLimit   string
	OauthMemoryRequest string
	OauthCpuLimit      string
	OauthCpuRequest    string

	PostgresMemoryLimit   string
	PostgresMemoryRequest string
	PostgresCpuLimit      string
	PostgresCpuRequest    string
	PostgresPVCSize       string
}

func New() Config {
	return Config{
		UPSContainerName:        getEnv("UPS_CONTAINER_NAME", "ups"),
		PostgresContainerName:   getEnv("POSTGRES_CONTAINER_NAME", "postgresql"),
		OauthProxyContainerName: getEnv("OAUTH_PROXY_CONTAINER_NAME", "ups-oauth-proxy"),

		UPSImage:        getEnv("UPS_IMAGE", "quay.io/aerogear/unifiedpush-configurable-container:2.3"),
		PostgresImage:   getEnv("POSTGRES_IMAGE", "docker.io/centos/postgresql-96-centos7:9.6"),
		OauthProxyImage: getEnv("OAUTH_PROXY_IMAGE", "quay.io/openshift/origin-oauth-proxy:4.2.0"),

		BackupImage: getEnv("BACKUP_IMAGE", "quay.io/integreatly/backup-container:1.0.8"),

		UPSMemoryLimit:   getEnv("UPS_MEMORY_LIMIT", "2Gi"),
		UPSMemoryRequest: getEnv("UPS_MEMORY_REQUEST", "512Mi"),
		UPSCpuLimit:      getEnv("UPS_CPU_LIMIT", "1"),
		UPSCpuRequest:    getEnv("UPS_CPU_REQUEST", "500m"),

		OauthMemoryLimit:   getEnv("OAUTH_MEMORY_LIMIT", "64Mi"),
		OauthMemoryRequest: getEnv("OAUTH_MEMORY_REQUEST", "32Mi"),
		OauthCpuLimit:      getEnv("OAUTH_CPU_LIMIT", "20m"),
		OauthCpuRequest:    getEnv("OAUTH_CPU_REQUEST", "10m"),

		PostgresMemoryLimit:   getEnv("POSTGRES_MEMORY_LIMIT", "512Mi"),
		PostgresMemoryRequest: getEnv("POSTGRES_MEMORY_REQUEST", "256Mi"),
		PostgresCpuLimit:      getEnv("POSTGRES_CPU_LIMIT", "1"),
		PostgresCpuRequest:    getEnv("POSTGRES_CPU_REQUEST", "250m"),
		PostgresPVCSize:       getEnv("POSTGRES_PVC_SIZE", "5Gi"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
