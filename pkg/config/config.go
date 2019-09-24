package config

import "os"

type Config struct {
	UPSContainerName        string
	PostgresContainerName   string
	OauthProxyContainerName string

	UPSImageStreamName string
	UPSImageStreamTag  string

	OauthProxyImageStreamName string
	OauthProxyImageStreamTag  string

	PostgresImageStreamNamespace string
	PostgresImageStreamName      string
	PostgresImageStreamTag       string

	UPSImageStreamInitialImage        string
	PostgresImageStreamInitialImage   string
	OauthProxyImageStreamInitialImage string

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

		UPSImageStreamName: getEnv("UPS_IMAGE_STREAM_NAME", "ups-imagestream"),
		UPSImageStreamTag:  getEnv("UPS_IMAGE_STREAM_TAG", "2.3"),

		OauthProxyImageStreamName: getEnv("OAUTH_PROXY_IMAGE_STREAM_NAME", "ups-oauth-proxy-imagestream"),
		OauthProxyImageStreamTag:  getEnv("OAUTH_PROXY_IMAGE_STREAM_TAG", "latest"),

		PostgresImageStreamNamespace: getEnv("POSTGRES_IMAGE_STREAM_NAMESPACE", "openshift"),
		PostgresImageStreamName:      getEnv("POSTGRES_IMAGE_STREAM_NAME", "postgresql"),
		// Used both to set the tag, and also for the "POSTGRES_VERSION" in the Secret
		PostgresImageStreamTag: getEnv("POSTGRES_IMAGE_STREAM_TAG", "10"),

		// these are used when the image stream does not exist and created for the first time by the operator
		UPSImageStreamInitialImage:        getEnv("UPS_IMAGE_STREAM_INITIAL_IMAGE", "quay.io/aerogear/unifiedpush-configurable-container:2.3"),
		OauthProxyImageStreamInitialImage: getEnv("OAUTH_PROXY_IMAGE_STREAM_INITIAL_IMAGE", "quay.io/openshift/origin-oauth-proxy:4.2.0"),

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
